package repository

import (
	"context"
	"database/sql"

	"ledger-service/internal/domain"
)

type AccountRepository interface {
	GetForUpdate(ctx context.Context, tx *sql.Tx, accountID int64) (*domain.Account, error)

	UpdateBalance(ctx context.Context, tx *sql.Tx, accountID int64, newBalance int64) error
}

type TransactionRepository interface {
	FindByIdempotencyKey(ctx context.Context, tx *sql.Tx, key string) (*domain.Transaction, error)

	Create(ctx context.Context, tx *sql.Tx, txn *domain.Transaction) error
}

type LedgerEntryRepository interface {
	Create(ctx context.Context, tx *sql.Tx, entry *domain.LedgerEntry) error
}

type TxManager interface {
	BeginTx(ctx context.Context) (*sql.Tx, error)
}