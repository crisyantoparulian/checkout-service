package health_check

import (
	"testing"

	healthmocks "github.com/crisyantoparulian/checkout-service/internal/app/usecase/health_check/mocks"
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

	mockHealthCheck := healthmocks.NewMockHealthCheckUsecase(ctrl)

	result := NewHandler().SetHealthCheckUsecase(mockHealthCheck).Validate()
	assert.NotNil(t, result)
}

func TestValidate_Panics(t *testing.T) {
	require.PanicsWithValue(t, "healthCheckUsecase is nil", func() {
		NewHandler().Validate()
	})
}
