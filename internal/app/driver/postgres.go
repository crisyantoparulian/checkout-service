package driver

import (
	"fmt"

	"github.com/crisyantoparulian/checkout-service/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewPostgresDatabase(cfg config.Config) (db *sqlx.DB, err error) {
	fmt.Println("Try New Database ...")

	dsn := cfg.GetPostgresDSN()
	db, err = sqlx.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	db.SetMaxIdleConns(cfg.DBMaxIdleConn)
	db.SetMaxOpenConns(cfg.DBMaxOpenConn)
	db.SetConnMaxLifetime(cfg.DBMaxConnLifetime)

	if err = db.Ping(); err != nil {
		panic(err)
	}

	return
}
