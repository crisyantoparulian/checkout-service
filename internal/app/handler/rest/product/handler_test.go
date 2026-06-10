package product

import (
	"testing"

	productmocks "github.com/crisyantoparulian/checkout-service/internal/app/usecase/product/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewHandler(t *testing.T) {
	h := NewHandler()
	assert.NotNil(t, h)
}

func TestValidate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProduct := productmocks.NewMockProductUsecase(ctrl)

	result := NewHandler().SetProductUsecase(mockProduct).Validate()
	assert.NotNil(t, result)
}

func TestValidate_Panics(t *testing.T) {
	require.PanicsWithValue(t, "productUsecase is nil", func() {
		NewHandler().Validate()
	})
}
