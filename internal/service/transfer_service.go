package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"ledger-service/internal/domain"
	"ledger-service/internal/repository"
)

type TransferService struct {
	db          *sql.DB
	accountRepo repository.AccountRepository
	txRepo      repository.TransactionRepository
	entryRepo   repository.LedgerEntryRepository
}

func NewTransferService(
	db *sql.DB,
	accountRepo repository.AccountRepository,
	txRepo repository.TransactionRepository,
	entryRepo repository.LedgerEntryRepository,
) *TransferService {
	return &TransferService{
		db:          db,
		accountRepo: accountRepo,
		txRepo:      txRepo,
		entryRepo:   entryRepo,
	}
}

func (s *TransferService) Transfer(ctx context.Context, req domain.TransferRequest, traceID string) (*domain.TransferResult, error) {
	if req.Amount <= 0 {
		return nil, domain.NewLedgerError(
			domain.ErrCodeInvalidAmount, traceID, domain.SeverityInfo, false,
			"transfer amount must be positive", nil,
		)
	}
	if req.FromAccountID == req.ToAccountID {
		return nil, domain.NewLedgerError(
			domain.ErrCodeInvalidAmount, traceID, domain.SeverityInfo, false,
			"cannot transfer to the same account", nil,
		)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, domain.NewLedgerError(
			domain.ErrCodeDBTxFailure, traceID, domain.SeverityCritical, true,
			"failed to begin transaction", err,
		)
	}

	defer tx.Rollback()

	existing, err := s.txRepo.FindByIdempotencyKey(ctx, tx, req.IdempotencyKey)
	if err != nil {
		return nil, domain.NewLedgerError(
			domain.ErrCodeDBTxFailure, traceID, domain.SeverityCritical, true,
			"failed to check idempotency key", err,
		)
	}
	if existing != nil {
		return &domain.TransferResult{
			TransactionID: existing.ID,
			Status:        existing.Status,
			Replayed:      true,
		}, nil
	}

	firstID, secondID := req.FromAccountID, req.ToAccountID
	if firstID > secondID {
		firstID, secondID = secondID, firstID
	}

	firstAcc, err := s.accountRepo.GetForUpdate(ctx, tx, firstID)
	if err != nil {
		return nil, s.wrapAccountLookupErr(err, firstID, traceID)
	}
	secondAcc, err := s.accountRepo.GetForUpdate(ctx, tx, secondID)
	if err != nil {
		return nil, s.wrapAccountLookupErr(err, secondID, traceID)
	}

	var fromAcc, toAcc *domain.Account
	if firstAcc.ID == req.FromAccountID {
		fromAcc, toAcc = firstAcc, secondAcc
	} else {
		fromAcc, toAcc = secondAcc, firstAcc
	}

	if fromAcc.Status != domain.AccountStatusActive {
		return nil, domain.NewLedgerError(
			domain.ErrCodeAccountFrozen, traceID, domain.SeverityInfo, false,
			"sender account is not active", nil,
		)
	}
	if toAcc.Status != domain.AccountStatusActive {
		return nil, domain.NewLedgerError(
			domain.ErrCodeAccountFrozen, traceID, domain.SeverityInfo, false,
			"receiver account is not active", nil,
		)
	}
	if fromAcc.Balance < req.Amount {
		return nil, domain.NewLedgerError(
			domain.ErrCodeInsufficientBalance, traceID, domain.SeverityInfo, false,
			"sender has insufficient balance", nil,
		)
	}

	newFromBalance := fromAcc.Balance - req.Amount
	newToBalance := toAcc.Balance + req.Amount

	if err := s.accountRepo.UpdateBalance(ctx, tx, fromAcc.ID, newFromBalance); err != nil {
		return nil, domain.NewLedgerError(
			domain.ErrCodeDBTxFailure, traceID, domain.SeverityCritical, true,
			"failed to debit sender account", err,
		)
	}
	if err := s.accountRepo.UpdateBalance(ctx, tx, toAcc.ID, newToBalance); err != nil {
		return nil, domain.NewLedgerError(
			domain.ErrCodeDBTxFailure, traceID, domain.SeverityCritical, true,
			"failed to credit receiver account", err,
		)
	}

	txID := uuid.New().String()

	txRecord := &domain.Transaction{
		ID:             txID,
		IdempotencyKey: req.IdempotencyKey,
		FromAccountID:  req.FromAccountID,
		ToAccountID:    req.ToAccountID,
		Amount:         req.Amount,
		Status:         domain.TransactionStatusCompleted,
	}

	if err := s.txRepo.Create(ctx, tx, txRecord); err != nil {
		return nil, domain.NewLedgerError(
			domain.ErrCodeIdempotencyConflict, traceID, domain.SeverityWarning, false,
			"idempotency key was used concurrently", err,
		)
	}

	debitEntry := &domain.LedgerEntry{
		TransactionID: txID,
		AccountID:     fromAcc.ID,
		Type:          domain.EntryTypeDebit,
		Amount:        req.Amount,
	}
	creditEntry := &domain.LedgerEntry{
		TransactionID: txID,
		AccountID:     toAcc.ID,
		Type:          domain.EntryTypeCredit,
		Amount:        req.Amount,
	}

	if err := s.entryRepo.Create(ctx, tx, debitEntry); err != nil {
		return nil, domain.NewLedgerError(
			domain.ErrCodeDBTxFailure, traceID, domain.SeverityCritical, true,
			"failed to write debit entry", err,
		)
	}
	if err := s.entryRepo.Create(ctx, tx, creditEntry); err != nil {
		return nil, domain.NewLedgerError(
			domain.ErrCodeDBTxFailure, traceID, domain.SeverityCritical, true,
			"failed to write credit entry", err,
		)
	}

	if err := ctx.Err(); err != nil {
		code := domain.ErrCodeContextCanceled
		if errors.Is(err, context.DeadlineExceeded) {
			code = domain.ErrCodeDeadlineExceeded
		}
		return nil, domain.NewLedgerError(
			code, traceID, domain.SeverityWarning, true,
			"context cancelled before commit", err,
		)
	}

	if err := tx.Commit(); err != nil {
		return nil, domain.NewLedgerError(
			domain.ErrCodeDBTxFailure, traceID, domain.SeverityCritical, true,
			"failed to commit transaction", err,
		)
	}

	return &domain.TransferResult{
		TransactionID: txID,
		Status:        domain.TransactionStatusCompleted,
		Replayed:      false,
	}, nil
}

func (s *TransferService) wrapAccountLookupErr(err error, accountID int64, traceID string) error {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.NewLedgerError(
			domain.ErrCodeAccountNotFound, traceID, domain.SeverityInfo, false,
			"account not found", err,
		)
	}
	return domain.NewLedgerError(
		domain.ErrCodeDBTxFailure, traceID, domain.SeverityCritical, true,
		"failed to lock account for update", err,
	)
}
