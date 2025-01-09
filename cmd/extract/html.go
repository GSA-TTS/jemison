package main

import (
	"encoding/json"
	"os"
	"strings"

	kv "github.com/GSA-TTS/jemison/internal/kv"
	"github.com/GSA-TTS/jemison/internal/queueing"
	"github.com/GSA-TTS/jemison/internal/util"
	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// https://github.com/PuerkitoBio/goquery/issues/443
// Someone wanting to pull a full-HTML text from a page...
// Metadata scraping
// https://jonathanmh.com/p/web-scraping-golang-goquery/
// Vectors
// https://alexgarcia.xyz/sqlite-vec/go.html
// https://www.zenrows.com/blog/goquery

func scrape_sel(sel *goquery.Selection) string {
	txt := sel.Text()
	repl := strings.ToLower(txt)
	return util.CollapseWhitespace(repl)
}

func _getTitle(doc *goquery.Document) string {
	// Some pages are just really malformed.
	// It turns out there are title tags elsewhere in the doc.
	title := ""
	doc.Find("title").Each(func(ndx int, sel *goquery.Selection) {
		if title == "" {
			title = scrape_sel(sel)
		}
	})
	return util.CollapseWhitespace(title)
}

func _getHeaders(doc *goquery.Document) map[string][]string {
	// Build an array of headers at each level
	headers := make(map[string][]string, 0)

	for _, tag := range []string{
		"h1",
		"h2",
		"h3",
		"h4",
		"h5",
		"h6",
		"h7",
		"h8",
	} {
		accum := make([]string, 0)
		doc.Find(tag).Each(func(ndx int, sel *goquery.Selection) {
			accum = append(accum, util.CollapseWhitespace(scrape_sel(sel)))
		})
		headers[tag] = accum
	}
	return headers
}

func _getBodyContent(doc *goquery.Document) string {
	// Remove some things.
	// Get rid of headers and footers.
	doc.Find("header").Remove()
	doc.Find("footer").Remove()
	doc.Find(".usa-banner").Remove()
	doc.Find(".usa-footer").Remove()

	content := ""
	for _, elem := range []string{
		"p",
		"li",
		"td",
		"div",
		"span",
		"a",
		"small",
		"b",
		"bold",
		"em",
		"i",
	} {
		// zap.L().Debug("looking for", zap.String("elem", elem))
		doc.Find(elem).Each(func(ndx int, sel *goquery.Selection) {
			// zap.L().Debug("element", zap.String("sel", scrape_sel(sel)))
			content += scrape_sel(sel)
		})
	}

	// Get rid of all extraneous whitespace
	return util.CollapseWhitespace(content)
}

// //////////////////
// extractHtml loads the following keys into the JSON
//
// * title: string
// * headers: []string (as JSON)
// * body : string

func extractHtml(obj *kv.S3JSON) {
	// rawFilename := obj.GetString("raw")
	rawFilename := uuid.NewString()
	// The file is not in this service... it's in the `fetch` bucket.`
	s3 := kv.NewS3("fetch")

	raw_key := obj.Key.Copy()
	raw_key.Extension = util.Raw
	zap.L().Debug("looking up raw key", zap.String("raw_key", raw_key.Render()))
	err := s3.S3ToFile(raw_key, rawFilename)
	if err != nil {
		zap.L().Error("could not create tempfile from s3",
			zap.String("raw_key", raw_key.Render()),
			zap.String("rawfile", rawFilename))
	}
	rawFile, err := os.Open(rawFilename)
	if err != nil {
		zap.L().Error("cannot open tempfile", zap.String("filename", rawFilename))
		return
	}
	defer func() {
		rawFile.Close()
		os.Remove(rawFilename)
	}()

	doc, err := goquery.NewDocumentFromReader(rawFile)
	if err != nil {
		zap.L().Error("cannot create new doc from raw file",
			zap.String("rawFilename", rawFilename),
			zap.String("rawKey", raw_key.Render()))
		return
	}

	title := _getTitle(doc)
	headers := _getHeaders(doc)
	content := _getBodyContent(doc)

	zap.L().Debug("found content",
		zap.Int("headers", len(headers)),
		zap.Int("content length", len(content)))

	// Store everything
	copied_key := obj.Key.Copy()
	copied_key.Extension = util.JSON
	// This is because we were holding an object from the "fetch" bucket.
	new_obj := kv.NewFromBytes(
		ThisServiceName,
		obj.Key.Scheme,
		obj.Key.Host,
		obj.Key.Path,
		obj.GetJSON())

	// Load up the object
	new_obj.Set("title", title)
	// Marshal headers to JSON
	jsonString, err := json.Marshal(headers)
	if err != nil {
		zap.L().Error("could not marshal headers to JSON", zap.String("title", title))
		return
	}
	new_obj.Set("headers", string(jsonString))
	new_obj.Set("body", content)
	err = new_obj.Save()

	if err != nil {
		zap.L().Error("could not save object", zap.String("key", new_obj.Key.Render()))
	}

	// Enqueue next steps
	ChQSHP <- queueing.QSHP{
		Queue:  "pack",
		Scheme: obj.Key.Scheme.String(),
		Host:   obj.Key.Host,
		Path:   obj.Key.Path,
	}

}
