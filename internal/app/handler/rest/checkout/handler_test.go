package checkout

import (
	"testing"

	checkoutmocks "github.com/crisyantoparulian/checkout-service/internal/app/usecase/checkout/mocks"
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

	mockCheckout := checkoutmocks.NewMockCheckoutUsecase(ctrl)

	result := NewHandler().SetCheckoutUsecase(mockCheckout).Validate()
	assert.NotNil(t, result)
}

func TestValidate_Panics(t *testing.T) {
	require.PanicsWithValue(t, "checkoutUsecase is nil", func() {
		NewHandler().Validate()
	})
}
