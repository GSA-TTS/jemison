package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/GSA-TTS/jemison/config"
	"github.com/GSA-TTS/jemison/internal/common"
	"github.com/GSA-TTS/jemison/internal/env"
	"github.com/GSA-TTS/jemison/internal/postgres"
	"github.com/GSA-TTS/jemison/internal/postgres/search_db"
	"github.com/GSA-TTS/jemison/internal/postgres/work_db"
	"github.com/GSA-TTS/jemison/internal/queueing"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var Databases sync.Map

var ChQSHP = make(chan queueing.QSHP)

var ThisServiceName = "serve"

var JDB *postgres.JemisonDB

func addMetadata(m map[string]any) map[string]any {
	pathCount, err := JDB.WorkDBQueries.PathsInDomain64Range(context.Background(),
		work_db.PathsInDomain64RangeParams{
			D64Start: m["d64_start"].(int64),
			D64End:   m["d64_end"].(int64),
		})
	if err != nil {
		zap.L().Error(err.Error())

		pathCount = 0
	}

	m["pageCount"] = pathCount

	bodyCount, err := JDB.SearchDBQueries.BodiesInDomain64Range(context.Background(),
		search_db.BodiesInDomain64RangeParams{
			D64Start: m["d64_start"].(int64),
			D64End:   m["d64_end"].(int64),
		})
	if err != nil {
		zap.L().Error(err.Error())

		bodyCount = 0
	}

	m["bodyCount"] = bodyCount

	return m
}

//nolint:funlen
func main() {
	env.InitGlobalEnv(ThisServiceName)

	InitializeQueues()

	go queueing.Enqueue(ChQSHP)

	s, _ := env.Env.GetUserService(ThisServiceName)
	templateFilesPath := s.GetParamString("template_files_path")
	staticFilesPath := s.GetParamString("static_files_path")

	externalHost := s.GetParamString("external_host")
	externalPort := s.GetParamInt64("external_port")

	JDB = postgres.NewJemisonDB()

	log.Println("environment initialized")

	zap.L().Info("serve environment",
		zap.String("template_files_path", templateFilesPath),
		zap.String("external_host", externalHost),
		zap.Int64("external_port", externalPort),
	)

	/////////////////////
	// Server/API
	engine := gin.Default()
	// engine.GET("/", func(c *gin.Context) {
	// 	c.Redirect(http.StatusMovedPermanently, "/search/"+start)
	// })
	// engine.GET("/search", func(c *gin.Context) {
	// 	c.Redirect(http.StatusMovedPermanently, "/search/"+start)
	// })
	engine.StaticFS("/static", gin.Dir(staticFilesPath, true))
	// engine.GET("/search/:host", ServeHost)

	engine.LoadHTMLGlob(templateFilesPath + "/*")

	baseParams := gin.H{
		"scheme":      "http",
		"search_host": "localhost",
		"search_port": "10000",
	}

	engine.GET("/:tld", func(c *gin.Context) {
		tld := config.GetTLD(c.Param("tld"))
		d64Start, _ := strconv.ParseInt(fmt.Sprintf("%02x00000000000000", tld), 16, 64)
		d64End, _ := strconv.ParseInt(fmt.Sprintf("%02xFFFFFFFFFFFF00", tld), 16, 64)
		baseParams["tld"] = c.Param("tld")
		delete(baseParams, "domain")
		delete(baseParams, "subdomain")
		baseParams["fqdn"] = c.Param("tld")
		baseParams["d64_start"] = d64Start
		baseParams["d64_end"] = d64End
		baseParams = addMetadata(baseParams)

		c.HTML(http.StatusOK, "index.tmpl", baseParams)
	})

	engine.GET("/:tld/:domain", func(c *gin.Context) {
		tld := c.Param("tld")
		domain := c.Param("domain")
		start := config.RDomainToDomain64(fmt.Sprintf("%s.%s", tld, domain))
		zap.L().Debug("rdomain", zap.String("start", start))

		d64Start, _ := strconv.ParseInt(fmt.Sprintf("%s00000000", start), 16, 64)
		d64End, _ := strconv.ParseInt(fmt.Sprintf("%sFFFFFF00", start), 16, 64)

		baseParams["tld"] = tld
		baseParams["domain"] = domain
		delete(baseParams, "subdomain")
		baseParams["fqdn"] = fmt.Sprintf("%s.%s", domain, tld)
		baseParams["d64_start"] = d64Start
		baseParams["d64_end"] = d64End
		baseParams = addMetadata(baseParams)

		c.HTML(http.StatusOK, "index.tmpl", baseParams)
	})

	engine.GET("/:tld/:domain/:subdomain", func(c *gin.Context) {
		tld := c.Param("tld")
		domain := c.Param("domain")
		subdomain := c.Param("subdomain")
		fqdn := fmt.Sprintf("%s.%s.%s", subdomain, domain, tld)
		start, _ := config.FQDNToDomain64(fqdn)
		d64Start := start
		d64End := start + 1

		baseParams["tld"] = tld
		baseParams["domain"] = domain
		baseParams["subdomain"] = subdomain
		baseParams["fqdn"] = fqdn
		baseParams["d64_start"] = d64Start
		baseParams["d64_end"] = d64End
		baseParams = addMetadata(baseParams)
		c.HTML(http.StatusOK, "index.tmpl", baseParams)
	})

	v1 := engine.Group("/api")
	{
		v1.GET("/heartbeat", common.Heartbeat)
		v1.POST("/search", SearchHandler)
	}

	zap.L().Info("listening to the music of the spheres",
		zap.String("port", env.Env.Port))
	// Local and Cloud should both get this from the environment.
	//nolint:gosec
	err := http.ListenAndServe(":"+env.Env.Port, engine)
	if err != nil {
		zap.Error(err)
	}
}
