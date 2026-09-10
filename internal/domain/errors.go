package domain

import "fmt"

type ErrorCode string

const (
	ErrCodeInsufficientBalance ErrorCode = "INSUFFICIENT_BALANCE"
	ErrCodeAccountNotFound     ErrorCode = "ACCOUNT_NOT_FOUND"
	ErrCodeAccountFrozen       ErrorCode = "ACCOUNT_FROZEN"
	ErrCodeInvalidAmount       ErrorCode = "INVALID_AMOUNT"
	ErrCodeContextCanceled     ErrorCode = "CONTEXT_CANCELED"
	ErrCodeDeadlineExceeded    ErrorCode = "DEADLINE_EXCEEDED"
	ErrCodeDBTxFailure         ErrorCode = "DB_TX_FAILURE"
	ErrCodeIdempotencyConflict ErrorCode = "IDEMPOTENCY_KEY_REUSED"
	ErrCodeInternal            ErrorCode = "INTERNAL_ERROR"
)

type Severity string

const (
	SeverityInfo     Severity = "INFO"
	SeverityWarning  Severity = "WARNING"  
	SeverityCritical Severity = "CRITICAL" 
)

type LedgerError struct {
	Code        ErrorCode
	TraceID     string
	Severity    Severity
	IsRetryable bool
	Message     string
	Err         error 
}

func (e *LedgerError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] trace=%s: %s: %v", e.Code, e.TraceID, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] trace=%s: %s", e.Code, e.TraceID, e.Message)
}

func (e *LedgerError) Unwrap() error {
	return e.Err
}

func NewLedgerError(code ErrorCode, traceID string, severity Severity, isRetryable bool, message string, err error) *LedgerError {
	return &LedgerError{
		Code:        code,
		TraceID:     traceID,
		Severity:    severity,
		IsRetryable: isRetryable,
		Message:     message,
		Err:         err,
	}
}