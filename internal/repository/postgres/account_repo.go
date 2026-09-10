package postgres

import (
	"context"
	"database/sql"

	"ledger-service/internal/domain"
)

type PostgresAccountRepository struct{}

func NewPostgresAccountRepository() *PostgresAccountRepository {
	return &PostgresAccountRepository{}
}

func (r *PostgresAccountRepository) GetForUpdate(ctx context.Context, tx *sql.Tx, accountID int64) (*domain.Account, error) {
	const query = `
		SELECT id, owner_name, balance, status, updated_at
		FROM accounts
		WHERE id = $1
		FOR UPDATE`

	var acc domain.Account
	err := tx.QueryRowContext(ctx, query, accountID).Scan(
		&acc.ID, &acc.OwnerName, &acc.Balance, &acc.Status, &acc.UpdatedAt,
	)
	if err != nil {
		return nil, err 
	}
	return &acc, nil
}

func (r *PostgresAccountRepository) UpdateBalance(ctx context.Context, tx *sql.Tx, accountID int64, newBalance int64) error {
	const query = `
		UPDATE accounts
		SET balance = $1, updated_at = now()
		WHERE id = $2`

	result, err := tx.ExecContext(ctx, query, newBalance, accountID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows 
	}
	return nil
}