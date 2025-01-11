package main

import (
	"log"
	"os"

	common "github.com/GSA-TTS/jemison/internal/common"
	"github.com/GSA-TTS/jemison/internal/env"
	"github.com/GSA-TTS/jemison/internal/queueing"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"go.uber.org/zap"
)

// GLOBAL TO THE APP
// One pool of connections for River.

// The work client, doing the work of `fetch`.
var servePool *pgxpool.Pool

var serveClient *river.Client[pgx.Tx]

type ServeWorker struct {
	river.WorkerDefaults[common.ServeArgs]
}

func InitializeQueues() {
	queueing.InitializeRiverQueues()

	ctx, sP, workers := common.CommonQueueInit()
	servePool = sP

	// Essentially adds a worker "type" to the work engine.
	river.AddWorker(workers, &ServeWorker{})

	// Grab the number of workers from the config.
	serveService, err := env.Env.GetUserService("serve")
	if err != nil {
		zap.L().Error("could not fetch service config")
		log.Println(err)
		os.Exit(1)
	}

	// Work client
	serveClient, err = river.NewClient(riverpgxv5.New(servePool), &river.Config{
		Queues: map[string]river.QueueConfig{
			"serve": {MaxWorkers: int(serveService.GetParamInt64("workers"))},
		},
		Workers: workers,
	})
	if err != nil {
		zap.L().Error("could not establish worker pool")
		log.Println(err)
		os.Exit(1)
	}

	// Start the work clients
	if err := serveClient.Start(ctx); err != nil {
		zap.L().Error("workers are not the means of production. exiting.")
		os.Exit(1)
	}
}
