package postgres

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/GSA-TTS/jemison/config"
	"github.com/GSA-TTS/jemison/internal/env"
	search_db "github.com/GSA-TTS/jemison/internal/postgres/search_db"
	work_db "github.com/GSA-TTS/jemison/internal/postgres/work_db"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type JemisonDB struct {
	Config          map[string]*pgxpool.Config
	Pool            map[string]*pgxpool.Pool
	WorkDBQueries   *work_db.Queries
	SearchDBQueries *search_db.Queries
	constCache      sync.Map
}

func NewJemisonDB() *JemisonDB {
	jdb := JemisonDB{
		Config: make(map[string]*pgxpool.Config),
		Pool:   make(map[string]*pgxpool.Pool),
	}

	for _, dbName := range []string{env.QueueDatabase, env.JemisonWorkDatabase, env.SearchDatabase} {
		dbString, err := env.Env.GetDatabaseURL(dbName)
		if err != nil {
			zap.L().Fatal("could not get db URL", zap.String("db_name", dbName))
		}

		cfg := Config(dbString)
		// Create database connection
		pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
		if err != nil {
			zap.L().Fatal("could not create pool",
				zap.String("err", err.Error()))
		}

		err = pool.Ping(context.Background())
		if err != nil {
			zap.L().Error(err.Error())
		}

		jdb.Config[dbName] = cfg
		jdb.Pool[dbName] = pool
	}

	jdb.WorkDBQueries = work_db.New(jdb.Pool[env.JemisonWorkDatabase])
	jdb.SearchDBQueries = search_db.New(jdb.Pool[env.SearchDatabase])

	return &jdb
}

func Config(dbString string) *pgxpool.Config {
	const defaultMaxConns = int32(100)

	const defaultMinConns = int32(0)

	const defaultMaxConnLifetime = time.Hour

	const defaultMaxConnIdleTime = time.Minute * 30

	const defaultHealthCheckPeriod = time.Minute

	const defaultConnectTimeout = time.Second * 5

	dbConfig, err := pgxpool.ParseConfig(dbString)
	if err != nil {
		log.Fatal("Failed to create a config, error: ", err)
	}

	dbConfig.MaxConns = defaultMaxConns
	dbConfig.MinConns = defaultMinConns
	dbConfig.MaxConnLifetime = defaultMaxConnLifetime
	dbConfig.MaxConnIdleTime = defaultMaxConnIdleTime
	dbConfig.HealthCheckPeriod = defaultHealthCheckPeriod
	dbConfig.ConnConfig.ConnectTimeout = defaultConnectTimeout

	return dbConfig
}

// The cache is a key/value store, so prepend
// keys to avoid collisions. It should be impossible,
// but still... that's the convention of these functions.
//
//nolint:gosec
func (jdb *JemisonDB) GetScheme(scheme string) int32 {
	if val, ok := jdb.constCache.Load("scheme:" + scheme); ok {
		v, assertOK := val.(int32)
		if !assertOK {
			zap.L().Error("could not convert scheme integer")
		}

		return v
	}

	schemeInt := config.GetScheme(scheme)
	// This is a guaranteed save conversion
	jdb.constCache.Store("scheme:"+scheme, int32(schemeInt))

	return int32(schemeInt)
}

func (jdb *JemisonDB) GetContentType(ct string) int {
	if val, ok := jdb.constCache.Load("contenttype:" + ct); ok {
		v, assertOK := val.(int)
		if !assertOK {
			zap.L().Error("could not convert content type integer")
		}

		return v
	}

	ctInt := config.GetContentType(ct)
	jdb.constCache.Store("contenttype:"+ct, ctInt)

	return ctInt
}

const HoursPerDay = 24

const DaysPerWeek = 7

const DaysPerBiWeek = 14

const DaysPerMonth = 30

const DaysPerQuarter = 3 * 30

const DaysPerBiAnnum = 6 * 30

const DaysPerAnnum = 12 * 30

func (jdb *JemisonDB) GetNextFetch(fqdn string) time.Time {
	var delta time.Duration

	schedule := config.GetSchedule(fqdn)

	switch schedule {
	case config.Daily:
		delta = time.Duration(HoursPerDay * time.Hour)
	case config.Weekly:
		delta = time.Duration(DaysPerWeek * HoursPerDay * time.Hour)
	case config.BiWeekly:
		delta = time.Duration(DaysPerBiWeek * HoursPerDay * time.Hour)
	case config.Monthly:
		delta = time.Duration(DaysPerMonth * HoursPerDay * time.Hour)
	case config.Quarterly:
		delta = time.Duration(DaysPerQuarter * HoursPerDay * time.Hour)
	case config.BiAnnually:
		delta = time.Duration(DaysPerBiAnnum * HoursPerDay * time.Hour)
	case config.Annually:
		delta = time.Duration(DaysPerAnnum * HoursPerDay * time.Hour)
	case config.Default:
		delta = time.Duration(DaysPerMonth * HoursPerDay * time.Hour)
	default:
		// Default to monthly.
		delta = time.Duration(DaysPerMonth * HoursPerDay * time.Hour)
	}

	nextFetch := time.Now().Add(delta)

	return nextFetch
}

func (jdb *JemisonDB) InThePast(delta time.Duration) time.Time {
	return time.Now().Add(delta * -1)
}
