package entity

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Product struct {
	UUID         uuid.UUID `db:"uuid"`
	SKU          string    `db:"sku"`
	Name         string    `db:"name"`
	PriceCents   int64     `db:"price_cents"`
	InventoryQty int       `db:"inventory_qty"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type Promotion struct {
	UUID        uuid.UUID      `db:"uuid"`
	Code        string         `db:"code"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
	IsActive    bool           `db:"is_active"`
	StartAt     sql.NullTime   `db:"start_at"`
	EndAt       sql.NullTime   `db:"end_at"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}

type PromotionRule struct {
	UUID               uuid.UUID       `db:"uuid" json:"uuid"`
	PromotionUUID      uuid.UUID       `db:"promotion_uuid" json:"promotion_uuid"`
	PromotionCode      string          `db:"promotion_code" json:"promotion_code"`
	PromotionName      string          `db:"promotion_name" json:"promotion_name"`
	PromotionType      string          `db:"promotion_type" json:"promotion_type"`
	TargetSKU          string          `db:"target_sku" json:"target_sku"`
	RewardSKU          sql.NullString  `db:"reward_sku" json:"reward_sku"`
	MinQuantity        sql.NullInt64   `db:"min_quantity" json:"min_quantity"`
	BuyQuantity        sql.NullInt64   `db:"buy_quantity" json:"buy_quantity"`
	PayQuantity        sql.NullInt64   `db:"pay_quantity" json:"pay_quantity"`
	FreeQuantity       sql.NullInt64   `db:"free_quantity" json:"free_quantity"`
	DiscountPercentage sql.NullFloat64 `db:"discount_percentage" json:"discount_percentage"`
	Priority           int             `db:"priority" json:"priority"`
	IsStackable        bool            `db:"is_stackable" json:"is_stackable"`
	CreatedAt          time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time       `db:"updated_at" json:"updated_at"`
}

type Checkout struct {
	UUID          uuid.UUID `db:"uuid"`
	Status        string    `db:"status"`
	SubtotalCents int64     `db:"subtotal_cents"`
	DiscountCents int64     `db:"discount_cents"`
	TotalCents    int64     `db:"total_cents"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

type CheckoutItem struct {
	UUID            uuid.UUID `db:"uuid"`
	CheckoutUUID    uuid.UUID `db:"checkout_uuid"`
	SKU             string    `db:"sku"`
	ProductName     string    `db:"product_name"`
	Quantity        int       `db:"quantity"`
	UnitPriceCents  int64     `db:"unit_price_cents"`
	TotalPriceCents int64     `db:"total_price_cents"`
	CreatedAt       time.Time `db:"created_at"`
}

type CheckoutPromotion struct {
	UUID              uuid.UUID       `db:"uuid"`
	CheckoutUUID      uuid.UUID       `db:"checkout_uuid"`
	PromotionUUID     uuid.NullUUID   `db:"promotion_uuid"`
	PromotionRuleUUID uuid.NullUUID   `db:"promotion_rule_uuid"`
	PromotionCode     string          `db:"promotion_code"`
	PromotionName     string          `db:"promotion_name"`
	PromotionType     string          `db:"promotion_type"`
	DiscountCents     int64           `db:"discount_cents"`
	RuleSnapshot      json.RawMessage `db:"rule_snapshot"`
	CreatedAt         time.Time       `db:"created_at"`
}

type InventoryMovement struct {
	UUID         uuid.UUID      `db:"uuid"`
	SKU          string         `db:"sku"`
	CheckoutUUID uuid.NullUUID  `db:"checkout_uuid"`
	MovementType string         `db:"movement_type"`
	Quantity     int            `db:"quantity"`
	StockBefore  int            `db:"stock_before"`
	StockAfter   int            `db:"stock_after"`
	Note         sql.NullString `db:"note"`
	CreatedAt    time.Time      `db:"created_at"`
}
