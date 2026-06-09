package repository

import (
	"context"

	"github.com/crisyantoparulian/checkout-service/internal/pkg/helper"
	"github.com/insaneadinesia/gobang/gotel"
	"github.com/jmoiron/sqlx"
)

type HealthCheck interface {
	PingDB(ctx context.Context) (err error)
}

type heatlhCheck struct {
	db *sqlx.DB
}

func NewHealthCheckRepository(db *sqlx.DB) HealthCheck {
	if db == nil {
		panic("database is nil")
	}

	return &heatlhCheck{
		db: db,
	}
}

func (r *heatlhCheck) PingDB(ctx context.Context) (err error) {
	ctx, span := gotel.DefaultTracer().Start(ctx, helper.GetFuncName())
	defer span.End()

	err = r.db.PingContext(ctx)
	return
}
