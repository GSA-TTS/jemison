package main

import (
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jadudm/eight/internal/common"
	"github.com/jadudm/eight/internal/env"
	"github.com/jadudm/eight/internal/queueing"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"go.uber.org/zap"
)

// The work client, doing the work of `fetch`
var dbPool *pgxpool.Pool
var walkClient *river.Client[pgx.Tx]
var fetchClient *river.Client[pgx.Tx]

type WalkWorker struct {
	river.WorkerDefaults[common.WalkArgs]
}

func InitializeQueues() {
	queueing.InitializeRiverQueues()

	ctx, pool, workers := common.CommonQueueInit()
	dbPool = pool

	// Essentially adds a worker "type" to the work engine.
	river.AddWorker(workers, &WalkWorker{})

	// Grab the number of workers from the config.
	walk_service, err := env.Env.GetUserService("walk")
	if err != nil {
		zap.L().Error("could not fetch service config")
		log.Println(err)
		os.Exit(1)
	}

	// Work client
	walkClient, err = river.NewClient(riverpgxv5.New(dbPool), &river.Config{
		Queues: map[string]river.QueueConfig{
			"walk": {MaxWorkers: int(walk_service.GetParamInt64("workers"))},
		},
		Workers: workers,
	})

	if err != nil {
		zap.L().Error("could not establish worker pool")
		log.Println(err)
		os.Exit(1)
	}

	// Insert-only client to `fetch`
	fetchClient, err = river.NewClient(riverpgxv5.New(dbPool), &river.Config{})
	if err != nil {
		zap.L().Error("could not establish insert-only client")
		os.Exit(1)
	}

	// Start the work clients
	if err := walkClient.Start(ctx); err != nil {
		zap.L().Error("workers are not the means of production. exiting.")
		os.Exit(42)
	}
}
