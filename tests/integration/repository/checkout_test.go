//go:build integration

package repository

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/crisyantoparulian/checkout-service/internal/app/entity"
	"github.com/crisyantoparulian/checkout-service/internal/app/repository"
	"github.com/crisyantoparulian/checkout-service/tests/integration/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckoutRepo_Integration(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	repo := repository.NewCheckoutRepository(tdb.DB)

	// Get seeded product and promotion UUIDs
	var macBookUUID, promoUUID, ruleUUID uuid.UUID
	tdb.DB.QueryRowx(`SELECT uuid FROM products WHERE sku = '43N23P'`).Scan(&macBookUUID)
	tdb.DB.QueryRowx(`SELECT uuid FROM promotions WHERE code = 'MACBOOK_FREE_RASPBERRY'`).Scan(&promoUUID)
	tdb.DB.QueryRowx(`SELECT uuid FROM promotion_rules WHERE promotion_uuid = $1`, promoUUID).Scan(&ruleUUID)

	t.Run("Create inserts checkout and assigns UUID", func(t *testing.T) {
		tx, err := tdb.DB.BeginTxx(context.Background(), nil)
		require.NoError(t, err)
		defer tx.Rollback()

		checkout := &entity.Checkout{
			Status:        "COMPLETED",
			SubtotalCents: 539999,
			DiscountCents: 0,
			TotalCents:    539999,
		}
		err = repo.Create(context.Background(), tx, checkout)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, checkout.UUID)
		assert.Equal(t, int64(539999), checkout.SubtotalCents)
	})

	t.Run("CreateItem inserts item with product_uuid", func(t *testing.T) {
		tx, err := tdb.DB.BeginTxx(context.Background(), nil)
		require.NoError(t, err)
		defer tx.Rollback()

		checkout := &entity.Checkout{Status: "COMPLETED", SubtotalCents: 539999, TotalCents: 539999}
		err = repo.Create(context.Background(), tx, checkout)
		require.NoError(t, err)

		item := &entity.CheckoutItem{
			CheckoutUUID:    checkout.UUID,
			ProductUUID:     macBookUUID,
			SKU:             "43N23P",
			ProductName:     "MacBook Pro",
			Quantity:        1,
			UnitPriceCents:  539999,
			TotalPriceCents: 539999,
		}
		err = repo.CreateItem(context.Background(), tx, item)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, item.UUID)
	})

	t.Run("CreatePromotion inserts with rule_snapshot", func(t *testing.T) {
		tx, err := tdb.DB.BeginTxx(context.Background(), nil)
		require.NoError(t, err)
		defer tx.Rollback()

		checkout := &entity.Checkout{Status: "COMPLETED", SubtotalCents: 542999, DiscountCents: 3000, TotalCents: 539999}
		err = repo.Create(context.Background(), tx, checkout)
		require.NoError(t, err)

		snapshot, _ := json.Marshal(map[string]string{"type": "FREE_PRODUCT"})
		promotion := &entity.CheckoutPromotion{
			CheckoutUUID:      checkout.UUID,
			PromotionUUID:     uuid.NullUUID{UUID: promoUUID, Valid: true},
			PromotionRuleUUID: uuid.NullUUID{UUID: ruleUUID, Valid: true},
			PromotionCode:     "MACBOOK_FREE_RASPBERRY",
			PromotionName:     "MacBook Pro Free Raspberry Pi B",
			PromotionType:     "FREE_PRODUCT",
			DiscountCents:     3000,
			RuleSnapshot:      json.RawMessage(snapshot),
		}
		err = repo.CreatePromotion(context.Background(), tx, promotion)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, promotion.UUID)
	})

	t.Run("GetByUUID retrieves created checkout", func(t *testing.T) {
		tx, err := tdb.DB.BeginTxx(context.Background(), nil)
		require.NoError(t, err)

		checkout := &entity.Checkout{Status: "COMPLETED", SubtotalCents: 1000, TotalCents: 1000}
		err = repo.Create(context.Background(), tx, checkout)
		require.NoError(t, err)
		require.NoError(t, tx.Commit())

		result, err := repo.GetByUUID(context.Background(), checkout.UUID)
		require.NoError(t, err)
		assert.Equal(t, checkout.UUID, result.UUID)
		assert.Equal(t, "COMPLETED", result.Status)
		assert.Equal(t, int64(1000), result.TotalCents)
	})

	t.Run("GetByUUID not found returns error", func(t *testing.T) {
		_, err := repo.GetByUUID(context.Background(), uuid.New())
		assert.Error(t, err)
	})

	t.Run("GetItems retrieves items by checkout_uuid", func(t *testing.T) {
		tx, err := tdb.DB.BeginTxx(context.Background(), nil)
		require.NoError(t, err)

		checkout := &entity.Checkout{Status: "COMPLETED", SubtotalCents: 539999, TotalCents: 539999}
		err = repo.Create(context.Background(), tx, checkout)
		require.NoError(t, err)

		item := &entity.CheckoutItem{
			CheckoutUUID:    checkout.UUID,
			ProductUUID:     macBookUUID,
			SKU:             "43N23P",
			ProductName:     "MacBook Pro",
			Quantity:        1,
			UnitPriceCents:  539999,
			TotalPriceCents: 539999,
		}
		err = repo.CreateItem(context.Background(), tx, item)
		require.NoError(t, err)
		require.NoError(t, tx.Commit())

		items, err := repo.GetItems(context.Background(), checkout.UUID)
		require.NoError(t, err)
		assert.Len(t, items, 1)
		assert.Equal(t, macBookUUID, items[0].ProductUUID)
		assert.Equal(t, "43N23P", items[0].SKU)
		assert.Equal(t, 1, items[0].Quantity)
	})

	t.Run("GetPromotions retrieves promotions by checkout_uuid", func(t *testing.T) {
		tx, err := tdb.DB.BeginTxx(context.Background(), nil)
		require.NoError(t, err)

		checkout := &entity.Checkout{Status: "COMPLETED", SubtotalCents: 542999, DiscountCents: 3000, TotalCents: 539999}
		err = repo.Create(context.Background(), tx, checkout)
		require.NoError(t, err)

		snapshot, _ := json.Marshal(map[string]string{"type": "FREE_PRODUCT"})
		promotion := &entity.CheckoutPromotion{
			CheckoutUUID:  checkout.UUID,
			PromotionCode: "TEST_PROMO",
			PromotionName: "Test Promotion",
			PromotionType: "FREE_PRODUCT",
			DiscountCents: 3000,
			RuleSnapshot:  json.RawMessage(snapshot),
		}
		err = repo.CreatePromotion(context.Background(), tx, promotion)
		require.NoError(t, err)
		require.NoError(t, tx.Commit())

		promotions, err := repo.GetPromotions(context.Background(), checkout.UUID)
		require.NoError(t, err)
		assert.Len(t, promotions, 1)
		assert.Equal(t, "TEST_PROMO", promotions[0].PromotionCode)
		assert.Equal(t, int64(3000), promotions[0].DiscountCents)
	})
}
