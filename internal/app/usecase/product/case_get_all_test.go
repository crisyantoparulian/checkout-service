package product

import (
	"context"
	"errors"
	"testing"

	"github.com/crisyantoparulian/checkout-service/internal/app/entity"
	"github.com/crisyantoparulian/checkout-service/internal/app/repository"
	"github.com/crisyantoparulian/checkout-service/internal/app/repository/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// --- GetAll ---

func TestGetAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProduct := mocks.NewMockProduct(ctrl)

	testProducts := []entity.Product{
		{UUID: uuid.New(), SKU: "43N23P", Name: "MacBook Pro", PriceCents: 539999, InventoryQty: 5},
		{UUID: uuid.New(), SKU: "120P90", Name: "Google Home", PriceCents: 4999, InventoryQty: 10},
	}

	tests := []struct {
		name          string
		req           GetAllProductRequest
		setup         func()
		wantErr       bool
		wantCount     int
		wantPage      int
		wantPerPage   int
		wantTotal     int64
		wantPageCount int
	}{
		{
			name: "default page and perPage when zero",
			req:  GetAllProductRequest{},
			setup: func() {
				mockProduct.EXPECT().
					GetAll(gomock.Any(), repository.ProductFilter{Limit: 20, Offset: 0}).
					Return(nil, nil)
				mockProduct.EXPECT().
					CountTotal(gomock.Any()).
					Return(int64(0), nil)
			},
			wantCount:     0,
			wantPage:      1,
			wantPerPage:   20,
			wantTotal:     0,
			wantPageCount: 0,
		},
		{
			name: "GetAll repo error",
			req:  GetAllProductRequest{Page: 1, PerPage: 10},
			setup: func() {
				mockProduct.EXPECT().
					GetAll(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "CountTotal repo error",
			req:  GetAllProductRequest{Page: 1, PerPage: 10},
			setup: func() {
				mockProduct.EXPECT().
					GetAll(gomock.Any(), gomock.Any()).
					Return(nil, nil)
				mockProduct.EXPECT().
					CountTotal(gomock.Any()).
					Return(int64(0), errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "happy path with products",
			req:  GetAllProductRequest{Page: 1, PerPage: 10},
			setup: func() {
				mockProduct.EXPECT().
					GetAll(gomock.Any(), repository.ProductFilter{Limit: 10, Offset: 0}).
					Return(testProducts, nil)
				mockProduct.EXPECT().
					CountTotal(gomock.Any()).
					Return(int64(2), nil)
			},
			wantCount:     2,
			wantPage:      1,
			wantPerPage:   10,
			wantTotal:     2,
			wantPageCount: 1,
		},
		{
			name: "happy path with empty products",
			req:  GetAllProductRequest{Page: 2, PerPage: 10},
			setup: func() {
				mockProduct.EXPECT().
					GetAll(gomock.Any(), repository.ProductFilter{Limit: 10, Offset: 10}).
					Return(nil, nil)
				mockProduct.EXPECT().
					CountTotal(gomock.Any()).
					Return(int64(5), nil)
			},
			wantCount:     0,
			wantPage:      2,
			wantPerPage:   10,
			wantTotal:     5,
			wantPageCount: 1,
		},
		{
			name: "pagination with multiple pages",
			req:  GetAllProductRequest{Page: 1, PerPage: 2},
			setup: func() {
				mockProduct.EXPECT().
					GetAll(gomock.Any(), repository.ProductFilter{Limit: 2, Offset: 0}).
					Return(testProducts, nil)
				mockProduct.EXPECT().
					CountTotal(gomock.Any()).
					Return(int64(15), nil)
			},
			wantCount:     2,
			wantPage:      1,
			wantPerPage:   2,
			wantTotal:     15,
			wantPageCount: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			u := NewUsecase().SetProductRepository(mockProduct)
			resp, err := u.GetAll(context.Background(), tt.req)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, resp.Products, tt.wantCount)
			assert.Equal(t, tt.wantPage, resp.Pagination.Page)
			assert.Equal(t, tt.wantPerPage, resp.Pagination.PerPage)
			assert.Equal(t, tt.wantTotal, resp.Pagination.TotalCount)
			assert.Equal(t, tt.wantPageCount, resp.Pagination.PageCount)

			for _, p := range resp.Products {
				assert.NotEmpty(t, p.UUID)
				assert.NotEmpty(t, p.SKU)
				assert.NotEmpty(t, p.Name)
				assert.NotEmpty(t, p.FormattedPrice)
			}
		})
	}
}

// --- formatMoney ---

func TestFormatMoney(t *testing.T) {
	tests := []struct {
		name  string
		cents int64
		want  string
	}{
		{name: "zero", cents: 0, want: "$0.00"},
		{name: "single digit cents", cents: 1050, want: "$10.50"},
		{name: "two digit cents", cents: 4999, want: "$49.99"},
		{name: "one cent", cents: 1, want: "$0.01"},
		{name: "large value", cents: 539999, want: "$5399.99"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, formatMoney(tt.cents))
		})
	}
}
