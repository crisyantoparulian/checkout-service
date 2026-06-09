package worker

import (
	"context"
	"encoding/json"

	"github.com/crisyantoparulian/checkout-service/internal/app/container"
	"github.com/crisyantoparulian/checkout-service/internal/pkg/worker"
	"github.com/hibiken/asynq"
	"github.com/insaneadinesia/gobang/logger"
)

func SetupHandler(container *container.Container) *asynq.ServeMux {
	mux := asynq.NewServeMux()

	mux.Use(TracingMiddleware)
	mux.Use(SetLoggingMiddleware(container))

	return mux
}

// Execute will extract the original payload and execute the function
func Execute(f func(ctx context.Context, payload any) (err error)) asynq.HandlerFunc {
	return func(ctx context.Context, task *asynq.Task) (err error) {
		defer func() {
			if err != nil {
				logger.Log.Error(ctx, "Execute Error", err.Error())
			}

			logger.Log.TDR(ctx)
		}()

		var payload worker.TaskPayload
		if err = json.Unmarshal(task.Payload(), &payload); err != nil {
			return
		}

		err = f(ctx, payload.Data)

		return
	}
}
