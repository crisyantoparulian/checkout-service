package repository

import (
	"context"

	"github.com/crisyantoparulian/checkout-service/internal/app/entity"
	"github.com/jmoiron/sqlx"
)

type Product interface {
	GetAll(ctx context.Context, filter ProductFilter) ([]entity.Product, error)
	CountTotal(ctx context.Context) (int64, error)
	GetBySKUsForUpdate(ctx context.Context, tx *sqlx.Tx, skus []string) ([]entity.Product, error)
	DeductStock(ctx context.Context, tx *sqlx.Tx, sku string, quantity int) error
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

func (r *product) GetBySKUsForUpdate(ctx context.Context, tx *sqlx.Tx, skus []string) ([]entity.Product, error) {
	var products []entity.Product
	query, args, err := sqlx.In(`
		SELECT uuid, sku, name, price_cents, inventory_qty, created_at, updated_at
		FROM products
		WHERE sku IN (?)
		FOR UPDATE
	`, skus)
	if err != nil {
		return products, err
	}
	query = sqlx.Rebind(sqlx.DOLLAR, query)
	err = tx.SelectContext(ctx, &products, query, args...)
	return products, err
}

func (r *product) DeductStock(ctx context.Context, tx *sqlx.Tx, sku string, quantity int) error {
	stmt, err := tx.PreparexContext(ctx, `
		UPDATE products
		SET inventory_qty = inventory_qty - $1,
			updated_at = CURRENT_TIMESTAMP
		WHERE sku = $2
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, quantity, sku)
	return err
}
