package transaction

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// ManagerInterface defines the interface for transaction management.
type ManagerInterface interface {
	WithTx(ctx context.Context, fn func(tx *sqlx.Tx) error) error
}

// Manager provides a general-purpose transaction wrapper
// that any usecase can use regardless of which repository it calls.
type Manager struct {
	db *sqlx.DB
}

func NewManager(db *sqlx.DB) *Manager {
	if db == nil {
		panic("database is nil")
	}
	return &Manager{db: db}
}

// WithTx begins a transaction, calls fn with the tx, and automatically
// rolls back on error or commits on success.
func (m *Manager) WithTx(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	tx, err := m.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	if err = fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
