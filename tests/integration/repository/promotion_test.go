//go:build integration

package repository

import (
	"context"
	"testing"

	"github.com/crisyantoparulian/checkout-service/internal/app/repository"
	"github.com/crisyantoparulian/checkout-service/tests/integration/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromotionRepo_Integration(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	repo := repository.NewPromotionRepository(tdb.DB)

	var macBookUUID, googleHomeUUID, alexaUUID uuid.UUID
	tdb.DB.QueryRowx(`SELECT uuid FROM products WHERE sku = '43N23P'`).Scan(&macBookUUID)
	tdb.DB.QueryRowx(`SELECT uuid FROM products WHERE sku = '120P90'`).Scan(&googleHomeUUID)
	tdb.DB.QueryRowx(`SELECT uuid FROM products WHERE sku = 'A304SD'`).Scan(&alexaUUID)

	t.Run("GetActiveRulesByTargetProductUUIDs returns rules for MacBook", func(t *testing.T) {
		rules, err := repo.GetActiveRulesByTargetProductUUIDs(context.Background(), []uuid.UUID{macBookUUID})
		require.NoError(t, err)
		assert.Len(t, rules, 1)
		assert.Equal(t, "FREE_PRODUCT", rules[0].PromotionType)
		assert.Equal(t, macBookUUID, rules[0].TargetProductUUID)
		assert.True(t, rules[0].RewardProductUUID.Valid)
	})

	t.Run("GetActiveRulesByTargetProductUUIDs returns rules for Google Home", func(t *testing.T) {
		rules, err := repo.GetActiveRulesByTargetProductUUIDs(context.Background(), []uuid.UUID{googleHomeUUID})
		require.NoError(t, err)
		assert.Len(t, rules, 1)
		assert.Equal(t, "BUY_X_PAY_Y", rules[0].PromotionType)
		assert.True(t, rules[0].BuyQuantity.Valid)
		assert.Equal(t, int64(3), rules[0].BuyQuantity.Int64)
		assert.True(t, rules[0].PayQuantity.Valid)
		assert.Equal(t, int64(2), rules[0].PayQuantity.Int64)
	})

	t.Run("GetActiveRulesByTargetProductUUIDs returns rules for Alexa", func(t *testing.T) {
		rules, err := repo.GetActiveRulesByTargetProductUUIDs(context.Background(), []uuid.UUID{alexaUUID})
		require.NoError(t, err)
		assert.Len(t, rules, 1)
		assert.Equal(t, "PERCENTAGE_DISCOUNT", rules[0].PromotionType)
		assert.True(t, rules[0].DiscountPercentage.Valid)
		assert.Equal(t, float64(10), rules[0].DiscountPercentage.Float64)
	})

	t.Run("GetActiveRulesByTargetProductUUIDs returns multiple rules", func(t *testing.T) {
		rules, err := repo.GetActiveRulesByTargetProductUUIDs(context.Background(), []uuid.UUID{macBookUUID, googleHomeUUID, alexaUUID})
		require.NoError(t, err)
		assert.Len(t, rules, 3)

		types := map[string]bool{}
		for _, r := range rules {
			types[r.PromotionType] = true
		}
		assert.True(t, types["FREE_PRODUCT"])
		assert.True(t, types["BUY_X_PAY_Y"])
		assert.True(t, types["PERCENTAGE_DISCOUNT"])
	})

	t.Run("GetActiveRulesByTargetProductUUIDs with non-existent UUID returns empty", func(t *testing.T) {
		rules, err := repo.GetActiveRulesByTargetProductUUIDs(context.Background(), []uuid.UUID{uuid.New()})
		require.NoError(t, err)
		assert.Len(t, rules, 0)
	})

	t.Run("rules contain promotion code and name", func(t *testing.T) {
		rules, err := repo.GetActiveRulesByTargetProductUUIDs(context.Background(), []uuid.UUID{macBookUUID})
		require.NoError(t, err)
		assert.Len(t, rules, 1)
		assert.Equal(t, "MACBOOK_FREE_RASPBERRY", rules[0].PromotionCode)
		assert.Equal(t, "MacBook Pro Free Raspberry Pi B", rules[0].PromotionName)
	})
}
