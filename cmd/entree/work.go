package main

import (
	"context"
	"strings"

	"github.com/GSA-TTS/jemison/internal/common"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

// We're the new front door.
// When a request comes in, we will run the algorithm described in
// docs/design_entree.md.

/*
// Matrix
// fullCrawl & !pass: check every timeout in the domain.
// fullCrawl & pass: re-crawl the whole domain now.
// !fullCrawl & !pass: check
// !fullCrawl & pass: fetch the page now
*/
func (w *EntreeWorker) Work(ctx context.Context, job *river.Job[common.EntreeArgs]) error {
	var kind string
	if job.Args.FullCrawl {
		kind = "full"
	} else {
		kind = "single"
	}

	// In case we don't have clean URLs...
	if len(job.Args.Path) > 0 {
		path := strings.TrimSpace(job.Args.Path)

		if path == "" {
			path = "/"
		}

		ec, err := NewEntreeCheck(kind, job.Args.Scheme, job.Args.Host, path, job.Args.HallPass)
		if err != nil {
			// If we cannot create a new EC object, we probably couldn't find the host.
			// A refined error message here would be good. But, what it means is we don't want to
			// requeue the job, and we don't want to proceed.
			return nil
		}

		EvaluateEntree(ec)
	} else {
		zap.L().Debug("skipping zero-length path", zap.String("host", job.Args.Host))
	}

	return nil
}
