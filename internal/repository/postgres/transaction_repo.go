package postgres

import (
	"context"
	"database/sql"

	"ledger-service/internal/domain"
)

type PostgresTransactionRepository struct{}

func NewPostgresTransactionRepository() *PostgresTransactionRepository {
	return &PostgresTransactionRepository{}
}

func (r *PostgresTransactionRepository) FindByIdempotencyKey(ctx context.Context, tx *sql.Tx, key string) (*domain.Transaction, error) {
	const query = `
		SELECT id, idempotency_key, from_account_id, to_account_id, amount, status, created_at
		FROM transactions
		WHERE idempotency_key = $1`

	var t domain.Transaction
	err := tx.QueryRowContext(ctx, query, key).Scan(
		&t.ID, &t.IdempotencyKey, &t.FromAccountID, &t.ToAccountID, &t.Amount, &t.Status, &t.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *PostgresTransactionRepository) Create(ctx context.Context, tx *sql.Tx, txn *domain.Transaction) error {
	const query = `
		INSERT INTO transactions (id, idempotency_key, from_account_id, to_account_id, amount, status)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := tx.ExecContext(ctx, query,
		txn.ID, txn.IdempotencyKey, txn.FromAccountID, txn.ToAccountID, txn.Amount, txn.Status,
	)
	return err
}