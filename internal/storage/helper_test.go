package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/docker/docker/client"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nikitaw13/gophermart/internal/model"
	"github.com/nikitaw13/gophermart/migrations"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	testUsername                     = "Gopher"
	testAnotherUsername              = "AnotherGopher"
	testYetAnotherUsername           = "YetAnotherGopher"
	testPassword                     = "GopherPassword"
	testPasswordHash                 = "$2a$10$uynMkJPlvZxu1I9stjjpCe50GNpbqPbDH7MzKmb.XI29Vt4VCWIWm"
	testOrderNumber                  = "4368910403255712"
	testAnotherOrderNumber           = "79927398713"
	testAccrualAmount                = 1000.00
	testWithdrawalAmount             = 500.00
	testWithdrawalInsufficientAmount = 1500.00
)

var (
	testFixedTime         = time.Now()
	testPool              *pgxpool.Pool
	testDatabaseTimeouts  = []time.Duration{time.Millisecond, time.Millisecond, time.Millisecond}
	testWantOrderResponse = model.OrderResponse{
		OrderNumber:  testOrderNumber,
		Status:       model.StatusNew,
		Amount:       0,
		UploadedTime: testFixedTime,
	}
	testWithdrawalRequest = model.WithdrawalRequest{
		OrderNumber: testOrderNumber,
		Amount:      testWithdrawalAmount,
	}
	testWithdrawalRequestInsufficientFunds = model.WithdrawalRequest{
		OrderNumber: testAnotherOrderNumber,
		Amount:      testWithdrawalInsufficientAmount,
	}
	testWantListWithdrawalsResponse = model.WithdrawalResponse{
		OrderNumber: testOrderNumber,
		Amount:      testWithdrawalAmount,
	}
)

var (
	prepareOrderFlow = fmt.Sprintf(`
INSERT INTO %s (username, password)
VALUES ($1, $2), ($3, $4)`, usersTableName)

	prepareWithdrawalFlow = fmt.Sprintf(`
	WITH new_user AS (
    INSERT INTO %s (username, password)
    VALUES ($1, $2)
    RETURNING user_id
	)
	INSERT INTO %s (user_id, order_number, status, amount)
	SELECT user_id, $3, 'PROCESSED', $4::numeric(10,2)
	FROM new_user
	RETURNING user_id`, usersTableName, accrualsTableName)

	prepareAccrualFlow = fmt.Sprintf(`
INSERT INTO %s (username, password)
VALUES ($1, $2)`, usersTableName)
)

func TestMain(m *testing.M) {
	if !dockerAvailable() {
		fmt.Println("docker not available — skipping integration tests")
		os.Exit(0)
	}

	ctx := context.Background()
	pg, err := postgres.Run(ctx, "postgres:17-alpine",
		postgres.WithDatabase("gophermart_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithCmd("postgres", "-c", "fsync=off"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start container: %v\n", err)
		os.Exit(1)
	}

	dsn, err := pg.ConnectionString(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dsn: %v\n", err)
		os.Exit(1)
	}

	testPool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pool: %v\n", err)
		os.Exit(1)
	}

	// Apply the embedded migrations to the test database.
	if err := runTestMigrations(ctx, dsn); err != nil {
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	testPool.Close()
	_ = pg.Terminate(ctx)
	os.Exit(code)
}

// dockerAvailable reports whether the Docker daemon is reachable.
func dockerAvailable() bool {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return false
	}
	defer cli.Close()
	_, err = cli.Ping(context.Background())
	return err == nil
}

// resetDB truncates all app tables between tests.
func resetDB(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), `
        DO $$
        DECLARE r record;
        BEGIN
            FOR r IN (
                SELECT tablename FROM pg_tables
                WHERE schemaname = 'public' AND tablename <> 'schema_migrations'
            ) LOOP
                EXECUTE 'TRUNCATE TABLE ' || quote_ident(r.tablename) || ' RESTART IDENTITY CASCADE';
            END LOOP;
        END $$;`)
	if err != nil {
		t.Fatalf("reset db: %v", err)
	}
}

// runTestMigrations applies embedded migrations to the test database.
func runTestMigrations(ctx context.Context, dsn string) error {
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}

	dbDrv, err := pgx.WithInstance(sqlDB, &pgx.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", src, "pgx", dbDrv)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
