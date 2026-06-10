package checkout

import (
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/crisyantoparulian/checkout-service/internal/app/repository/mocks"
	txmocks "github.com/crisyantoparulian/checkout-service/internal/pkg/transaction/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUsecase(t *testing.T) {
	u := NewUsecase()
	assert.NotNil(t, u)
}

func TestValidate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTxManager := txmocks.NewMockManagerInterface(ctrl)
	mockCheckout := mocks.NewMockCheckout(ctrl)
	mockProduct := mocks.NewMockProduct(ctrl)
	mockPromotion := mocks.NewMockPromotion(ctrl)
	mockInventory := mocks.NewMockInventory(ctrl)

	u := NewUsecase().
		SetTxManager(mockTxManager).
		SetCheckoutRepository(mockCheckout).
		SetProductRepository(mockProduct).
		SetPromotionRepository(mockPromotion).
		SetInventoryRepository(mockInventory)

	result := u.Validate()
	assert.NotNil(t, result)
}

func TestValidate_Panics(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockTxManager := txmocks.NewMockManagerInterface(ctrl)

	tests := []struct {
		name     string
		setup    func() *usecase
		panicMsg string
	}{
		{
			name:     "nil txManager",
			setup:    func() *usecase { return NewUsecase() },
			panicMsg: "txManager is nil",
		},
		{
			name: "nil checkoutRepository",
			setup: func() *usecase {
				return NewUsecase().SetTxManager(mockTxManager)
			},
			panicMsg: "checkoutRepository is nil",
		},
		{
			name: "nil productRepository",
			setup: func() *usecase {
				return NewUsecase().
					SetTxManager(mockTxManager).
					SetCheckoutRepository(mocks.NewMockCheckout(ctrl))
			},
			panicMsg: "productRepository is nil",
		},
		{
			name: "nil promotionRepository",
			setup: func() *usecase {
				return NewUsecase().
					SetTxManager(mockTxManager).
					SetCheckoutRepository(mocks.NewMockCheckout(ctrl)).
					SetProductRepository(mocks.NewMockProduct(ctrl))
			},
			panicMsg: "promotionRepository is nil",
		},
		{
			name: "nil inventoryRepository",
			setup: func() *usecase {
				return NewUsecase().
					SetTxManager(mockTxManager).
					SetCheckoutRepository(mocks.NewMockCheckout(ctrl)).
					SetProductRepository(mocks.NewMockProduct(ctrl)).
					SetPromotionRepository(mocks.NewMockPromotion(ctrl))
			},
			panicMsg: "inventoryRepository is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := tt.setup()
			require.PanicsWithValue(t, tt.panicMsg, func() {
				u.Validate()
			})
		})
	}
}
