package product

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	productusecase "github.com/crisyantoparulian/checkout-service/internal/app/usecase/product"
	productmocks "github.com/crisyantoparulian/checkout-service/internal/app/usecase/product/mocks"
	pkgvalidator "github.com/crisyantoparulian/checkout-service/internal/pkg/validator"
	"github.com/go-playground/validator/v10"
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

func TestGetAll(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		setup      func(*productmocks.MockProductUsecase)
		wantErr    bool
		wantStatus int
	}{
		{
			name:   "success with query params",
			target: "/products?page=2&per_page=10",
			setup: func(mockProduct *productmocks.MockProductUsecase) {
				mockProduct.EXPECT().
					GetAll(gomock.Any(), productusecase.GetAllProductRequest{Page: 2, PerPage: 10}).
					Return(productusecase.GetAllProductResponse{Products: []productusecase.ProductResponse{}}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "success with default query params",
			target: "/products",
			setup: func(mockProduct *productmocks.MockProductUsecase) {
				mockProduct.EXPECT().
					GetAll(gomock.Any(), productusecase.GetAllProductRequest{}).
					Return(productusecase.GetAllProductResponse{Products: []productusecase.ProductResponse{}}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "validation error returns error",
			target: "/products?page=0&per_page=101",
			setup: func(mockProduct *productmocks.MockProductUsecase) {
			},
			wantErr: true,
		},
		{
			name:   "usecase error returns error",
			target: "/products?page=1&per_page=20",
			setup: func(mockProduct *productmocks.MockProductUsecase) {
				mockProduct.EXPECT().
					GetAll(gomock.Any(), productusecase.GetAllProductRequest{Page: 1, PerPage: 20}).
					Return(productusecase.GetAllProductResponse{}, errors.New("usecase error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProduct := productmocks.NewMockProductUsecase(ctrl)
			tt.setup(mockProduct)

			e := echo.New()
			e.Validator = &echoValidator{validator: pkgvalidator.SetupValidator()}
			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			h := NewHandler().SetProductUsecase(mockProduct)
			err := h.GetAll(c)

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
