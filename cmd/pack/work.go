package main

import (
	"context"

	"github.com/GSA-TTS/jemison/internal/common"
	"github.com/GSA-TTS/jemison/internal/kv"
	"github.com/GSA-TTS/jemison/internal/util"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

func (w *PackWorker) Work(_ context.Context, job *river.Job[common.PackArgs]) error {
	// It comes in with the GuestbookId. That's all we need (plus the S3 object).
	s3 := kv.NewS3("extract")
	key := util.CreateS3Key(util.ToScheme(job.Args.Scheme), job.Args.Host, job.Args.Path, util.JSON)

	s3json, err := s3.S3PathToS3JSON(key)
	if err != nil {
		zap.L().Error("could not fetch object for packing",
			zap.String("key", s3json.Key.Render()))

		return nil
	}

	contentType := s3json.GetString("content-type")
	switch contentType {
	case "text/html":
		packHTML(s3json)
	case "application/pdf":
		packPdf(s3json)
	}

	return nil
}
