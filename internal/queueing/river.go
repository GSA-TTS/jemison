package queueing

import (
	"context"
	"log"

	"github.com/GSA-TTS/jemison/internal/env"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"go.uber.org/zap"
)

func InitializeRiverQueues() {
	// Set up a pool
	connectionString, err := env.Env.GetDatabaseURL(env.QueueDatabase)
	if err != nil {
		zap.L().Fatal("cannot find db connection string",
			zap.String("database", env.QueueDatabase))
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		zap.L().Fatal("cannot create database pool for migrations")
	}

	defer pool.Close()

	// Run the migrations, always.
	migrator, err := rivermigrate.New(riverpgxv5.New(pool), nil)
	if err != nil {
		zap.L().Info("could not create a river migrator")
	}

	_, err = migrator.Migrate(ctx, rivermigrate.DirectionUp, &rivermigrate.MigrateOpts{})
	if err != nil {
		zap.L().Info("could not run the river migrator")
	}
}

func RunRiverMigrator() {
	ctx := context.Background()
	// Set up a pool
	connectionString, err := env.Env.GetDatabaseURL(env.QueueDatabase)
	if err != nil {
		log.Println("RIVER cannot find connection string for", env.QueueDatabase)
		log.Fatal(err)
	}

	pool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		zap.L().Fatal("could not get pool for river migrator")
	}
	defer pool.Close()

	// Run the migrations, always.
	migrator, err := rivermigrate.New(riverpgxv5.New(pool), nil)
	if err != nil {
		zap.L().Error("river could not create river migrator. exiting.")
		zap.L().Fatal(err.Error())
	}

	_, err = migrator.Migrate(ctx, rivermigrate.DirectionUp, &rivermigrate.MigrateOpts{})
	if err != nil {
		zap.L().Error("river could not run river migrations. exiting.")
		zap.L().Fatal(err.Error())
	}
}
