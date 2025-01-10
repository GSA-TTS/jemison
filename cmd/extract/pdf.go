//nolint:godox
package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"

	kv "github.com/GSA-TTS/jemison/internal/kv"
	"github.com/GSA-TTS/jemison/internal/queueing"
	"github.com/GSA-TTS/jemison/internal/util"
	"github.com/google/uuid"
	"github.com/johbar/go-poppler"
	"go.uber.org/zap"
)

//nolint:funlen
func extractPdf(obj *kv.S3JSON) {
	tempFilename := uuid.NewString()
	raw_copy := obj.Key.Copy()
	raw_copy.Extension = util.Raw

	err := obj.S3.S3ToFile(raw_copy, tempFilename)
	if err != nil {
		zap.L().Error("could not copy s3 object to file",
			zap.String("raw_copy", raw_copy.Render()),
			zap.String("tempFilename", tempFilename))
	}

	defer func() {
		err := os.Remove(tempFilename)
		if err != nil {
			zap.L().Error("could not remove temp file",
				zap.String("filename", tempFilename))
		}
	}()

	fi, err := os.Stat(tempFilename)
	if err != nil {
		zap.L().Fatal(err.Error())
	}

	size := fi.Size()
	zap.L().Debug("tempFilename size", zap.Int64("size", size))

	if size > MAX_FILESIZE {
		// Give up on big files.
		// FIXME: we need to clean up the bucket, too, and delete PDFs there
		zap.L().Debug("file too large, not processing")

		return
	}

	doc, err := poppler.Open(tempFilename)

	if err != nil {
		zap.L().Warn("poppler failed to open pdf",
			zap.String("raw_filename", tempFilename),
			zap.String("key", obj.Key.Render()))

		return
	} else {
		// Pull the metadata out, and include in every object.
		info := doc.Info()

		for page_no := 0; page_no < doc.GetNPages(); page_no++ {
			page_number_anchor := fmt.Sprintf("#page=%d", page_no+1)
			copied_key := obj.Key.Copy()
			copied_key.Path = copied_key.Path + page_number_anchor
			copied_key.Extension = util.JSON

			page := doc.GetPage(page_no)
			// obj.Set("content", util.RemoveStopwords(page.Text()))
			obj.Set("content", page.Text())
			obj.Set("path", copied_key.Path)
			obj.Set("pdf_page_number", fmt.Sprintf("%d", page_no+1))
			obj.Set("title", info.Title)
			obj.Set("creation-date", strconv.Itoa(info.CreationDate))
			obj.Set("modification-date", strconv.Itoa(info.ModificationDate))
			obj.Set("pdf-version", info.PdfVersion)
			obj.Set("pages", strconv.Itoa(info.Pages))

			new_obj := kv.NewFromBytes(
				ThisServiceName,
				obj.Key.Scheme,
				obj.Key.Host,
				obj.Key.Path,
				obj.GetJSON(),
			)

			err = new_obj.Save()
			if err != nil {
				zap.L().Error("could not save object to s3",
					zap.String("key", new_obj.Key.Render()))
			}

			page.Close()

			// Enqueue next steps
			ChQSHP <- queueing.QSHP{
				Queue:  "pack",
				Scheme: obj.Key.Scheme.String(),
				Host:   obj.Key.Host,
				Path:   obj.Key.Path,
			}
			// https://weaviate.io/blog/gomemlimit-a-game-changer-for-high-memory-applications
			// https://stackoverflow.com/questions/38972003/how-to-stop-the-golang-gc-and-trigger-it-manually
			runtime.GC()
		}
	}

	doc.Close()
}
