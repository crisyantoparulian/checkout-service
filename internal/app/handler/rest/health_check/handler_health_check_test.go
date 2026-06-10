package health_check

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	healthusecase "github.com/crisyantoparulian/checkout-service/internal/app/usecase/health_check"
	healthmocks "github.com/crisyantoparulian/checkout-service/internal/app/usecase/health_check/mocks"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestHealthCheck(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*healthmocks.MockHealthCheckUsecase)
		wantStatus int
		wantBody   string
	}{
		{
			name: "success returns OK status",
			setup: func(mockHealthCheck *healthmocks.MockHealthCheckUsecase) {
				mockHealthCheck.EXPECT().
					HealthCheck(gomock.Any()).
					Return(healthusecase.StatusCheck{DBStatus: "OK"}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   "OK",
		},
		{
			name: "usecase error returns ERROR status",
			setup: func(mockHealthCheck *healthmocks.MockHealthCheckUsecase) {
				mockHealthCheck.EXPECT().
					HealthCheck(gomock.Any()).
					Return(healthusecase.StatusCheck{DBStatus: "ERROR"}, errors.New("db error"))
			},
			wantStatus: http.StatusOK,
			wantBody:   "ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHealthCheck := healthmocks.NewMockHealthCheckUsecase(ctrl)
			tt.setup(mockHealthCheck)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			h := NewHandler().SetHealthCheckUsecase(mockHealthCheck)
			err := h.HealthCheck(c)

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.wantBody)
		})
	}
}
