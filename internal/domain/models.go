package domain

import "time"

type AccountStatus string

const (
	AccountStatusActive AccountStatus = "ACTIVE"
	AccountStatusFrozen AccountStatus = "FROZEN"
)

type Account struct {
	ID        int64
	OwnerName string
	Balance   int64 
	Status    AccountStatus
	UpdatedAt time.Time
}

type EntryType string

const (
	EntryTypeDebit  EntryType = "DEBIT"  
	EntryTypeCredit EntryType = "CREDIT" 
)

type LedgerEntry struct {
	ID            int64
	TransactionID string
	AccountID     int64
	Type          EntryType
	Amount        int64 
	CreatedAt     time.Time
}

type TransactionStatus string

const (
	TransactionStatusCompleted TransactionStatus = "COMPLETED"
	TransactionStatusFailed    TransactionStatus = "FAILED"
)

type Transaction struct {
	ID             string 
	IdempotencyKey string
	FromAccountID  int64
	ToAccountID    int64
	Amount         int64
	Status         TransactionStatus
	CreatedAt      time.Time
}

type TransferRequest struct {
	IdempotencyKey string
	FromAccountID  int64
	ToAccountID    int64
	Amount         int64 
}

type TransferResult struct {
	TransactionID string
	Status        TransactionStatus
	Replayed      bool 
}