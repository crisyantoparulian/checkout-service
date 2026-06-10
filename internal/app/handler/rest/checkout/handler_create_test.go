package checkout

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	checkoutusecase "github.com/crisyantoparulian/checkout-service/internal/app/usecase/checkout"
	checkoutmocks "github.com/crisyantoparulian/checkout-service/internal/app/usecase/checkout/mocks"
	pkgvalidator "github.com/crisyantoparulian/checkout-service/internal/pkg/validator"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type echoValidator struct {
	validator *validator.Validate
}

func (v *echoValidator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setup      func(*checkoutmocks.MockCheckoutUsecase)
		wantErr    bool
		wantStatus int
	}{
		{
			name: "success",
			body: `{"items":[{"product_uuid":"` + uuid.New().String() + `","quantity":1}]}`,
			setup: func(mockCheckout *checkoutmocks.MockCheckoutUsecase) {
				mockCheckout.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(checkoutusecase.CheckoutResponse{CheckoutID: uuid.New().String(), Status: "COMPLETED"}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "invalid body returns error",
			body: `{invalid-json}`,
			setup: func(mockCheckout *checkoutmocks.MockCheckoutUsecase) {
			},
			wantErr: true,
		},
		{
			name: "validation error returns error",
			body: `{"items":[]}`,
			setup: func(mockCheckout *checkoutmocks.MockCheckoutUsecase) {
			},
			wantErr: true,
		},
		{
			name: "usecase error returns error",
			body: `{"items":[{"product_uuid":"` + uuid.New().String() + `","quantity":1}]}`,
			setup: func(mockCheckout *checkoutmocks.MockCheckoutUsecase) {
				mockCheckout.EXPECT().
					Create(gomock.Any(), gomock.Any()).
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
			e.Validator = &echoValidator{validator: pkgvalidator.SetupValidator()}
			req := httptest.NewRequest(http.MethodPost, "/checkouts", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			h := NewHandler().SetCheckoutUsecase(mockCheckout)
			err := h.Create(c)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), "Request Successfully Processed")
		})
	}
}
