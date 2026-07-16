package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/harundarat/investment-ledger-platform/internal/service"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type transactionContextKey struct{}

type dbtx interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type transactionManager struct {
	db *pgxpool.Pool
}

func NewTransactionManager(db *pgxpool.Pool) service.TransactionManager {
	return &transactionManager{db: db}
}

func (tm *transactionManager) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	tx, err := tm.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer rollbackTx(ctx, tx)

	if err := fn(context.WithValue(ctx, transactionContextKey{}, tx)); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func databaseFromContext(ctx context.Context, db *pgxpool.Pool) dbtx {
	if tx, ok := ctx.Value(transactionContextKey{}).(pgx.Tx); ok {
		return tx
	}

	return db
}

func rollbackTx(ctx context.Context, tx pgx.Tx) {
	if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		log.Printf("rollback transaction: %v", err)
	}
}

var _ service.TransactionManager = (*transactionManager)(nil)
