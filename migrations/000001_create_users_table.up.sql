CREATE TABLE IF NOT EXISTS users (
    user_id   BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username  VARCHAR(72) NOT NULL UNIQUE,
    password  VARCHAR(72) NOT NULL
);

CREATE TYPE order_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');

CREATE TABLE IF NOT EXISTS accruals_journal (
    order_id        BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(user_id),
    order_number    VARCHAR(72) NOT NULL UNIQUE,
    status          order_status NOT NULL DEFAULT 'NEW',
    amount          NUMERIC(10, 2) CHECK (status != 'PROCESSED' OR amount IS NOT NULL),
    uploaded_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS accruals_journal_user_id_idx on accruals_journal(user_id);
CREATE INDEX IF NOT EXISTS accruals_journal_status_idx ON accruals_journal(status)
    WHERE status IN ('NEW', 'PROCESSING');

CREATE TABLE IF NOT EXISTS withdrawals_journal (
    withdrawal_id   BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    order_number    VARCHAR(72) NOT NULL UNIQUE,
    user_id         BIGINT NOT NULL REFERENCES users(user_id),
    amount          NUMERIC(10, 2) NOT NULL CHECK (amount > 0), 
    processed_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS withdrawals_journal_user_id_idx on withdrawals_journal(user_id);
