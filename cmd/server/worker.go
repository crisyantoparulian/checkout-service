package server

import (
	"github.com/crisyantoparulian/checkout-service/internal/app/container"
	"github.com/crisyantoparulian/checkout-service/internal/app/server/worker"
	"github.com/spf13/cobra"
)

func NewWorkerServer() *cobra.Command {
	return &cobra.Command{
		Use:   "worker",
		Short: "Run Worker Server",
		Long:  "Run Worker Server",
		Run: func(cmd *cobra.Command, args []string) {
			container := container.Setup()
			worker.StartWorkerService(container)
		},
	}
}
