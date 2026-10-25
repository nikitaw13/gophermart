package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// Config holds the gophermart runtime configuration assembled from built-in
// defaults, environment variables and command-line flags.
type Config struct {
	logLevel            string
	dsn                 string
	jwtSecretKey        string
	accrualAddress      string
	serviceAddress      string
	accrualWorkerCount  int
	accrualPollInterval time.Duration
}

// LoadConfig assembles the configuration from built-in defaults, environment
// variables and command-line flags, validates it, generates an ephemeral JWT
// secret when none is configured, and returns the resulting Config.
func LoadConfig() (Config, error) {
	cfg := Config{
		logLevel:            "INFO",
		dsn:                 "",
		jwtSecretKey:        "",
		accrualAddress:      "http://127.0.0.1:8080",
		serviceAddress:      "127.0.0.1:8888",
		accrualWorkerCount:  3,
		accrualPollInterval: 5 * time.Second,
	}

	envLogLevel, found := os.LookupEnv("LOG_LEVEL")
	if found {
		cfg.logLevel = envLogLevel
	}

	envDSN, found := os.LookupEnv("DATABASE_URI")
	if found {
		cfg.dsn = envDSN
	}

	envJWTSecretKey, found := os.LookupEnv("JWT_SECRET_KEY")
	if found {
		cfg.jwtSecretKey = envJWTSecretKey
	}

	envAccrualAddress, found := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS")
	if found {
		cfg.accrualAddress = envAccrualAddress
	}

	envServiceAddress, found := os.LookupEnv("RUN_ADDRESS")
	if found {
		cfg.serviceAddress = envServiceAddress
	}

	envAccrualWorkerCount, found := os.LookupEnv("ACCRUAL_WORKER_COUNT")
	if found {
		parsedAccrualWorkerCount, err := strconv.Atoi(envAccrualWorkerCount)
		if err != nil {
			return cfg, fmt.Errorf("error converting ACCRUAL_WORKER_COUNT: %w", err)
		}
		cfg.accrualWorkerCount = parsedAccrualWorkerCount
	}

	envAccrualPollInterval, found := os.LookupEnv("ACCRUAL_POLL_INTERVAL")
	if found {
		parsedAccrualPollInterval, err := time.ParseDuration(envAccrualPollInterval)
		if err != nil {
			return cfg, fmt.Errorf("error parsing ACCRUAL_POLL_INTERVAL: %w", err)
		}
		cfg.accrualPollInterval = parsedAccrualPollInterval
	}

	flag.StringVar(&cfg.logLevel, "l", cfg.logLevel, "log level (DEBUG, INFO, WARN, ERROR)")
	flag.StringVar(&cfg.dsn, "d", cfg.dsn, "database connection DSN")
	flag.StringVar(&cfg.jwtSecretKey, "k", cfg.jwtSecretKey, "secret key used to sign JWT tokens")
	flag.StringVar(&cfg.accrualAddress, "r", cfg.accrualAddress, "accrual system URL, e.g. http://127.0.0.1:8080")
	flag.StringVar(&cfg.serviceAddress, "a", cfg.serviceAddress, "loyalty service address and port to listen on")
	flag.IntVar(&cfg.accrualWorkerCount, "c", cfg.accrualWorkerCount, "number of concurrent accrual workers")
	flag.DurationVar(&cfg.accrualPollInterval, "i", cfg.accrualPollInterval, "interval between accrual system polls, e.g. 1s")
	flag.Parse()

	err := cfg.Validate()
	if err != nil {
		return cfg, fmt.Errorf("validate config values: %w", err)
	}

	if cfg.jwtSecretKey == "" {
		buf := make([]byte, 32)
		_, err := rand.Read(buf)
		if err != nil {
			return cfg, fmt.Errorf("generate JWT secret: %w", err)
		}
		cfg.jwtSecretKey = base64.StdEncoding.EncodeToString(buf)
		log.Print("JWT secret is not provided, using ephemeral; sessions won't survive restart")
	}

	return cfg, nil
}

// Validate checks that all required settings are present and well-formed.
func (c Config) Validate() error {
	if c.logLevel == "" {
		return errors.New("log level is not provided")
	}

	_, err := zap.ParseAtomicLevel(c.logLevel)
	if err != nil {
		return fmt.Errorf("invalid log level %q: %w", c.logLevel, err)
	}
	if c.dsn == "" {
		return errors.New("database DSN is not provided")
	}
	parsedURL, err := url.Parse(c.accrualAddress)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return errors.New("accrual system address must include scheme and host, e.g. http://127.0.0.1:8080")
	}
	if c.serviceAddress == "" {
		return errors.New("loyalty service address and port is not provided")
	}
	if c.accrualWorkerCount < 1 {
		return errors.New("accrual worker count must be at least 1")
	}
	if c.accrualPollInterval < time.Second {
		return errors.New("accrual poll interval must be at least 1s")
	}

	return nil
}
