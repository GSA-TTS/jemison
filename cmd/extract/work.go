//nolint:godox
package main

import (
	"context"

	"github.com/GSA-TTS/jemison/internal/common"
	"github.com/GSA-TTS/jemison/internal/env"
	"github.com/GSA-TTS/jemison/internal/kv"
	"github.com/GSA-TTS/jemison/internal/util"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

const MAX_FILESIZE = 5000000

// FIXME: This is checking the size of the JSON,
// not the size of the .raw file.
func isTooLarge(obj *kv.S3JSON) bool {
	return obj.Size() > MAX_FILESIZE
}

func extract(obj *kv.S3JSON) {
	mime_type := obj.GetString("content-type")
	s, _ := env.Env.GetUserService(ThisServiceName)

	switch mime_type {
	case "text/html":
		if s.GetParamBool("extract_html") {
			extractHtml(obj)
		}
	case "application/pdf":
		if s.GetParamBool("extract_pdf") {
			if !isTooLarge(obj) {
				extractPdf(obj)
			} else {
				// FIXME: This should be deleted at this point, if we get here.
				zap.L().Error("s3json object too large",
					zap.String("host", obj.Key.Host), zap.String("path", obj.Key.Path))
			}
		}
	}
}

func (w *ExtractWorker) Work(ctx context.Context, job *river.Job[common.ExtractArgs]) error {
	zap.L().Info("extracting",
		zap.String("host", job.Args.Host),
		zap.String("path", job.Args.Path))

	s3json := kv.NewEmptyS3JSON("fetch",
		util.ToScheme(job.Args.Scheme),
		job.Args.Host,
		job.Args.Path)

	err := s3json.Load()
	if err != nil {
		zap.L().Error("could not load s3 JSON",
			zap.String("host", job.Args.Host),
			zap.String("path", job.Args.Path))
	}

	extract(s3json)

	zap.L().Debug("extraction finished")

	return nil
}
