package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/GSA-TTS/jemison/config"
	"github.com/GSA-TTS/jemison/internal/common"
	"github.com/GSA-TTS/jemison/internal/env"
	"github.com/GSA-TTS/jemison/internal/postgres"
	"github.com/GSA-TTS/jemison/internal/queueing"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var Databases sync.Map
var ChQSHP = make(chan queueing.QSHP)
var ThisServiceName = "resultsapi"
var JDB *postgres.JemisonDB

type PostBody struct {
	Terms     string `json:"terms"`
	Host      string `json:"host"`
	D64_start string `json:"d64_start"`
	D64_end   string `json:"d64_end"`
}

func setupQueues() {
	env.InitGlobalEnv(ThisServiceName)

	InitializeQueues()

	go queueing.Enqueue(ChQSHP)
}

func parseAffiliate(affiliate string) (string, string, string) {
	tld := ""
	domain := ""
	subdomain := ""
	delimiter := "."
	results := strings.Split(affiliate, delimiter)
	if len(results) == 3 {
		subdomain = results[0]
		domain = results[1]
		tld = results[2]
	} else if len(results) == 2 {
		domain = results[0]
		tld = results[1]
	} else {
		tld = results[0]
	}
	return subdomain, domain, tld
}

func getD64(affiliate string) (string, string) {
	var subdomain, domain, tld string
	subdomain, domain, tld = parseAffiliate(affiliate)
	zap.L().Info("PARSED OF "+affiliate+" :",
		zap.String("subdomain", subdomain),
		zap.String("domain", domain),
		zap.String("tld", tld))

	var d64_start, d64_end int64

	// top level domain
	d64_start, _ = strconv.ParseInt(fmt.Sprintf("%02x00000000000000", tld), 16, 64)
	d64_end, _ = strconv.ParseInt(fmt.Sprintf("%02xFFFFFFFFFFFF00", tld), 16, 64)

	// domain
	if domain != "" {
		start := config.RDomainToDomain64(fmt.Sprintf("%s.%s", tld, domain))
		d64_start, _ = strconv.ParseInt(fmt.Sprintf("%s00000000", start), 16, 64)
		d64_end, _ = strconv.ParseInt(fmt.Sprintf("%sFFFFFF00", start), 16, 64)
	} else {
		s_d64_start := fmt.Sprintf("%d", d64_start)
		s_d64_end := fmt.Sprintf("%d", d64_end)
		return s_d64_start, s_d64_end
	}

	//subdomain
	if subdomain != "" {
		fqdn := fmt.Sprintf("%s.%s.%s", subdomain, domain, tld)
		start, _ := config.FQDNToDomain64(fqdn)
		d64_start = start
		d64_end = start + 1
	}

	s_d64_start := fmt.Sprintf("%d", d64_start)
	s_d64_end := fmt.Sprintf("%d", d64_end)
	return s_d64_start, s_d64_end
}

func doTheSearch(affiliate string, searchQuery string) []SearchResult {
	domain64Start, domain64End := getD64(affiliate + ".gov")
	zap.L().Info("DOMAIN 64 OF "+affiliate+" :",
		zap.String("domain64Start", domain64Start),
		zap.String("domain64End", domain64End))
	sri := SearchRequestInput{
		Host:          affiliate + ".gov",
		Path:          "",
		Terms:         searchQuery,
		Domain64Start: domain64Start,
		Domain64End:   domain64End,
	}

	rows, duration, err := runQuery(sri)
	zap.L().Info("Queried Answer:",
		zap.Any("rows: ", rows),
		zap.Any("duration", duration),
		zap.Any("err", err))

	return rows
}

func setUpEngine(staticFilesPath string, templateFilesPath string) *gin.Engine {
	engine := gin.Default()

	// will we need the two instructions below? I think not because there will be no ui
	engine.StaticFS("/static", gin.Dir(staticFilesPath, true))
	engine.LoadHTMLGlob(templateFilesPath + "/*")

	engine.GET("/:search", func(c *gin.Context) {
		//required query parameters
		affiliate := c.Query("affiliate")
		searchQuery := c.Query("query")
		zap.L().Info("Query Data: ",
			zap.String("affiliate", affiliate),
			zap.String("query", searchQuery))

		res := doTheSearch(affiliate, searchQuery)
		//optional query parameters
		// enable_highlighting := c.Query("enable_highlighting")
		// offset := c.Query("offset")
		// sort_by := c.Query("sort_by")
		// sitelimit := c.Query("sitelimit")

		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"res": res,
		})
	})

	v1 := engine.Group("/api")
	{
		v1.GET("/heartbeat", common.Heartbeat)
	}

	return engine
}

func main() {
	env.InitGlobalEnv(ThisServiceName)
	setupQueues()

	s, _ := env.Env.GetUserService(ThisServiceName)
	templateFilesPath := s.GetParamString("template_files_path")
	staticFilesPath := s.GetParamString("static_files_path")

	JDB = postgres.NewJemisonDB()

	zap.L().Info("environment initialized",
		zap.String("ThisServiceName", ThisServiceName))
	engine := setUpEngine(staticFilesPath, templateFilesPath)
	zap.L().Info("listening from resultsapi",
		zap.String("port", env.Env.Port))

	// Local and Cloud should both get this from the environment.
	//nolint:gosec
	err := http.ListenAndServe(":"+env.Env.Port, engine)
	if err != nil {
		zap.L().Error("could not launch HTTP server", zap.Error(err))
	}
}
