//nolint:godox
package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/GSA-TTS/jemison/config"
	"github.com/GSA-TTS/jemison/internal/postgres/search_db"
	"github.com/gin-gonic/gin"
	"github.com/kljensen/snowball"
	"go.uber.org/zap"
)

type SearchRequestInput struct {
	Host          string `json:"host"`
	Path          string `json:"path"`
	Terms         string `json:"terms"`
	Domain64Start string `json:"d64_start"`
	Domain64End   string `json:"d64_end"`
}

type SearchResult struct {
	Terms      string
	PathID     int64
	PageTitle  string
	Kind       int64
	Weight     float64
	Rank       float64
	Text       string
	PathString string
	Snippet    string
	FQDN       string
}

func to64(s string) int64 {
	v, _ := strconv.Atoi(s)

	return int64(v)
}

// Would just be * with SQLite.
var _stemmed = ":*"

//nolint:funlen
func runQuery(sri SearchRequestInput) ([]SearchResult, time.Duration, error) {
	start := time.Now()

	zap.L().Debug("searching for Domain64 range",
		zap.Int64("start", to64(sri.Domain64Start)),
		zap.Int64("end", to64(sri.Domain64End)))

	// Don't only use the stemmed words
	existingTerms := strings.Split(sri.Terms, " ")
	zap.L().Debug("EXISTING TERMS", zap.Strings("terms", existingTerms))

	query := NewQuery()

	for _, et := range existingTerms {
		et = strings.TrimSpace(et)
		stemmed, err := snowball.Stem(et, "english", true)
		zap.L().Debug("stemmed result", zap.String("et", et), zap.String("stemmed", stemmed))

		if err != nil {
			zap.L().Debug("stemming error", zap.String("err", err.Error()))
		}

		query.AddToQuery(Or(et, stemmed+_stemmed))
	}

	improvedTermsString := query.ToString()

	zap.L().Debug("search string",
		zap.String("original", sri.Terms),
		zap.String("Q", fmt.Sprintln(query)),
		zap.String("improved", improvedTermsString))

	res, err := JDB.SearchDBQueries.SearchContent(context.Background(),
		search_db.SearchContentParams{
			Query:    improvedTermsString,
			D64Start: to64(sri.Domain64Start),
			D64End:   to64(sri.Domain64End),
		})

	duration := time.Since(start)

	cleaned := make([]SearchResult, 0)

	for _, r := range res {
		// FIXME: the database structure is forcing us into an N+1 queries
		// situation... Not good.
		d64, _ := config.HexToDec64(r.ToHex)
		fqdn, _ := config.Domain64ToFQDN(d64)

		zap.L().Debug("results for domain",
			zap.Int64("domain64", d64), zap.String("fqdn", fqdn))

		title, err := JDB.SearchDBQueries.GetTitle(context.Background(),
			search_db.GetTitleParams{
				Domain64: d64,
				Path:     r.Path,
			})
		if err != nil {
			title = "<no title>"
		}

		path, err := JDB.SearchDBQueries.GetPath(context.Background(),
			search_db.GetPathParams{
				Domain64: d64,
				Path:     r.Path,
			})
		if err != nil {
			path = r.Path
		}

		cleaned = append(cleaned, SearchResult{
			Terms:      improvedTermsString,
			PageTitle:  title,
			PathString: path,
			Snippet:    string(r.Snippet),
			Rank:       float64(r.Rank),
			FQDN:       fqdn,
		})
	}

	//nolint:wrapcheck
	return cleaned, duration, err
}

func SearchHandler(c *gin.Context) {
	var sri SearchRequestInput

	if err := c.BindJSON(&sri); err != nil {
		return
	}

	rows, duration, err := runQuery(sri)
	if err != nil {
		c.IndentedJSON(http.StatusOK, gin.H{
			"result":  "err",
			"message": err.Error(),
			"elapsed": duration,
			"results": nil,
		})

		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{
		"result":  "ok",
		"elapsed": duration,
		"results": rows,
	})
}
