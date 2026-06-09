package migration

import (
	"time"

	"github.com/jmoiron/sqlx"
)

// Migration information & command struct
type Migration struct {
	ID        int       `db:"id"`
	Name      string    `db:"name"`
	Batch     int       `db:"batch"`
	CreatedAt time.Time `db:"created_at"`

	Up   func(*sqlx.Tx) error
	Down func(*sqlx.Tx) error

	done bool
}

// Migrator : collection of migration
type Migrator struct {
	db         *sqlx.DB
	Migrations map[string]*Migration
	MaxBatch   int
}
