package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/GSA-TTS/jemison/config"
	"github.com/GSA-TTS/jemison/internal/common"
	"github.com/GSA-TTS/jemison/internal/env"
	"github.com/GSA-TTS/jemison/internal/postgres"
	"github.com/GSA-TTS/jemison/internal/postgres/work_db"
	"github.com/GSA-TTS/jemison/internal/queueing"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

var ThisServiceName = "entree"

var ChQSHP = make(chan queueing.QSHP)

var JDB *postgres.JemisonDB

func main() {
	env.InitGlobalEnv(ThisServiceName)
	InitializeQueues()

	engine := common.InitializeAPI()

	JDB = postgres.NewJemisonDB()

	log.Println("environment initialized")

	// We need to load hosts into the `hosts` table for the entree logic to work.
	// Use the allowed hosts.
	for _, fqdn := range config.GetListOfHosts(env.Env.AllowedHosts) {
		d64, err := config.FQDNToDomain64(fqdn)
		if err != nil {
			zap.L().Error("could not get Domain64 for FQDN", zap.String("fqdn", fqdn))
		} else {
			zap.L().Debug("inserting fqdn/d64 to hosts", zap.String("fqdn", fqdn), zap.Int64("d64", d64))
			_, err := JDB.WorkDBQueries.UpsertUniqueHost(context.Background(),
				work_db.UpsertUniqueHostParams{
					Domain64: pgtype.Int8{
						Valid: true,
						Int64: d64,
					},
					NextFetch: pgtype.Timestamp{
						Valid:            true,
						InfinityModifier: 0,
						Time:             time.Now().Add(30 * 24 * time.Hour),
					},
				})

			if err != nil {
				zap.L().Error("error upserting domain64 value", zap.Int64("domain64", d64))
			}
		}
	}

	/*
		// FIXME: This would pre-load the crontab with
		// values based on the Domain64 JSON.
		//crontab(env.Env.AllowedHosts)
	*/

	go queueing.Enqueue(ChQSHP)

	zap.L().Info("listening to the music of the spheres",
		zap.String("port", env.Env.Port))
	// Local and Cloud should both get this from the environment.
	err := http.ListenAndServe(":"+env.Env.Port, engine)
	if err != nil {
		zap.Error(err)
	}
}
