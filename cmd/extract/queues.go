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

// GLOBAL TO THE APP
// One pool of connections for River.
// The work client, doing the work of `extract`

type ExtractWorker struct {
	river.WorkerDefaults[common.ExtractArgs]
}

func InitializeQueues() {
	var extractClient *river.Client[pgx.Tx]

	var extractPool *pgxpool.Pool

	queueing.InitializeRiverQueues()

	var err error

	ctx, extractPool, workers := common.CommonQueueInit()

	zap.L().Debug("initialized common queues")

	river.AddWorker(workers, &ExtractWorker{})

	// Grab the number of workers from the config.
	extractService, err := env.Env.GetUserService("extract")
	if err != nil {
		zap.L().Error("could not fetch service config")
		log.Println(err)
		os.Exit(1)
	}

	// Work client
	extractClient, err = river.NewClient(riverpgxv5.New(extractPool), &river.Config{
		Queues: map[string]river.QueueConfig{
			"extract": {MaxWorkers: int(extractService.GetParamInt64("workers"))},
		},
		Workers: workers,
	})
	if err != nil {
		zap.L().Error("could not establish worker pool")
		log.Println(err)
		os.Exit(1)
	}

	// Start the work clients
	if err := extractClient.Start(ctx); err != nil {
		zap.L().Error("workers are not the means of production. exiting.")
		os.Exit(1)
	}
}
