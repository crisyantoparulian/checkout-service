package repository

import (
	"context"

	"github.com/crisyantoparulian/checkout-service/internal/app/entity"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Product interface {
	GetAll(ctx context.Context, filter ProductFilter) ([]entity.Product, error)
	CountTotal(ctx context.Context) (int64, error)
	GetByUUIDs(ctx context.Context, uuids []uuid.UUID) ([]entity.Product, error)
	GetByUUIDsForUpdate(ctx context.Context, tx *sqlx.Tx, uuids []uuid.UUID) ([]entity.Product, error)
	DeductStock(ctx context.Context, tx *sqlx.Tx, productUUID uuid.UUID, quantity int) error
}

type ProductFilter struct {
	Limit  int
	Offset int
}

type product struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) Product {
	if db == nil {
		panic("database is nil")
	}
	return &product{db: db}
}

func (r *product) GetAll(ctx context.Context, filter ProductFilter) ([]entity.Product, error) {
	var products []entity.Product
	err := r.db.SelectContext(ctx, &products, `
		SELECT uuid, sku, name, price_cents, inventory_qty, created_at, updated_at
		FROM products
		ORDER BY sku ASC
		LIMIT $1 OFFSET $2
	`, filter.Limit, filter.Offset)
	return products, err
}

func (r *product) CountTotal(ctx context.Context) (int64, error) {
	var total int64
	err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM products`)
	return total, err
}

func (r *product) GetByUUIDs(ctx context.Context, uuids []uuid.UUID) ([]entity.Product, error) {
	var products []entity.Product
	query, args, err := sqlx.In(`
		SELECT uuid, sku, name, price_cents, inventory_qty, created_at, updated_at
		FROM products
		WHERE uuid IN (?)
	`, uuids)
	if err != nil {
		return products, err
	}
	query = sqlx.Rebind(sqlx.DOLLAR, query)
	err = r.db.SelectContext(ctx, &products, query, args...)
	return products, err
}

func (r *product) GetByUUIDsForUpdate(ctx context.Context, tx *sqlx.Tx, uuids []uuid.UUID) ([]entity.Product, error) {
	var products []entity.Product
	query, args, err := sqlx.In(`
		SELECT uuid, sku, name, price_cents, inventory_qty, created_at, updated_at
		FROM products
		WHERE uuid IN (?)
		FOR UPDATE
	`, uuids)
	if err != nil {
		return products, err
	}
	query = sqlx.Rebind(sqlx.DOLLAR, query)
	err = tx.SelectContext(ctx, &products, query, args...)
	return products, err
}

func (r *product) DeductStock(ctx context.Context, tx *sqlx.Tx, productUUID uuid.UUID, quantity int) error {
	stmt, err := tx.PreparexContext(ctx, `
		UPDATE products
		SET inventory_qty = inventory_qty - $1,
			updated_at = CURRENT_TIMESTAMP
		WHERE uuid = $2
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, quantity, productUUID)
	return err
}
