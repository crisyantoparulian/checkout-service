package checkout

import (
	"database/sql"
	"testing"

	"github.com/crisyantoparulian/checkout-service/internal/app/entity"
	"github.com/crisyantoparulian/checkout-service/internal/pkg/constants"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testProducts() map[string]entity.Product {
	return map[string]entity.Product{
		testMacBookUUID.String():    {UUID: testMacBookUUID, SKU: "43N23P", Name: "MacBook Pro", PriceCents: 539999, InventoryQty: 5},
		testRaspberryUUID.String():  {UUID: testRaspberryUUID, SKU: "234234", Name: "Raspberry Pi B", PriceCents: 3000, InventoryQty: 2},
		testGoogleHomeUUID.String(): {UUID: testGoogleHomeUUID, SKU: "120P90", Name: "Google Home", PriceCents: 4999, InventoryQty: 10},
		testAlexaUUID.String():      {UUID: testAlexaUUID, SKU: "A304SD", Name: "Alexa Speaker", PriceCents: 10950, InventoryQty: 10},
	}
}

func testFreeProductRule() entity.PromotionRule {
	return entity.PromotionRule{
		UUID:              uuid.New(),
		PromotionUUID:     uuid.New(),
		PromotionCode:     "MACBOOK_FREE_RASPBERRY",
		PromotionName:     "MacBook Pro Free Raspberry Pi B",
		PromotionType:     constants.PromotionTypeFreeProduct,
		TargetProductUUID: testMacBookUUID,
		TargetSKU:         "43N23P",
		RewardProductUUID: uuid.NullUUID{UUID: testRaspberryUUID, Valid: true},
		RewardSKU:         sql.NullString{String: "234234", Valid: true},
		MinQuantity:       sql.NullInt64{Int64: 1, Valid: true},
		FreeQuantity:      sql.NullInt64{Int64: 1, Valid: true},
	}
}

func testBuyXPayYRule() entity.PromotionRule {
	return entity.PromotionRule{
		UUID:              uuid.New(),
		PromotionUUID:     uuid.New(),
		PromotionCode:     "GOOGLE_HOME_BUY_3_PAY_2",
		PromotionName:     "Google Home Buy 3 Pay 2",
		PromotionType:     constants.PromotionTypeBuyXPayY,
		TargetProductUUID: testGoogleHomeUUID,
		TargetSKU:         "120P90",
		BuyQuantity:       sql.NullInt64{Int64: 3, Valid: true},
		PayQuantity:       sql.NullInt64{Int64: 2, Valid: true},
	}
}

func testPercentageRule() entity.PromotionRule {
	return entity.PromotionRule{
		UUID:               uuid.New(),
		PromotionUUID:      uuid.New(),
		PromotionCode:      "ALEXA_10_PERCENT_DISCOUNT",
		PromotionName:      "Alexa Speaker 10 Percent Discount",
		PromotionType:      constants.PromotionTypePercentageDiscount,
		TargetProductUUID:  testAlexaUUID,
		TargetSKU:          "A304SD",
		MinQuantity:        sql.NullInt64{Int64: 3, Valid: true},
		DiscountPercentage: sql.NullFloat64{Float64: 10, Valid: true},
	}
}

func TestApplyPromotions(t *testing.T) {
	tests := []struct {
		name         string
		quantities   map[string]int
		products     map[string]entity.Product
		rules        []entity.PromotionRule
		wantDiscount int64
		wantCount    int
		wantErr      bool
	}{
		{
			name:         "free product - MacBook free Raspberry",
			quantities:   map[string]int{testMacBookUUID.String(): 1, testRaspberryUUID.String(): 1},
			products:     testProducts(),
			rules:        []entity.PromotionRule{testFreeProductRule()},
			wantDiscount: 3000,
			wantCount:    1,
		},
		{
			name:         "buy X pay Y - Google Home 3 for 2",
			quantities:   map[string]int{testGoogleHomeUUID.String(): 3},
			products:     testProducts(),
			rules:        []entity.PromotionRule{testBuyXPayYRule()},
			wantDiscount: 4999,
			wantCount:    1,
		},
		{
			name:         "percentage discount - Alexa 10%",
			quantities:   map[string]int{testAlexaUUID.String(): 3},
			products:     testProducts(),
			rules:        []entity.PromotionRule{testPercentageRule()},
			wantDiscount: 3285,
			wantCount:    1,
		},
		{
			name:         "no promotion - quantity below threshold",
			quantities:   map[string]int{testGoogleHomeUUID.String(): 1, testAlexaUUID.String(): 1},
			products:     testProducts(),
			rules:        []entity.PromotionRule{testBuyXPayYRule(), testPercentageRule()},
			wantDiscount: 0,
			wantCount:    0,
		},
		{
			name: "multiple promotions applied",
			quantities: map[string]int{
				testMacBookUUID.String():    1,
				testRaspberryUUID.String():  1,
				testGoogleHomeUUID.String(): 3,
				testAlexaUUID.String():      3,
			},
			products:     testProducts(),
			rules:        []entity.PromotionRule{testFreeProductRule(), testBuyXPayYRule(), testPercentageRule()},
			wantDiscount: 11284,
			wantCount:    3,
		},
		{
			name:         "free product - reward not in cart auto added",
			quantities:   map[string]int{testMacBookUUID.String(): 1},
			products:     testProducts(),
			rules:        []entity.PromotionRule{testFreeProductRule()},
			wantDiscount: 3000,
			wantCount:    1,
		},
		{
			name: "free product - freeQty capped to reward stock",
			quantities: map[string]int{
				testMacBookUUID.String():   5,
				testRaspberryUUID.String(): 1,
			},
			products:     testProducts(),
			rules:        []entity.PromotionRule{testFreeProductRule()},
			wantDiscount: 6000,
			wantCount:    1,
		},
		{
			name:       "free product - invalid rule missing min quantity",
			quantities: map[string]int{testMacBookUUID.String(): 1, testRaspberryUUID.String(): 1},
			products:   testProducts(),
			rules: []entity.PromotionRule{{
				PromotionType:     constants.PromotionTypeFreeProduct,
				TargetProductUUID: testMacBookUUID,
				MinQuantity:       sql.NullInt64{Valid: false},
				FreeQuantity:      sql.NullInt64{Int64: 1, Valid: true},
				RewardProductUUID: uuid.NullUUID{UUID: testRaspberryUUID, Valid: true},
			}},
			wantErr: true,
		},
		{
			name:       "buy X pay Y - invalid rule buyQty <= payQty",
			quantities: map[string]int{testGoogleHomeUUID.String(): 3},
			products:   testProducts(),
			rules: []entity.PromotionRule{{
				PromotionType:     constants.PromotionTypeBuyXPayY,
				TargetProductUUID: testGoogleHomeUUID,
				BuyQuantity:       sql.NullInt64{Int64: 2, Valid: true},
				PayQuantity:       sql.NullInt64{Int64: 3, Valid: true},
			}},
			wantErr: true,
		},
		{
			name:         "percentage - qty below minQuantity",
			quantities:   map[string]int{testAlexaUUID.String(): 2},
			products:     testProducts(),
			rules:        []entity.PromotionRule{testPercentageRule()},
			wantDiscount: 0,
			wantCount:    0,
		},
		{
			name:       "percentage - invalid rule missing minQuantity",
			quantities: map[string]int{testAlexaUUID.String(): 3},
			products:   testProducts(),
			rules: []entity.PromotionRule{{
				PromotionType:      constants.PromotionTypePercentageDiscount,
				TargetProductUUID:  testAlexaUUID,
				MinQuantity:        sql.NullInt64{Valid: false},
				DiscountPercentage: sql.NullFloat64{Float64: 10, Valid: true},
			}},
			wantErr: true,
		},
		{
			name:         "empty rules",
			quantities:   map[string]int{testMacBookUUID.String(): 1},
			products:     testProducts(),
			rules:        []entity.PromotionRule{},
			wantDiscount: 0,
			wantCount:    0,
		},
		{
			name:       "unknown promotion type skipped",
			quantities: map[string]int{testMacBookUUID.String(): 1},
			products:   testProducts(),
			rules: []entity.PromotionRule{{
				PromotionType:     "UNKNOWN_TYPE",
				TargetProductUUID: testMacBookUUID,
			}},
			wantDiscount: 0,
			wantCount:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewUsecase()
			promotions, rewardQuantityByUUID, discount, err := u.applyPromotions(tt.quantities, tt.products, tt.rules)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, rewardQuantityByUUID)
			assert.Equal(t, tt.wantDiscount, discount)
			assert.Len(t, promotions, tt.wantCount)

			if tt.wantCount > 0 {
				for _, p := range promotions {
					assert.NotEmpty(t, p.PromotionCode)
					assert.NotEmpty(t, p.PromotionType)
					assert.Greater(t, p.DiscountCents, int64(0))
				}
			}
		})
	}
}
