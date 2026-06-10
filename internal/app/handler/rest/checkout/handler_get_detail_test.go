package checkout

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	checkoutusecase "github.com/crisyantoparulian/checkout-service/internal/app/usecase/checkout"
	checkoutmocks "github.com/crisyantoparulian/checkout-service/internal/app/usecase/checkout/mocks"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetDetail(t *testing.T) {
	checkoutID := uuid.New().String()

	tests := []struct {
		name       string
		checkoutID string
		setup      func(*checkoutmocks.MockCheckoutUsecase)
		wantErr    bool
		wantStatus int
	}{
		{
			name:       "success",
			checkoutID: checkoutID,
			setup: func(mockCheckout *checkoutmocks.MockCheckoutUsecase) {
				mockCheckout.EXPECT().
					GetDetail(gomock.Any(), checkoutID).
					Return(checkoutusecase.CheckoutResponse{CheckoutID: checkoutID, Status: "COMPLETED"}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "usecase error returns error",
			checkoutID: checkoutID,
			setup: func(mockCheckout *checkoutmocks.MockCheckoutUsecase) {
				mockCheckout.EXPECT().
					GetDetail(gomock.Any(), checkoutID).
					Return(checkoutusecase.CheckoutResponse{}, errors.New("usecase error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockCheckout := checkoutmocks.NewMockCheckoutUsecase(ctrl)
			tt.setup(mockCheckout)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/checkouts/"+tt.checkoutID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("uuid")
			c.SetParamValues(tt.checkoutID)

			h := NewHandler().SetCheckoutUsecase(mockCheckout)
			err := h.GetDetail(c)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.checkoutID)
		})
	}
}
