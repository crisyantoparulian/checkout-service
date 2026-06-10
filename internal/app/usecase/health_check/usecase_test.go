package health_check

import (
	"testing"

	"github.com/crisyantoparulian/checkout-service/internal/app/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewUsecase(t *testing.T) {
	u := NewUsecase()
	assert.NotNil(t, u)
}

func TestValidate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHealthCheck := mocks.NewMockHealthCheck(ctrl)

	result := NewUsecase().SetHealthCheckRepository(mockHealthCheck).Validate()
	assert.NotNil(t, result)
}

func TestValidate_Panics(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *usecase
		panicMsg string
	}{
		{
			name:     "nil healthCheckRepository",
			setup:    func() *usecase { return NewUsecase() },
			panicMsg: "healthCheckRepository is nil",
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
