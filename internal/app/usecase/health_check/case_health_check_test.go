package health_check

import (
	"context"
	"errors"
	"testing"

	"github.com/crisyantoparulian/checkout-service/internal/app/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestHealthCheck(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(*mocks.MockHealthCheck)
		wantDBStatus string
		wantErr      bool
	}{
		{
			name: "database ping success returns OK",
			setup: func(mockHealthCheck *mocks.MockHealthCheck) {
				mockHealthCheck.EXPECT().PingDB(gomock.Any()).Return(nil)
			},
			wantDBStatus: "OK",
		},
		{
			name: "database ping error returns ERROR",
			setup: func(mockHealthCheck *mocks.MockHealthCheck) {
				mockHealthCheck.EXPECT().PingDB(gomock.Any()).Return(errors.New("db error"))
			},
			wantDBStatus: "ERROR",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHealthCheck := mocks.NewMockHealthCheck(ctrl)
			tt.setup(mockHealthCheck)

			u := NewUsecase().SetHealthCheckRepository(mockHealthCheck)
			resp, err := u.HealthCheck(context.Background())

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantDBStatus, resp.DBStatus)
		})
	}
}
