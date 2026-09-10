CREATE TABLE accounts (
    id          BIGSERIAL PRIMARY KEY,
    owner_name  TEXT NOT NULL,
    balance     BIGINT NOT NULL CHECK (balance >= 0),
    status      TEXT NOT NULL DEFAULT 'ACTIVE'
                CHECK (status IN ('ACTIVE', 'FROZEN')),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE transactions (
    id               UUID PRIMARY KEY,
    idempotency_key  TEXT NOT NULL UNIQUE, 
    from_account_id  BIGINT NOT NULL REFERENCES accounts(id),
    to_account_id    BIGINT NOT NULL REFERENCES accounts(id),
    amount           BIGINT NOT NULL CHECK (amount > 0),
    status           TEXT NOT NULL
                     CHECK (status IN ('COMPLETED', 'FAILED')),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE ledger_entries (
    id              BIGSERIAL PRIMARY KEY,
    transaction_id  UUID NOT NULL REFERENCES transactions(id),
    account_id      BIGINT NOT NULL REFERENCES accounts(id),
    entry_type      TEXT NOT NULL CHECK (entry_type IN ('DEBIT', 'CREDIT')),
    amount          BIGINT NOT NULL CHECK (amount > 0),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_transactions_idempotency_key ON transactions(idempotency_key);
CREATE INDEX idx_ledger_entries_account_id ON ledger_entries(account_id);