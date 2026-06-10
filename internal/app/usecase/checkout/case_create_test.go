package checkout

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/crisyantoparulian/checkout-service/internal/app/entity"
	"github.com/crisyantoparulian/checkout-service/internal/app/repository/mocks"
	"github.com/crisyantoparulian/checkout-service/internal/pkg/apperror"
	"github.com/crisyantoparulian/checkout-service/internal/pkg/constants"
	txmocks "github.com/crisyantoparulian/checkout-service/internal/pkg/transaction/mocks"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var (
	testMacBookUUID    = uuid.New()
	testRaspberryUUID  = uuid.New()
	testGoogleHomeUUID = uuid.New()
	testAlexaUUID      = uuid.New()
)

func testProductByUUID() map[string]entity.Product {
	return map[string]entity.Product{
		testMacBookUUID.String():    {UUID: testMacBookUUID, SKU: "43N23P", Name: "MacBook Pro", PriceCents: 539999, InventoryQty: 5},
		testRaspberryUUID.String():  {UUID: testRaspberryUUID, SKU: "234234", Name: "Raspberry Pi B", PriceCents: 3000, InventoryQty: 2},
		testGoogleHomeUUID.String(): {UUID: testGoogleHomeUUID, SKU: "120P90", Name: "Google Home", PriceCents: 4999, InventoryQty: 10},
		testAlexaUUID.String():      {UUID: testAlexaUUID, SKU: "A304SD", Name: "Alexa Speaker", PriceCents: 10950, InventoryQty: 10},
	}
}

// --- buildCheckoutQuantities ---

func TestBuildCheckoutQuantities(t *testing.T) {
	tests := []struct {
		name              string
		req               CreateCheckoutRequest
		wantQuantityByKey map[string]int
		wantUUIDCount     int
		wantErr           bool
		wantErrCode       string
	}{
		{
			name: "valid single item",
			req: CreateCheckoutRequest{
				Items: []CheckoutItemRequest{
					{ProductUUID: testMacBookUUID.String(), Quantity: 2},
				},
			},
			wantQuantityByKey: map[string]int{testMacBookUUID.String(): 2},
			wantUUIDCount:     1,
		},
		{
			name: "valid multiple items",
			req: CreateCheckoutRequest{
				Items: []CheckoutItemRequest{
					{ProductUUID: testMacBookUUID.String(), Quantity: 1},
					{ProductUUID: testGoogleHomeUUID.String(), Quantity: 3},
				},
			},
			wantQuantityByKey: map[string]int{
				testMacBookUUID.String():    1,
				testGoogleHomeUUID.String(): 3,
			},
			wantUUIDCount: 2,
		},
		{
			name: "duplicate product UUIDs aggregate quantities",
			req: CreateCheckoutRequest{
				Items: []CheckoutItemRequest{
					{ProductUUID: testMacBookUUID.String(), Quantity: 2},
					{ProductUUID: testMacBookUUID.String(), Quantity: 3},
				},
			},
			wantQuantityByKey: map[string]int{testMacBookUUID.String(): 5},
			wantUUIDCount:     1,
		},
		{
			name: "zero quantity returns error",
			req: CreateCheckoutRequest{
				Items: []CheckoutItemRequest{
					{ProductUUID: testMacBookUUID.String(), Quantity: 0},
				},
			},
			wantErr:     true,
			wantErrCode: "INVALID_QUANTITY",
		},
		{
			name: "negative quantity returns error",
			req: CreateCheckoutRequest{
				Items: []CheckoutItemRequest{
					{ProductUUID: testMacBookUUID.String(), Quantity: -1},
				},
			},
			wantErr:     true,
			wantErrCode: "INVALID_QUANTITY",
		},
		{
			name: "invalid UUID format returns error",
			req: CreateCheckoutRequest{
				Items: []CheckoutItemRequest{
					{ProductUUID: "not-a-uuid", Quantity: 1},
				},
			},
			wantErr:     true,
			wantErrCode: "PRODUCT_NOT_FOUND",
		},
		{
			name:              "empty items returns empty map",
			req:               CreateCheckoutRequest{Items: []CheckoutItemRequest{}},
			wantQuantityByKey: map[string]int{},
			wantUUIDCount:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quantityByUUID, productUUIDs, err := buildCheckoutQuantities(tt.req)
			if tt.wantErr {
				require.Error(t, err)
				appErr, ok := err.(*apperror.ApplicationError)
				require.True(t, ok)
				assert.Equal(t, tt.wantErrCode, appErr.ErrorCode)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantQuantityByKey, quantityByUUID)
			assert.Len(t, productUUIDs, tt.wantUUIDCount)
		})
	}
}

// --- validateAndRefreshProducts ---

func TestValidateAndRefreshProducts(t *testing.T) {
	tests := []struct {
		name        string
		quantityMap map[string]int
		products    []entity.Product
		existingMap map[string]entity.Product
		wantErr     bool
		wantErrCode string
	}{
		{
			name:        "all products found with sufficient stock",
			quantityMap: map[string]int{testMacBookUUID.String(): 3},
			products: []entity.Product{
				{UUID: testMacBookUUID, SKU: "43N23P", Name: "MacBook Pro", InventoryQty: 5},
			},
			existingMap: nil,
		},
		{
			name:        "product not found",
			quantityMap: map[string]int{testMacBookUUID.String(): 1},
			products:    []entity.Product{},
			existingMap: nil,
			wantErr:     true,
			wantErrCode: "PRODUCT_NOT_FOUND",
		},
		{
			name:        "insufficient stock",
			quantityMap: map[string]int{testMacBookUUID.String(): 10},
			products: []entity.Product{
				{UUID: testMacBookUUID, SKU: "43N23P", Name: "MacBook Pro", InventoryQty: 5},
			},
			existingMap: nil,
			wantErr:     true,
			wantErrCode: "INSUFFICIENT_STOCK",
		},
		{
			name:        "exact stock match succeeds",
			quantityMap: map[string]int{testMacBookUUID.String(): 5},
			products: []entity.Product{
				{UUID: testMacBookUUID, SKU: "43N23P", Name: "MacBook Pro", InventoryQty: 5},
			},
			existingMap: nil,
		},
		{
			name:        "nil existingMap creates new map",
			quantityMap: map[string]int{testMacBookUUID.String(): 1},
			products: []entity.Product{
				{UUID: testMacBookUUID, SKU: "43N23P", Name: "MacBook Pro", InventoryQty: 5},
			},
			existingMap: nil,
		},
		{
			name:        "existing map is updated in place",
			quantityMap: map[string]int{testMacBookUUID.String(): 1},
			products: []entity.Product{
				{UUID: testMacBookUUID, SKU: "43N23P", Name: "MacBook Pro Updated", InventoryQty: 3},
			},
			existingMap: map[string]entity.Product{
				testMacBookUUID.String(): {UUID: testMacBookUUID, SKU: "43N23P", Name: "MacBook Pro", InventoryQty: 5},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validateAndRefreshProducts(tt.quantityMap, tt.products, tt.existingMap)
			if tt.wantErr {
				require.Error(t, err)
				appErr, ok := err.(*apperror.ApplicationError)
				require.True(t, ok)
				assert.Equal(t, tt.wantErrCode, appErr.ErrorCode)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}

// --- buildCheckoutItems ---

func TestBuildCheckoutItems(t *testing.T) {
	tests := []struct {
		name         string
		req          CreateCheckoutRequest
		productMap   map[string]entity.Product
		wantCount    int
		wantSubtotal int64
	}{
		{
			name: "single item",
			req: CreateCheckoutRequest{
				Items: []CheckoutItemRequest{
					{ProductUUID: testMacBookUUID.String(), Quantity: 1},
				},
			},
			productMap:   testProductByUUID(),
			wantCount:    1,
			wantSubtotal: 539999,
		},
		{
			name: "multiple items with correct subtotal",
			req: CreateCheckoutRequest{
				Items: []CheckoutItemRequest{
					{ProductUUID: testMacBookUUID.String(), Quantity: 1},
					{ProductUUID: testGoogleHomeUUID.String(), Quantity: 3},
				},
			},
			productMap:   testProductByUUID(),
			wantCount:    2,
			wantSubtotal: 539999 + 4999*3,
		},
		{
			name:         "empty items returns empty slice and zero subtotal",
			req:          CreateCheckoutRequest{Items: []CheckoutItemRequest{}},
			productMap:   testProductByUUID(),
			wantCount:    0,
			wantSubtotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, subtotal := buildCheckoutItems(tt.req, tt.productMap)
			assert.Len(t, items, tt.wantCount)
			assert.Equal(t, tt.wantSubtotal, subtotal)

			if tt.wantCount > 0 {
				for _, item := range items {
					assert.NotEqual(t, uuid.Nil, item.ProductUUID)
					assert.NotEmpty(t, item.SKU)
					assert.NotEmpty(t, item.ProductName)
					assert.Equal(t, tt.productMap[item.ProductUUID.String()].PriceCents*int64(item.Quantity), item.TotalPriceCents)
				}
			}
		})
	}
}

// --- appendLockedUUIDs ---

func TestAppendLockedUUIDs(t *testing.T) {
	rewardUUID := uuid.New()

	tests := []struct {
		name           string
		productUUIDs   []uuid.UUID
		rules          []entity.PromotionRule
		wantLen        int
		wantContains   []uuid.UUID
		wantNotContain []uuid.UUID
	}{
		{
			name:         "no rules returns only product UUIDs",
			productUUIDs: []uuid.UUID{testMacBookUUID},
			rules:        nil,
			wantLen:      1,
			wantContains: []uuid.UUID{testMacBookUUID},
		},
		{
			name:         "reward UUID appended",
			productUUIDs: []uuid.UUID{testMacBookUUID},
			rules: []entity.PromotionRule{
				{RewardProductUUID: uuid.NullUUID{UUID: rewardUUID, Valid: true}},
			},
			wantLen:      2,
			wantContains: []uuid.UUID{testMacBookUUID, rewardUUID},
		},
		{
			name:         "reward already in productUUIDs is deduplicated",
			productUUIDs: []uuid.UUID{testMacBookUUID, testRaspberryUUID},
			rules: []entity.PromotionRule{
				{RewardProductUUID: uuid.NullUUID{UUID: testRaspberryUUID, Valid: true}},
			},
			wantLen:      2,
			wantContains: []uuid.UUID{testMacBookUUID, testRaspberryUUID},
		},
		{
			name:         "invalid RewardProductUUID skipped",
			productUUIDs: []uuid.UUID{testMacBookUUID},
			rules: []entity.PromotionRule{
				{RewardProductUUID: uuid.NullUUID{Valid: false}},
			},
			wantLen:        1,
			wantNotContain: []uuid.UUID{uuid.Nil},
		},
		{
			name:         "multiple rules with overlapping rewards deduplicated",
			productUUIDs: []uuid.UUID{testMacBookUUID},
			rules: []entity.PromotionRule{
				{RewardProductUUID: uuid.NullUUID{UUID: rewardUUID, Valid: true}},
				{RewardProductUUID: uuid.NullUUID{UUID: rewardUUID, Valid: true}},
			},
			wantLen:      2,
			wantContains: []uuid.UUID{testMacBookUUID, rewardUUID},
		},
		{
			name:         "empty productUUIDs with rules returns only rewards",
			productUUIDs: []uuid.UUID{},
			rules: []entity.PromotionRule{
				{RewardProductUUID: uuid.NullUUID{UUID: rewardUUID, Valid: true}},
			},
			wantLen:      1,
			wantContains: []uuid.UUID{rewardUUID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := appendLockedUUIDs(tt.productUUIDs, tt.rules)
			assert.Len(t, result, tt.wantLen)

			resultSet := map[uuid.UUID]bool{}
			for _, id := range result {
				resultSet[id] = true
			}
			for _, id := range tt.wantContains {
				assert.True(t, resultSet[id], "expected %s in result", id)
			}
			for _, id := range tt.wantNotContain {
				assert.False(t, resultSet[id], "did not expect %s in result", id)
			}
		})
	}
}

// --- Create (integration of all components) ---

func TestCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTxManager := txmocks.NewMockManagerInterface(ctrl)
	mockCheckout := mocks.NewMockCheckout(ctrl)
	mockProduct := mocks.NewMockProduct(ctrl)
	mockPromotion := mocks.NewMockPromotion(ctrl)
	mockInventory := mocks.NewMockInventory(ctrl)

	newUsecase := func() *usecase {
		return NewUsecase().
			SetTxManager(mockTxManager).
			SetCheckoutRepository(mockCheckout).
			SetProductRepository(mockProduct).
			SetPromotionRepository(mockPromotion).
			SetInventoryRepository(mockInventory)
	}

	products := []entity.Product{
		{UUID: testMacBookUUID, SKU: "43N23P", Name: "MacBook Pro", PriceCents: 539999, InventoryQty: 5},
		{UUID: testRaspberryUUID, SKU: "234234", Name: "Raspberry Pi B", PriceCents: 3000, InventoryQty: 2},
		{UUID: testGoogleHomeUUID, SKU: "120P90", Name: "Google Home", PriceCents: 4999, InventoryQty: 10},
	}

	freeProductRule := entity.PromotionRule{
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

	tests := []struct {
		name    string
		req     CreateCheckoutRequest
		setup   func()
		wantErr bool
	}{
		{
			name: "invalid quantity returns error early",
			req: CreateCheckoutRequest{
				Items: []CheckoutItemRequest{
					{ProductUUID: testMacBookUUID.String(), Quantity: 0},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{
			name: "promotion repo error",
			req: CreateCheckoutRequest{
				Items: []CheckoutItemRequest{
					{ProductUUID: testMacBookUUID.String(), Quantity: 1},
				},
			},
			setup: func() {
				mockPromotion.EXPECT().
					GetActiveRulesByTargetProductUUIDs(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "transaction error on GetByUUIDsForUpdate",
			req: CreateCheckoutRequest{
				Items: []CheckoutItemRequest{
					{ProductUUID: testMacBookUUID.String(), Quantity: 1},
				},
			},
			setup: func() {
				mockPromotion.EXPECT().
					GetActiveRulesByTargetProductUUIDs(gomock.Any(), gomock.Any()).
					Return(nil, nil)
				mockTxManager.EXPECT().
					WithTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
						return fn(nil)
					})
				mockProduct.EXPECT().
					GetByUUIDsForUpdate(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("lock error"))
			},
			wantErr: true,
		},
		{
			name: "product not found in locked products",
			req: CreateCheckoutRequest{
				Items: []CheckoutItemRequest{
					{ProductUUID: testMacBookUUID.String(), Quantity: 1},
				},
			},
			setup: func() {
				mockPromotion.EXPECT().
					GetActiveRulesByTargetProductUUIDs(gomock.Any(), gomock.Any()).
					Return(nil, nil)
				mockTxManager.EXPECT().
					WithTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
						return fn(nil)
					})
				mockProduct.EXPECT().
					GetByUUIDsForUpdate(gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]entity.Product{}, nil)
			},
			wantErr: true,
		},
		{
			name: "insufficient stock in locked products",
			req: CreateCheckoutRequest{
				Items: []CheckoutItemRequest{
					{ProductUUID: testMacBookUUID.String(), Quantity: 10},
				},
			},
			setup: func() {
				mockPromotion.EXPECT().
					GetActiveRulesByTargetProductUUIDs(gomock.Any(), gomock.Any()).
					Return(nil, nil)
				mockTxManager.EXPECT().
					WithTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
						return fn(nil)
					})
				mockProduct.EXPECT().
					GetByUUIDsForUpdate(gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]entity.Product{{UUID: testMacBookUUID, SKU: "43N23P", Name: "MacBook Pro", PriceCents: 539999, InventoryQty: 2}}, nil)
			},
			wantErr: true,
		},
		{
			name: "checkout repo Create error",
			req: CreateCheckoutRequest{
				Items: []CheckoutItemRequest{
					{ProductUUID: testMacBookUUID.String(), Quantity: 1},
				},
			},
			setup: func() {
				mockPromotion.EXPECT().
					GetActiveRulesByTargetProductUUIDs(gomock.Any(), gomock.Any()).
					Return(nil, nil)
				mockTxManager.EXPECT().
					WithTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
						return fn(nil)
					})
				mockProduct.EXPECT().
					GetByUUIDsForUpdate(gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]entity.Product{products[0]}, nil)
				mockCheckout.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("insert error"))
			},
			wantErr: true,
		},
		{
			name: "happy path - single item no promotions",
			req: CreateCheckoutRequest{
				Items: []CheckoutItemRequest{
					{ProductUUID: testMacBookUUID.String(), Quantity: 1},
				},
			},
			setup: func() {
				mockPromotion.EXPECT().
					GetActiveRulesByTargetProductUUIDs(gomock.Any(), gomock.Any()).
					Return(nil, nil)
				mockTxManager.EXPECT().
					WithTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
						return fn(nil)
					})
				mockProduct.EXPECT().
					GetByUUIDsForUpdate(gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]entity.Product{products[0]}, nil)
				mockCheckout.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				mockCheckout.EXPECT().
					CreateItem(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				mockProduct.EXPECT().
					DeductStock(gomock.Any(), gomock.Any(), testMacBookUUID, 1).
					Return(nil)
				mockInventory.EXPECT().
					CreateMovement(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "happy path - multiple items with promotion",
			req: CreateCheckoutRequest{
				Items: []CheckoutItemRequest{
					{ProductUUID: testMacBookUUID.String(), Quantity: 1},
					{ProductUUID: testRaspberryUUID.String(), Quantity: 1},
				},
			},
			setup: func() {
				mockPromotion.EXPECT().
					GetActiveRulesByTargetProductUUIDs(gomock.Any(), gomock.Any()).
					Return([]entity.PromotionRule{freeProductRule}, nil)
				mockTxManager.EXPECT().
					WithTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
						return fn(nil)
					})
				mockProduct.EXPECT().
					GetByUUIDsForUpdate(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(products[:2], nil)
				mockCheckout.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				mockCheckout.EXPECT().
					CreateItem(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil).Times(2)
				mockCheckout.EXPECT().
					CreatePromotion(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				mockProduct.EXPECT().
					DeductStock(gomock.Any(), gomock.Any(), testMacBookUUID, 1).
					Return(nil)
				mockProduct.EXPECT().
					DeductStock(gomock.Any(), gomock.Any(), testRaspberryUUID, 1).
					Return(nil)
				mockInventory.EXPECT().
					CreateMovement(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil).Times(2)
			},
			wantErr: false,
		},
		{
			name: "DeductStock error rolls back",
			req: CreateCheckoutRequest{
				Items: []CheckoutItemRequest{
					{ProductUUID: testMacBookUUID.String(), Quantity: 1},
				},
			},
			setup: func() {
				mockPromotion.EXPECT().
					GetActiveRulesByTargetProductUUIDs(gomock.Any(), gomock.Any()).
					Return(nil, nil)
				mockTxManager.EXPECT().
					WithTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
						return fn(nil)
					})
				mockProduct.EXPECT().
					GetByUUIDsForUpdate(gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]entity.Product{products[0]}, nil)
				mockCheckout.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				mockCheckout.EXPECT().
					CreateItem(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				mockProduct.EXPECT().
					DeductStock(gomock.Any(), gomock.Any(), testMacBookUUID, 1).
					Return(errors.New("deduct error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			u := newUsecase()
			resp, err := u.Create(context.Background(), tt.req)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, resp.CheckoutID)
			assert.Equal(t, statusCompleted, resp.Status)
			assert.NotNil(t, resp.Items)
			assert.NotNil(t, resp.AppliedPromotions)
		})
	}
}
