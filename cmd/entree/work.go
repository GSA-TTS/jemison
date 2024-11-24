package main

import (
	"context"

	"github.com/GSA-TTS/jemison/internal/common"
	"github.com/GSA-TTS/jemison/internal/util"
	"github.com/riverqueue/river"
)

func (w *EntreeWorker) Work(ctx context.Context, job *river.Job[common.EntreeArgs]) error {
	// We're the new front door.
	// When a request comes in, we will run the algorithm described in
	// docs/design_entree.md.

	// Matrix
	// fullCrawl & !pass: check every timeout in the domain.
	// fullCrawl & pass: re-crawl the whole domain now.
	// !fullCrawl & !pass: check
	// !fullCrawl & pass: fetch the page now

	kind := "NOT_VALID_KIND"
	if job.Args.FullCrawl {
		kind = "full"
	} else {
		kind = "single"
	}

	// In case we don't have clean URLs...
	path := "INVALID_PATH"
	if len(job.Args.Path) > 1 {
		path = util.TrimSuffix(job.Args.Path, "/")
	} else {
		path = job.Args.Path
	}
	ec, err := NewEntreeCheck(kind, job.Args.Scheme, job.Args.Host, path, job.Args.HallPass)
	if err != nil {
		// If we cannot create a new EC object, we probably couldn't find the host.
		// A refined error message here would be good. But, what it means is we don't want to
		// requeue the job, and we don't want to proceed.
		return nil
	}
	EvaluateEntree(ec)

	return nil
}
