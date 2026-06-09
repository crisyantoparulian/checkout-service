package repository

import (
	"context"

	"github.com/crisyantoparulian/checkout-service/internal/app/entity"
	"github.com/jmoiron/sqlx"
)

type Promotion interface {
	GetActiveRulesByTargetSKUs(ctx context.Context, skus []string) ([]entity.PromotionRule, error)
}

type promotion struct {
	db *sqlx.DB
}

func NewPromotionRepository(db *sqlx.DB) Promotion {
	if db == nil {
		panic("database is nil")
	}
	return &promotion{db: db}
}

func (r *promotion) GetActiveRulesByTargetSKUs(ctx context.Context, skus []string) ([]entity.PromotionRule, error) {
	var rules []entity.PromotionRule
	query, args, err := sqlx.In(`
		SELECT
			pr.uuid,
			pr.promotion_uuid,
			p.code AS promotion_code,
			p.name AS promotion_name,
			pr.promotion_type,
			pr.target_sku,
			pr.reward_sku,
			pr.min_quantity,
			pr.buy_quantity,
			pr.pay_quantity,
			pr.free_quantity,
			pr.discount_percentage,
			pr.priority,
			pr.is_stackable,
			pr.created_at,
			pr.updated_at
		FROM promotion_rules pr
		JOIN promotions p ON p.uuid = pr.promotion_uuid
		WHERE p.is_active = TRUE
			AND (p.start_at IS NULL OR p.start_at <= CURRENT_TIMESTAMP)
			AND (p.end_at IS NULL OR p.end_at >= CURRENT_TIMESTAMP)
			AND pr.target_sku IN (?)
		ORDER BY pr.priority ASC, pr.created_at ASC
	`, skus)
	if err != nil {
		return nil, err
	}
	query = sqlx.Rebind(sqlx.DOLLAR, query)
	err = r.db.SelectContext(ctx, &rules, query, args...)
	return rules, err
}
