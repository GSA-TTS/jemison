package main

import (
	"context"

	"github.com/GSA-TTS/jemison/config"
	"github.com/GSA-TTS/jemison/internal/kv"
	"github.com/GSA-TTS/jemison/internal/postgres/search_db"
	"go.uber.org/zap"
)

func packPdf(s3json *kv.S3JSON) {
	// We have more fields than before.
	d64, err := config.FQDNToDomain64(s3json.Key.Host)
	if err != nil {
		zap.L().Error("could not get Domain64 for host", zap.String("host", s3json.Key.Host))
	}

	///////////////////////
	// PATH
	err = JDB.SearchDBQueries.InsertRawContent(
		context.Background(),
		search_db.InsertRawContentParams{
			Domain64: d64,
			Path:     s3json.GetString("path"),
			Tag:      "path",
			Content:  s3json.GetString("path"),
		})
	if err != nil {
		zap.L().Error("could not insert path when packing",
			zap.String("_key", s3json.GetString("_key")),
			zap.String("err", err.Error()),
			zap.String("path", s3json.GetString("path")),
		)
	}

	///////////////////////
	// TITLE
	err = JDB.SearchDBQueries.InsertRawContent(
		context.Background(),
		search_db.InsertRawContentParams{
			Domain64: d64,
			Path:     s3json.GetString("path"),
			Tag:      "title",
			Content:  s3json.GetString("title") + " (PDF page " + s3json.GetString("pdf_page_number") + ")",
		})
	if err != nil {
		zap.L().Error("could not insert title when packing",
			zap.String("_key", s3json.GetString("_key")),
			zap.String("err", err.Error()),
			zap.String("path", s3json.GetString("title")),
		)
	}

	///////////////////////
	// BODY
	err = JDB.SearchDBQueries.InsertRawContent(
		context.Background(),
		search_db.InsertRawContentParams{
			Domain64: d64,
			// s3json.Key.Path yields the path to the PDF without a page number
			// which does not help when we're searching/returning results.
			Path:    s3json.GetString("path"),
			Tag:     "body",
			Content: s3json.GetString("content"),
		})
	if err != nil {
		zap.L().Error("could not insert raw content",
			zap.Int64("domain64", d64),
			zap.String("path", s3json.GetString("path")))
	}
}
