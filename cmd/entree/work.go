package main

import (
	"context"
	"strings"

	"github.com/GSA-TTS/jemison/internal/common"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

// entree-1             | {"level":"fatal","timestamp":"2024-11-27T13:42:14.431Z","caller":"work_db/custom.go:29","msg":"could not connect to DB1","pid":13,"stacktrace":"github.com/GSA-TTS/jemison/internal/work_db/work_db.GetWorkDbQueryContext\n\t/home/jadudm/git/search/jemison/app/internal/work_db/work_db/custom.go:29\ngithub.com/GSA-TTS/jemison/internal/work_db/work_db.UpdateNextFetch\n\t/home/jadudm/git/search/jemison/app/internal/work_db/work_db/custom.go:36\nmain.EvaluateEntree\n\t/home/jadudm/git/search/jemison/app/cmd/entree/accept_logic.go:81\nmain.(*EntreeWorker).Work\n\t/home/jadudm/git/search/jemison/app/cmd/entree/work.go:45\ngithub.com/riverqueue/river.(*wrapperWorkUnit[...]).Work\n\t/home/jadudm/go/pkg/mod/github.com/riverqueue/river@v0.13.0/work_unit_wrapper.go:30\ngithub.com/riverqueue/river.(*jobExecutor).execute.func2\n\t/home/jadudm/go/pkg/mod/github.com/riverqueue/river@v0.13.0/job_executor.go:216\ngithub.com/riverqueue/river.(*jobExecutor).execute\n\t/home/jadudm/go/pkg/mod/github.com/riverqueue/river@v0.13.0/job_executor.go:239\ngithub.com/riverqueue/river.(*jobExecutor).Execute\n\t/home/jadudm/go/pkg/mod/github.com/riverqueue/river@v0.13.0/job_executor.go:157"}

func (w *EntreeWorker) Work(ctx context.Context, job *river.Job[common.EntreeArgs]) error {
	// We're the new front door.
	// When a request comes in, we will run the algorithm described in
	// docs/design_entree.md.

	// Matrix
	// fullCrawl & !pass: check every timeout in the domain.
	// fullCrawl & pass: re-crawl the whole domain now.
	// !fullCrawl & !pass: check
	// !fullCrawl & pass: fetch the page now

	var kind string
	if job.Args.FullCrawl {
		kind = "full"
	} else {
		kind = "single"
	}

	// In case we don't have clean URLs...
	if len(job.Args.Path) > 0 {
		path := strings.TrimSpace(job.Args.Path)
		//path = util.TrimSuffix(path, "/")
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
