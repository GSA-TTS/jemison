package main

import (
	"log"
	"os"

	"github.com/GSA-TTS/jemison/internal/common"
	"github.com/GSA-TTS/jemison/internal/env"
	"github.com/GSA-TTS/jemison/internal/queueing"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"go.uber.org/zap"
)

// The work client, doing the work of `fetch`.
var dbPool *pgxpool.Pool

var walkClient *river.Client[pgx.Tx]

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
	walkService, err := env.Env.GetUserService("walk")
	if err != nil {
		zap.L().Error("could not fetch service config")
		log.Println(err)
		os.Exit(1)
	}

	// Work client
	walkClient, err = river.NewClient(riverpgxv5.New(dbPool), &river.Config{
		Queues: map[string]river.QueueConfig{
			"walk": {MaxWorkers: int(walkService.GetParamInt64("workers"))},
		},
		Workers: workers,
	})
	if err != nil {
		zap.L().Error("could not establish worker pool")
		log.Println(err)
		os.Exit(1)
	}

	// Start the work clients
	if err := walkClient.Start(ctx); err != nil {
		zap.L().Error("workers are not the means of production. exiting.")
		os.Exit(1)
	}
}
