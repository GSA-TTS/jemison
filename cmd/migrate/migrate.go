package main

import (
	"embed"
	"fmt"
	"log"
	"net/url"

	"github.com/GSA-TTS/jemison/internal/env"
	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	"go.uber.org/zap"
)

// Carry our migrations with us as part of the build.
// This eliminates wondering where they are when we deploy.

//go:embed work_db/db/migrations/*.sql
var workFS embed.FS

//go:embed search_db/db/migrations/*.sql
var searchFS embed.FS

type location struct {
	FS            embed.FS
	MigrationsDir string
}

// Assumes config has been read.
func MigrateDB(dbUri string, loc location) {
	db1_url, err := env.Env.GetDatabaseUrl(dbUri)
	if err != nil {
		zap.L().Fatal("could not get url for",
			zap.String("URI", dbUri))
	}

	u, _ := url.Parse(db1_url)
	db := dbmate.New(u)
	db.FS = loc.FS
	db.MigrationsDir = []string{loc.MigrationsDir}

	log.Println("Migrations:")

	migrations, err := db.FindMigrations()
	if err != nil {
		panic(err)
	}

	for _, m := range migrations {
		fmt.Println(m.Version, m.FilePath)
	}

	log.Println("\nApplying...")

	err = db.CreateAndMigrate()
	if err != nil {
		panic(err)
	}
}

func MigrateJemisonDB() {
	dbs := make(map[string]location)

	dbs[env.JemisonWorkDatabase] = location{
		FS:            workFS,
		MigrationsDir: "work_db/db/migrations",
	}

	dbs[env.SearchDatabase] = location{
		FS:            searchFS,
		MigrationsDir: "search_db/db/migrations",
	}

	for k, v := range dbs {
		MigrateDB(k, v)
	}
}
