package main

import (
	"context"
	"fmt"
	"time"

	"github.com/GSA-TTS/jemison/internal/queueing"
	"github.com/GSA-TTS/jemison/internal/work_db/work_db"
	"go.uber.org/zap"
)

// The front line of questions involve whether or not
// it is a single URL and if there is a hall pass.

// FIXME: add the host_id here. Pass it through
type EntreeCheck struct {
	// "full" or "single"
	Kind     string
	HallPass bool
	Scheme   string
	Host     string
	HostId   int64
	Path     string
}

// FIXME: someday, it would be good to decide what is limited to package,
// and what can be accessed outside. Then, these become lowercase...

func NewEntreeCheck(kind, scheme, host, path string, hallPass bool) (*EntreeCheck, error) {
	ctx := context.Background()
	host_id, err := WDB.Queries.GetHostId(ctx, host)
	if err != nil {
		zap.L().Debug("could not get host id",
			zap.String("host", host),
			zap.String("err", err.Error()))
		return nil, fmt.Errorf("could not get host id")
	}

	return &EntreeCheck{
		Kind:     kind,
		HallPass: hallPass,
		Scheme:   scheme,
		Host:     host,
		HostId:   host_id,
		Path:     path,
	}, nil
}

func EvaluateEntree(ec *EntreeCheck) {
	it_shall_pass := false

	if IsSingleWithPass(ec) {
		zap.L().Debug("is-single-with-pass",
			zap.String("host", ec.Host), zap.String("path", ec.Path))
		it_shall_pass = true
	} else if IsSingleNoPass(ec) {
		zap.L().Debug("is-single-no-pass",
			zap.String("host", ec.Host), zap.String("path", ec.Path))
		it_shall_pass = true
	} else if IsFullWithPass(ec) {
		zap.L().Debug("is-full-with-pass",
			zap.String("host", ec.Host), zap.String("path", ec.Path))
		SetHostNextFetchToYesterday(ec)
		SetGuestbookFetchToYesterdayForHost(ec)
		it_shall_pass = true
	} else if IsFullNoPass(ec) {
		zap.L().Debug("is-full-no-pass",
			zap.String("host", ec.Host), zap.String("path", ec.Path))
		it_shall_pass = true
	} else {
		zap.L().Debug("no entree evaluation criteria met",
			zap.String("host", ec.Host), zap.String("path", ec.Path))
		it_shall_pass = false
	}

	if it_shall_pass {
		// We need to update the guestbook now, because we will end up re-walking
		// the page if we don't. This is true in each case.
		// Fetch will update a second time.
		WDB.Queries.UpdateNextFetch(work_db.FetchUpdateParams{
			Scheme:       ec.Scheme,
			Host:         ec.Host,
			Path:         ec.Path,
			LastModified: time.Now(),
		})

		ChQSHP <- queueing.QSHP{
			Queue:  "fetch",
			Scheme: ec.Scheme,
			Host:   ec.Host,
			Path:   ec.Path,
		}
	}
}

// The most likely situation for a single URL with a
// pass is a partner indicating they want a single page
// updated immediately as part of a site revision.
//
// Possible side-effects:
//   - Fetch the page
//   - Update the last_fetch in guestbook
func IsSingleWithPass(ec *EntreeCheck) bool {
	// This just allows us to queue this onward to `fetch`.
	// Fetch will handle guestbook updates.
	return ec.Kind == "single" && ec.HallPass
}

// A single URL with no pass is most likely a URL
// coming in from `walk`. This is the 99% case.
//
// Possible side-effects:
//   - Fetch the page
//   - Update last_fetch in guestbook
func IsSingleNoPass(ec *EntreeCheck) bool {
	return ec.Kind == "single" && !ec.HallPass && CheckIfIsInGuestbook(ec)
}

func CheckIfIsInGuestbook(ec *EntreeCheck) bool {
	// This is currently multiple database hits.
	// It is a place for optimization. But the first
	// implementation is for simplicity.
	if isInGuestbook(ec) {
		return CheckIfAfterGuestbookNextFetch(ec)
	} else {
		return CheckIfAfterHostNextFetch(ec)
	}
}

// This is if we are re-running a site at a time that
// is in-between scheduled fetches.
//
// Possible side-effects:
//   - Set next_fetch in hosts table for *yesterday*
//   - Set next_fetch for all known pages in guestbook to *yesterday*
//   - Set last_fetch in guestbook
//   - Reset next_fetch in hosts table after completion
func IsFullWithPass(ec *EntreeCheck) bool {
	return ec.Kind == "full" && ec.HallPass
}

// This is probably a nightly enqueue.
//
// Possible side-effects:
//   - None. It runs on what is in the DBs.
func IsFullNoPass(ec *EntreeCheck) bool {
	return ec.Kind == "full" && !ec.HallPass && CheckIfAfterHostNextFetch(ec)
}

// Support functions

// func GetQ() (*work_db.Queries, context.Context, *pgx.Conn) {
// 	ctx := context.Background()
// 	db_string, err := env.Env.GetDatabaseUrl(env.JemisonWorkDatabase)
// 	if err != nil {
// 		zap.L().Fatal("could not get db URL for work-db")
// 	}
// 	conn, err := pgx.Connect(ctx, db_string)
// 	if err != nil {
// 		zap.L().Fatal("could not connect to work-db")
// 	}
// 	queries := work_db.New(conn)
// 	return queries, ctx, conn
// }

func isInGuestbook(ec *EntreeCheck) bool {
	ctx := context.Background()
	b, err := WDB.Queries.CheckEntryExistsInGuestbook(ctx, ec.HostId)
	if err != nil {
		zap.L().Fatal("could not check if in guestbook",
			zap.Int64("host_id", ec.HostId))
	}
	return b
}

func CheckIfAfterGuestbookNextFetch(ec *EntreeCheck) bool {
	ctx := context.Background()
	entry, err := WDB.Queries.GetGuestbookEntry(ctx, work_db.GetGuestbookEntryParams{
		Host: ec.HostId,
		Path: ec.Path,
	})
	if err != nil {
		// If it isn't in the guestbook, then return `true`,
		// because we want to fetch the page.
		return true
	}

	return time.Now().After(entry.NextFetch.Time)
}

func CheckIfAfterHostNextFetch(ec *EntreeCheck) bool {
	ctx := context.Background()
	ts, err := WDB.Queries.GetHostNextFetch(ctx, ec.HostId)
	if err != nil {
		// If it isn't in the host table, then return false
		return false
	}

	return time.Now().After(ts.Time)
}

func SetHostNextFetchToYesterday(ec *EntreeCheck) {
	ctx := context.Background()
	err := WDB.Queries.SetHostNextFetchToYesterday(ctx, ec.Host)
	if err != nil {
		zap.L().Error("could not set host fetch to yesterday",
			zap.String("host", ec.Host))
	}
}

func SetGuestbookFetchToYesterdayForHost(ec *EntreeCheck) {
	ctx := context.Background()
	err := WDB.Queries.SetGuestbookFetchToYesterdayForHost(ctx, ec.HostId)
	if err != nil {
		zap.L().Fatal("could not set guestbook to yesterday for host",
			zap.String("host", ec.Host))
	}
}
