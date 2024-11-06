package main

import (
	"context"
	"log"

	"github.com/GSA-TTS/jemison/internal/common"
	"github.com/GSA-TTS/jemison/internal/env"
	"github.com/GSA-TTS/jemison/internal/kv"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

func extract(obj kv.Object) {
	mime_type := obj.GetMimeType()
	s, _ := env.Env.GetUserService("extract")

	switch mime_type {
	case "text/html":
		if s.GetParamBool("extract_html") {
			log.Println("EXTRACT HTML")
			extractHtml(obj)
		}
	case "application/pdf":
		if s.GetParamBool("extract_pdf") {
			log.Println("EXTRACT PDF")
			extractPdf(obj)
		}
	}
}

func (w *ExtractWorker) Work(ctx context.Context, job *river.Job[common.ExtractArgs]) error {

	zap.L().Debug("extracting", zap.String("key", job.Args.Key))

	obj, err := fetchStorage.Get(job.Args.Key)
	if err != nil {
		zap.L().Error("could not fetch key from bucket",
			zap.String("key", job.Args.Key))
		return err
	}

	extract(obj)

	zap.L().Debug("extraction finished")
	return nil

}
