package postgres

import (
	"context"
	"fmt"
	"sms-gateway/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// key used to store the transaction in the context
type transactionContextKey struct{}

// DBTX is an interface that abstracts the database operations for executing SQL commands and queries.
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type ContextDB struct {
	db DBTX
}

// type assertion to ensure ContextDB implements DBTX interface
var _ DBTX = (*ContextDB)(nil)
var _ domain.UnitOfWork = (*ContextDB)(nil)

// NewContextDB creates a new ContextDB instance with the provided DBTX.
func NewContextDB(db DBTX) *ContextDB {
	return &ContextDB{db: db}
}

// Tx implements DBTX.
func (db *ContextDB) Tx(ctx context.Context, atomicFn func(ctx context.Context) error) error {
	pool, ok := db.db.(*pgxpool.Pool)
	if !ok {
		return fmt.Errorf("postgres context db requires *pgxpool.Pool to begin transaction, got %T", db.db)
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	ctxWithTx := contextWithTx(ctx, tx)
	err = atomicFn(ctxWithTx)
	return err
}

func (db *ContextDB) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	return db.executor(ctx).Exec(ctx, sql, arguments...)
}

func (db *ContextDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return db.executor(ctx).Query(ctx, sql, args...)
}

func (db *ContextDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return db.executor(ctx).QueryRow(ctx, sql, args...)
}

func (db *ContextDB) executor(ctx context.Context) DBTX {
	if tx, ok := ctx.Value(transactionContextKey{}).(DBTX); ok && tx != nil {
		return tx
	}
	return db.db
}

func contextWithTx(ctx context.Context, tx DBTX) context.Context {
	return context.WithValue(ctx, transactionContextKey{}, tx)
}
