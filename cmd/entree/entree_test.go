//nolint:all
package main

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/GSA-TTS/jemison/internal/env"
	work_db "github.com/GSA-TTS/jemison/internal/postgres/work_db"
	"github.com/jackc/pgx/v5"
	"github.com/zeebo/assert"
	"go.uber.org/zap"
)

// Tests need a backend
// docker compose -f backend.yaml up

func setup( /* t *testing.T */ ) func(t *testing.T) {
	os.Setenv("ENV", "LOCALHOST")
	env.InitGlobalEnv("entree")
	env.SetupLogging("entree")
	return func(t *testing.T) {
		// t.Log("teardown test case")
	}
}

func TestIsSingleWithPass(t *testing.T) {
	setup()
	ec, _ := NewEntreeCheck("single",
		"https", "www.imls.gov", "/", true)
	assert.True(t, IsSingleWithPass(ec))
}

func TestIsSingleNoPass(t *testing.T) {
	setup()
	ec, _ := NewEntreeCheck("single",
		"https", "www.imls.gov", "/", false)
	assert.False(t, IsSingleWithPass(ec))
}

func TestIsFullWithPass(t *testing.T) {
	setup()
	ec, _ := NewEntreeCheck("full",
		"https", "www.imls.gov", "/", true)
	assert.True(t, IsFullWithPass(ec))
}

func TestIsFullNoPass(t *testing.T) {
	setup()
	ec, _ := NewEntreeCheck("full",
		"https", "www.imls.gov", "/", false)
	assert.False(t, IsFullNoPass(ec))
}

func TestEvaluateEntreeSinglePass(t *testing.T) {
	setup()
	ec, _ := NewEntreeCheck("single",
		"https", "www.imls.gov", "/", true)
	assert.True(t, IsSingleWithPass(ec))
	EvaluateEntree(ec)
	_, ctx, conn := GetQueuesDb()
	defer conn.Close(ctx)

	cmdtag, err := conn.Exec(ctx,
		`SELECT COUNT(*) 
		FROM river_job 
		WHERE args->>'host' = 'www.imls.gov'`)
	if err != nil {
		t.Error(err)
	}
	assert.True(t, strings.Contains(cmdtag.String(), "SELECT 1"))

	// Cleanup
	cmdtag, err = conn.Exec(ctx, `DELETE FROM river_job WHERE args->>'host' = 'www.imls.gov'`)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, cmdtag.RowsAffected(), 1)
}

type HostRow struct {
	Id        int64
	Host      string
	NextFetch time.Time
}

func TestSetHostNextFetchToYesterday(t *testing.T) {
	setup()
	ec, _ := NewEntreeCheck("full",
		"https", "www.imls.gov", "/", true)
	assert.True(t, IsFullWithPass(ec))

	// This modifies the state of the DB
	SetHostNextFetchToYesterday(ec)

	_, ctx, conn := GetWorkDB()
	defer conn.Close(ctx)

	// See what we did to the DB.
	rows, err := conn.Query(ctx,
		`SELECT host, next_fetch FROM hosts WHERE host = $1`, "www.imls.gov")
	if err != nil {
		t.Error(err)
	}
	defer rows.Close()
	log.Println(rows.CommandTag())

	for rows.Next() {
		var hr HostRow
		err = rows.Scan(&hr.Host, &hr.NextFetch)
		if err != nil {
			t.Error(err)
		}
		log.Println("HR", hr)
		assert.Equal(t, hr.Host, "www.imls.gov")
		assert.True(t, time.Now().After(hr.NextFetch))
	}

	// Put the DB back the way it was.
	_, err = conn.Exec(ctx,
		`UPDATE hosts
			SET
				next_fetch = NOW()+make_interval(days => +8)
			WHERE
    		host = $1`, "www.imls.gov")
	if err != nil {
		t.Error(err)
	}
}

func TestSetGuestbookFetchToYesterdayForHost(t *testing.T) {
	setup()
	ec, _ := NewEntreeCheck("full",
		"https", "www.imls.gov", "/", true)
	assert.True(t, IsFullWithPass(ec))

	// This modifies the state of the DB
	// In this test, there are no entries.
	SetGuestbookFetchToYesterdayForHost(ec)

	_, ctx, conn := GetWorkDB()
	defer conn.Close(ctx)

	// See what we did to the DB.
	rows, err := conn.Query(ctx,
		`SELECT host, next_fetch FROM guestbook WHERE domain64 = $1`, ec.Domain64)
	if err != nil {
		t.Error(err)
	}
	defer rows.Close()

	assert.Equal(t, rows.CommandTag().RowsAffected(), 0)
}

func TestSetGuestbookFetchToYesterdayForHost2(t *testing.T) {
	setup()

	ec, _ := NewEntreeCheck("full",
		"https", "www.imls.gov", "/", true)
	assert.True(t, IsFullWithPass(ec))

	// Insert a guestbook entry to fake things
	_, ctx, conn := GetWorkDB()
	defer conn.Close(ctx)

	// Delete everything from the guestbook for this test.
	_, err := conn.Exec(ctx, "TRUNCATE guestbook")

	if err != nil {
		t.Error(err)
	}

	_, err = conn.Exec(ctx,
		`INSERT INTO guestbook
		(scheme, domain64, path, last_updated, last_fetched, next_fetch)
		VALUES ($1, $2, $3, $4, $5, $6)
		`,
		"https", ec.Domain64, "/",
		time.Now(), time.Now(), time.Now().Add(time.Duration(3*24*time.Hour)))

	if err != nil {
		t.Error(err)
	}

	// This modifies the state of the DB
	// In this test, there are no entries.
	SetGuestbookFetchToYesterdayForHost(ec)

	// See what we did to the DB.
	rows, err := conn.Query(ctx,
		`SELECT host, next_fetch FROM guestbook WHERE domain64 = $1`, ec.Domain64)
	if err != nil {
		t.Error(err)
	}
	defer rows.Close()

	for rows.Next() {
		var hr HostRow
		err = rows.Scan(&hr.Id, &hr.NextFetch)
		if err != nil {
			t.Error(err)
		}
		log.Println("HR", hr)
		assert.Equal(t, hr.Id, ec.Domain64)
		assert.True(t, time.Now().After(hr.NextFetch))
	}

	// Put the DB back the way it was.
	_, err = conn.Exec(ctx,
		`DELETE FROM guestbook
			WHERE
	  		domain64 = $1`, ec.Domain64)
	if err != nil {
		t.Error(err)
	}
}

func GetQueuesDb() (*work_db.Queries, context.Context, *pgx.Conn) {
	ctx := context.Background()
	db_string, err := env.Env.GetDatabaseUrl(env.QueueDatabase)
	if err != nil {
		zap.L().Fatal("could not get db URL for queues-db")
	}
	conn, err := pgx.Connect(ctx, db_string)
	if err != nil {
		zap.L().Fatal("could not connect to queues-db")
	}
	queries := work_db.New(conn)
	return queries, ctx, conn
}

func GetWorkDB() (*work_db.Queries, context.Context, *pgx.Conn) {
	ctx := context.Background()
	db_string, err := env.Env.GetDatabaseUrl(env.JemisonWorkDatabase)
	if err != nil {
		zap.L().Fatal("could not get db URL for work-db")
	}
	conn, err := pgx.Connect(ctx, db_string)
	if err != nil {
		zap.L().Fatal("could not connect to work-db")
	}
	queries := work_db.New(conn)
	return queries, ctx, conn
}
