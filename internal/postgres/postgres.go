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

	for _, db_name := range []string{env.QueueDatabase, env.JemisonWorkDatabase, env.SearchDatabase} {
		db_string, err := env.Env.GetDatabaseUrl(db_name)
		if err != nil {
			zap.L().Fatal("could not get db URL", zap.String("db_name", db_name))
		}

		cfg := Config(db_string)
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

		jdb.Config[db_name] = cfg
		jdb.Pool[db_name] = pool
	}

	jdb.WorkDBQueries = work_db.New(jdb.Pool[env.JemisonWorkDatabase])
	jdb.SearchDBQueries = search_db.New(jdb.Pool[env.SearchDatabase])

	return &jdb
}

func Config(db_string string) *pgxpool.Config {
	const defaultMaxConns = int32(100)

	const defaultMinConns = int32(0)

	const defaultMaxConnLifetime = time.Hour

	const defaultMaxConnIdleTime = time.Minute * 30

	const defaultHealthCheckPeriod = time.Minute

	const defaultConnectTimeout = time.Second * 5

	dbConfig, err := pgxpool.ParseConfig(db_string)
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
		v, assert_ok := val.(int32)
		if !assert_ok {
			zap.L().Error("could not convert scheme integer")
		}

		return v
	} else {
		scheme_int := config.GetScheme(scheme)
		// This is a guaranteed save conversion
		jdb.constCache.Store("scheme:"+scheme, int32(scheme_int))

		return int32(scheme_int)
	}
}

func (jdb *JemisonDB) GetContentType(ct string) int {
	if val, ok := jdb.constCache.Load("contenttype:" + ct); ok {
		v, assert_ok := val.(int)
		if !assert_ok {
			zap.L().Error("could not convert content type integer")
		}

		return v
	} else {
		ct_int := config.GetContentType(ct)
		jdb.constCache.Store("contenttype:"+ct, ct_int)

		return ct_int
	}
}

const HOURS_PER_DAY = 24

const DAYS_PER_WEEK = 7

const DAYS_PER_BIWEEK = 14

const DAYS_PER_MONTH = 30

const DAYS_PER_QUARTER = 3 * 30

const DAYS_PER_BIANNUM = 6 * 30

const DAYS_PER_ANNUM = 12 * 30

func (jdb *JemisonDB) GetNextFetch(fqdn string) time.Time {
	var delta time.Duration

	schedule := config.GetSchedule(fqdn)

	switch schedule {
	case config.Daily:
		delta = time.Duration(HOURS_PER_DAY * time.Hour)
	case config.Weekly:
		delta = time.Duration(DAYS_PER_WEEK * HOURS_PER_DAY * time.Hour)
	case config.BiWeekly:
		delta = time.Duration(DAYS_PER_BIWEEK * HOURS_PER_DAY * time.Hour)
	case config.Monthly:
		delta = time.Duration(DAYS_PER_MONTH * HOURS_PER_DAY * time.Hour)
	case config.Quarterly:
		delta = time.Duration(DAYS_PER_QUARTER * HOURS_PER_DAY * time.Hour)
	case config.BiAnnually:
		delta = time.Duration(DAYS_PER_BIANNUM * HOURS_PER_DAY * time.Hour)
	case config.Annually:
		delta = time.Duration(DAYS_PER_ANNUM * HOURS_PER_DAY * time.Hour)
	default:
		// Default to monthly.
		delta = time.Duration(DAYS_PER_MONTH * HOURS_PER_DAY * time.Hour)
	}

	next_fetch := time.Now().Add(delta)

	return next_fetch
}

func (jdb *JemisonDB) InThePast(delta time.Duration) time.Time {
	return time.Now().Add(delta * -1)
}
