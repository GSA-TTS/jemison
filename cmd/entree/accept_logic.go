//nolint:godox
package main

import (
	"context"
	"time"

	"github.com/GSA-TTS/jemison/config"
	work_db "github.com/GSA-TTS/jemison/internal/postgres/work_db"
	"github.com/GSA-TTS/jemison/internal/queueing"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

const SINGLE_PASS = "single"

const FULL_PASS = "full"

// The front line of questions involve whether or not
// it is a single URL and if there is a hall pass.

// FIXME: add the host_id here. Pass it through.
type EntreeCheck struct {
	// "full" or "single"
	Kind     string
	HallPass bool
	Scheme   string
	Host     string
	Domain64 int64
	Path     string
}

// FIXME: someday, it would be good to decide what is limited to package,
// and what can be accessed outside. Then, these become lowercase...

func NewEntreeCheck(kind, scheme, host, path string, hallPass bool) (*EntreeCheck, error) {
	// host_id, err := JDB.WorkDBQueries.GetHostId(ctx, host)
	d64, err := config.FQDNToDomain64(host)
	if err != nil {
		//nolint:wrapcheck
		return nil, err
	}

	return &EntreeCheck{
		Kind:     kind,
		HallPass: hallPass,
		Scheme:   scheme,
		Host:     host,
		Domain64: d64,
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

	// FIXME: We set the fetch to yesterday, then set it to now (below)?
	// This feels wrong. Redundant. One of these is not needed?
	// Or... is it necessary with multiple workers? Probably.

	if it_shall_pass {
		// We need to update the guestbook now, because we will end up re-walking
		// the page if we don't. This is true in each case.
		// Fetch will update a second time.
		scheme := JDB.GetScheme(ec.Scheme)
		next_fetch := JDB.GetNextFetch(ec.Host)

		_, err := JDB.WorkDBQueries.UpdateGuestbookNextFetch(context.Background(),
			work_db.UpdateGuestbookNextFetchParams{
				Scheme:   scheme,
				Domain64: ec.Domain64,
				Path:     ec.Path,
				NextFetch: pgtype.Timestamp{
					Time:             next_fetch,
					Valid:            true,
					InfinityModifier: 0,
				},
			},
		)
		if err != nil {
			zap.L().Error("failed to update guestbook next fetch",
				zap.Int64("domain64", ec.Domain64), zap.String("path", ec.Path))
		}

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
	return ec.Kind == SINGLE_PASS && ec.HallPass
}

// A single URL with no pass is most likely a URL
// coming in from `walk`. This is the 99% case.
//
// Possible side-effects:
//   - Fetch the page
//   - Update last_fetch in guestbook
func IsSingleNoPass(ec *EntreeCheck) bool {
	return ec.Kind == SINGLE_PASS && !ec.HallPass && CheckIfIsInGuestbook(ec)
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
	return ec.Kind == FULL_PASS && ec.HallPass
}

// This is probably a nightly enqueue.
//
// Possible side-effects:
//   - None. It runs on what is in the DBs.
func IsFullNoPass(ec *EntreeCheck) bool {
	return ec.Kind == FULL_PASS && !ec.HallPass && CheckIfAfterHostNextFetch(ec)
}

// Support functions

func isInGuestbook(ec *EntreeCheck) bool {
	ctx := context.Background()

	b, err := JDB.WorkDBQueries.CheckEntryExistsInGuestbook(ctx, ec.Domain64)
	if err != nil {
		zap.L().Fatal("could not check if in guestbook",
			zap.Int64("domain64", ec.Domain64),
			zap.String("domain64_hex", config.Dec64ToHex(ec.Domain64)))
	}

	return b
}

func CheckIfAfterGuestbookNextFetch(ec *EntreeCheck) bool {
	ctx := context.Background()

	entry, err := JDB.WorkDBQueries.GetGuestbookEntry(ctx, work_db.GetGuestbookEntryParams{
		Domain64: ec.Domain64,
		Path:     ec.Path,
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

	ts, err := JDB.WorkDBQueries.GetHostNextFetch(ctx, ec.Domain64)
	if err != nil {
		// If it isn't in the host table, then return false
		return false
	}

	return time.Now().After(ts.Time)
}

func SetHostNextFetchToYesterday(ec *EntreeCheck) {
	ctx := context.Background()

	err := JDB.WorkDBQueries.SetHostNextFetchToYesterday(ctx, ec.Domain64)
	if err != nil {
		zap.L().Error("could not set host fetch to yesterday",
			zap.String("host", ec.Host))
	}
}

func SetGuestbookFetchToYesterdayForHost(ec *EntreeCheck) {
	ctx := context.Background()

	err := JDB.WorkDBQueries.SetGuestbookFetchToYesterdayForHost(ctx, ec.Domain64)
	if err != nil {
		zap.L().Fatal("could not set guestbook to yesterday for host",
			zap.String("host", ec.Host))
	}
}
