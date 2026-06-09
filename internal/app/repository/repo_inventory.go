package repository

import (
	"context"

	"github.com/crisyantoparulian/checkout-service/internal/app/entity"
	"github.com/jmoiron/sqlx"
)

type Inventory interface {
	CreateMovement(ctx context.Context, tx *sqlx.Tx, movement *entity.InventoryMovement) error
}

type inventory struct {
	db *sqlx.DB
}

func NewInventoryRepository(db *sqlx.DB) Inventory {
	if db == nil {
		panic("database is nil")
	}
	return &inventory{db: db}
}

func (r *inventory) CreateMovement(ctx context.Context, tx *sqlx.Tx, movement *entity.InventoryMovement) error {
	stmt, err := tx.PreparexContext(ctx, `
		INSERT INTO inventory_movements (sku, checkout_uuid, movement_type, quantity, stock_before, stock_after, note, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP)
		RETURNING uuid
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.GetContext(ctx, &movement.UUID, movement.SKU, movement.CheckoutUUID, movement.MovementType, movement.Quantity, movement.StockBefore, movement.StockAfter, movement.Note)
}
