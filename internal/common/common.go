package common

import (
	"context"
	"os"

	"github.com/GSA-TTS/jemison/internal/env"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

func GetPool(databaseURL string) (context.Context, *pgxpool.Pool) {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		zap.L().Error("could not establish database pool; exiting",
			zap.String("database_url", databaseURL),
		)
		os.Exit(1)
	}

	return ctx, pool
}

//nolint:revive
func CommonQueueInit() (context.Context, *pgxpool.Pool, *river.Workers) {
	var err error

	databaseURL, err := env.Env.GetDatabaseURL(env.QueueDatabase)
	if err != nil {
		zap.L().Error("unable to get connection string; exiting",
			zap.String("database", env.QueueDatabase),
		)
		os.Exit(1)
	}

	// Establsih the database
	ctx, pool := GetPool(databaseURL)
	// Create a pool of workers
	workers := river.NewWorkers()

	return ctx, pool, workers
}

func CtxTx(pool *pgxpool.Pool) (context.Context, pgx.Tx) {
	ctx := context.Background()

	tx, err := pool.Begin(ctx)
	if err != nil {
		zap.L().Panic("cannot init tx from pool")
	}

	return ctx, tx
}
