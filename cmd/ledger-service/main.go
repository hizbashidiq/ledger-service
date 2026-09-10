package main

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"ledger-service/internal/domain"
	"ledger-service/internal/repository/postgres"
	"ledger-service/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	
	DB_URL := "postgres://postgres:postgres@localhost:5432/ledger?sslmode=disable"

	db, err := sql.Open("postgres", DB_URL)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	accountRepo := postgres.NewPostgresAccountRepository()
	txRepo := postgres.NewPostgresTransactionRepository()
	entryRepo := postgres.NewPostgresLedgerEntryRepository()

	transferService := service.NewTransferService(db, accountRepo, txRepo, entryRepo)

	//Demo call: transfer 10000 from account 1 to account 2
	traceID := uuid.New().String()
	ctx, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()

	req := domain.TransferRequest{
		IdempotencyKey: uuid.New().String(),
		FromAccountID:  1,
		ToAccountID:    2,
		Amount:         10000,
	}

	result, err := transferService.Transfer(ctx, req, traceID)
	if err != nil {
		logger.Error("transfer failed", "trace_id", traceID, "error", err)
		os.Exit(1)
	}

	logger.Info("transfer succeeded",
		"trace_id", traceID,
		"transaction_id", result.TransactionID,
		"replayed", result.Replayed,
	)
}