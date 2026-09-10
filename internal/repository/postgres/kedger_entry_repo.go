package postgres

import (
	"context"
	"database/sql"

	"ledger-service/internal/domain"
)

type PostgresLedgerEntryRepository struct{}

func NewPostgresLedgerEntryRepository() *PostgresLedgerEntryRepository {
	return &PostgresLedgerEntryRepository{}
}

func (r *PostgresLedgerEntryRepository) Create(ctx context.Context, tx *sql.Tx, entry *domain.LedgerEntry) error {
	const query = `
		INSERT INTO ledger_entries (transaction_id, account_id, entry_type, amount)
		VALUES ($1, $2, $3, $4)`

	_, err := tx.ExecContext(ctx, query, entry.TransactionID, entry.AccountID, entry.Type, entry.Amount)
	return err
}