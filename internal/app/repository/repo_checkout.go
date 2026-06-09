package repository

import (
	"context"
	"encoding/json"

	"github.com/crisyantoparulian/checkout-service/internal/app/entity"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Checkout interface {
	Create(ctx context.Context, tx *sqlx.Tx, checkout *entity.Checkout) error
	CreateItem(ctx context.Context, tx *sqlx.Tx, item *entity.CheckoutItem) error
	CreatePromotion(ctx context.Context, tx *sqlx.Tx, promotion *entity.CheckoutPromotion) error
	GetByUUID(ctx context.Context, checkoutUUID uuid.UUID) (entity.Checkout, error)
	GetItems(ctx context.Context, checkoutUUID uuid.UUID) ([]entity.CheckoutItem, error)
	GetPromotions(ctx context.Context, checkoutUUID uuid.UUID) ([]entity.CheckoutPromotion, error)
}

type checkout struct {
	db *sqlx.DB
}

func NewCheckoutRepository(db *sqlx.DB) Checkout {
	if db == nil {
		panic("database is nil")
	}
	return &checkout{db: db}
}

func (r *checkout) Create(ctx context.Context, tx *sqlx.Tx, checkout *entity.Checkout) error {
	stmt, err := tx.PreparexContext(ctx, `
		INSERT INTO checkouts (status, subtotal_cents, discount_cents, total_cents, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING uuid
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.GetContext(ctx, &checkout.UUID, checkout.Status, checkout.SubtotalCents, checkout.DiscountCents, checkout.TotalCents)
}

func (r *checkout) CreateItem(ctx context.Context, tx *sqlx.Tx, item *entity.CheckoutItem) error {
	stmt, err := tx.PreparexContext(ctx, `
		INSERT INTO checkout_items (checkout_uuid, sku, product_name, quantity, unit_price_cents, total_price_cents, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
		RETURNING uuid
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.GetContext(ctx, &item.UUID, item.CheckoutUUID, item.SKU, item.ProductName, item.Quantity, item.UnitPriceCents, item.TotalPriceCents)
}

func (r *checkout) CreatePromotion(ctx context.Context, tx *sqlx.Tx, promotion *entity.CheckoutPromotion) error {
	stmt, err := tx.PreparexContext(ctx, `
		INSERT INTO checkout_promotions (
			checkout_uuid, promotion_uuid, promotion_rule_uuid, promotion_code, promotion_name,
			promotion_type, discount_cents, rule_snapshot, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP)
		RETURNING uuid
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.GetContext(ctx, &promotion.UUID, promotion.CheckoutUUID, promotion.PromotionUUID, promotion.PromotionRuleUUID, promotion.PromotionCode, promotion.PromotionName, promotion.PromotionType, promotion.DiscountCents, json.RawMessage(promotion.RuleSnapshot))
}

func (r *checkout) GetByUUID(ctx context.Context, checkoutUUID uuid.UUID) (entity.Checkout, error) {
	var checkout entity.Checkout
	err := r.db.GetContext(ctx, &checkout, `
		SELECT uuid, status, subtotal_cents, discount_cents, total_cents, created_at, updated_at
		FROM checkouts
		WHERE uuid = $1
		LIMIT 1
	`, checkoutUUID)
	return checkout, err
}

func (r *checkout) GetItems(ctx context.Context, checkoutUUID uuid.UUID) ([]entity.CheckoutItem, error) {
	var items []entity.CheckoutItem
	err := r.db.SelectContext(ctx, &items, `
		SELECT uuid, checkout_uuid, sku, product_name, quantity, unit_price_cents, total_price_cents, created_at
		FROM checkout_items
		WHERE checkout_uuid = $1
		ORDER BY created_at ASC
	`, checkoutUUID)
	return items, err
}

func (r *checkout) GetPromotions(ctx context.Context, checkoutUUID uuid.UUID) ([]entity.CheckoutPromotion, error) {
	var promotions []entity.CheckoutPromotion
	err := r.db.SelectContext(ctx, &promotions, `
		SELECT uuid, checkout_uuid, promotion_uuid, promotion_rule_uuid, promotion_code, promotion_name, promotion_type, discount_cents, rule_snapshot, created_at
		FROM checkout_promotions
		WHERE checkout_uuid = $1
		ORDER BY created_at ASC
	`, checkoutUUID)
	return promotions, err
}
