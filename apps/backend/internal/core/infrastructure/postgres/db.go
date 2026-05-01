package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

type txKey struct{}

// dbExecutor is satisfied by both *sql.DB and *sql.Tx, allowing repositories
// to work transparently inside or outside an outer transaction.
type dbExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// execer returns the *sql.Tx stored in ctx by RunInTx, falling back to db.
func execer(ctx context.Context, db *sql.DB) dbExecutor {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return tx
	}
	return db
}

// DB wraps *sql.DB and implements domain.Transactor.
type DB struct{ *sql.DB }

// NewDB wraps a raw *sql.DB so it can be passed as a domain.Transactor to services.
func NewDB(db *sql.DB) *DB { return &DB{db} }

// RunInTx executes fn inside a database transaction. If fn returns an error the
// transaction is rolled back; otherwise it is committed.
func (d *DB) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	tx, err := d.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		return err
	}
	return tx.Commit()
}
