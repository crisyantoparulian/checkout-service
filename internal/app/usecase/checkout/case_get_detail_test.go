package checkout

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/crisyantoparulian/checkout-service/internal/app/entity"
	"github.com/crisyantoparulian/checkout-service/internal/app/repository/mocks"
	"github.com/crisyantoparulian/checkout-service/internal/pkg/apperror"
	"github.com/crisyantoparulian/checkout-service/internal/pkg/constants"
	"github.com/crisyantoparulian/checkout-service/internal/pkg/transaction"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDetail(t *testing.T) {
	checkoutUUID := uuid.New()

	tests := []struct {
		name           string
		checkoutID     string
		setupMocks     func(checkoutRepo *mocks.MockCheckout)
		wantErr        bool
		wantErrCode    string
		wantStatus     string
		wantItemCount  int
		wantPromoCount int
	}{
		{
			name:        "invalid UUID",
			checkoutID:  "not-a-uuid",
			setupMocks:  func(_ *mocks.MockCheckout) {},
			wantErr:     true,
			wantErrCode: "",
		},
		{
			name:       "checkout not found",
			checkoutID: checkoutUUID.String(),
			setupMocks: func(cr *mocks.MockCheckout) {
				cr.EXPECT().GetByUUID(gomock.Any(), checkoutUUID).Return(entity.Checkout{}, sql.ErrNoRows)
			},
			wantErr:     true,
			wantErrCode: constants.CODE_CHECKOUT_NOT_FOUND,
		},
		{
			name:       "GetByUUID generic error",
			checkoutID: checkoutUUID.String(),
			setupMocks: func(cr *mocks.MockCheckout) {
				cr.EXPECT().GetByUUID(gomock.Any(), checkoutUUID).Return(entity.Checkout{}, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name:       "GetItems error",
			checkoutID: checkoutUUID.String(),
			setupMocks: func(cr *mocks.MockCheckout) {
				cr.EXPECT().GetByUUID(gomock.Any(), checkoutUUID).Return(entity.Checkout{UUID: checkoutUUID, Status: "COMPLETED"}, nil)
				cr.EXPECT().GetItems(gomock.Any(), checkoutUUID).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name:       "GetPromotions error",
			checkoutID: checkoutUUID.String(),
			setupMocks: func(cr *mocks.MockCheckout) {
				cr.EXPECT().GetByUUID(gomock.Any(), checkoutUUID).Return(entity.Checkout{UUID: checkoutUUID, Status: "COMPLETED"}, nil)
				cr.EXPECT().GetItems(gomock.Any(), checkoutUUID).Return([]entity.CheckoutItem{}, nil)
				cr.EXPECT().GetPromotions(gomock.Any(), checkoutUUID).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name:       "happy path with items and promotions",
			checkoutID: checkoutUUID.String(),
			setupMocks: func(cr *mocks.MockCheckout) {
				cr.EXPECT().GetByUUID(gomock.Any(), checkoutUUID).Return(entity.Checkout{
					UUID:          checkoutUUID,
					Status:        "COMPLETED",
					SubtotalCents: 542999,
					DiscountCents: 3000,
					TotalCents:    539999,
					CreatedAt:     time.Now(),
				}, nil)
				cr.EXPECT().GetItems(gomock.Any(), checkoutUUID).Return([]entity.CheckoutItem{
					{ProductUUID: testMacBookUUID, SKU: "43N23P", ProductName: "MacBook Pro", Quantity: 1, UnitPriceCents: 539999, TotalPriceCents: 539999},
				}, nil)
				cr.EXPECT().GetPromotions(gomock.Any(), checkoutUUID).Return([]entity.CheckoutPromotion{
					{PromotionCode: "TEST", PromotionName: "Test Promo", PromotionType: "FREE_PRODUCT", DiscountCents: 3000},
				}, nil)
			},
			wantErr:        false,
			wantStatus:     "COMPLETED",
			wantItemCount:  1,
			wantPromoCount: 1,
		},
		{
			name:       "happy path with no promotions",
			checkoutID: checkoutUUID.String(),
			setupMocks: func(cr *mocks.MockCheckout) {
				cr.EXPECT().GetByUUID(gomock.Any(), checkoutUUID).Return(entity.Checkout{
					UUID:          checkoutUUID,
					Status:        "COMPLETED",
					SubtotalCents: 539999,
					TotalCents:    539999,
					CreatedAt:     time.Now(),
				}, nil)
				cr.EXPECT().GetItems(gomock.Any(), checkoutUUID).Return([]entity.CheckoutItem{
					{ProductUUID: testMacBookUUID, SKU: "43N23P", ProductName: "MacBook Pro", Quantity: 1, UnitPriceCents: 539999, TotalPriceCents: 539999},
				}, nil)
				cr.EXPECT().GetPromotions(gomock.Any(), checkoutUUID).Return([]entity.CheckoutPromotion{}, nil)
			},
			wantErr:        false,
			wantStatus:     "COMPLETED",
			wantItemCount:  1,
			wantPromoCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockCheckout := mocks.NewMockCheckout(ctrl)
			tt.setupMocks(mockCheckout)

			u := NewUsecase().
				SetTxManager(&transaction.Manager{}).
				SetCheckoutRepository(mockCheckout).
				SetProductRepository(mocks.NewMockProduct(ctrl)).
				SetPromotionRepository(mocks.NewMockPromotion(ctrl)).
				SetInventoryRepository(mocks.NewMockInventory(ctrl))

			resp, err := u.GetDetail(context.Background(), tt.checkoutID)
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrCode != "" {
					appErr, ok := err.(*apperror.ApplicationError)
					require.True(t, ok)
					assert.Equal(t, tt.wantErrCode, appErr.ErrorCode)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.Status)
			assert.Len(t, resp.Items, tt.wantItemCount)
			assert.Len(t, resp.AppliedPromotions, tt.wantPromoCount)
			assert.NotEmpty(t, resp.CreatedAt)
		})
	}
}
