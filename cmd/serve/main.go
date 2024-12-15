package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/GSA-TTS/jemison/config"
	"github.com/GSA-TTS/jemison/internal/common"
	"github.com/GSA-TTS/jemison/internal/env"
	"github.com/GSA-TTS/jemison/internal/postgres"
	"github.com/GSA-TTS/jemison/internal/queueing"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var Databases sync.Map //map[string]*sql.DB
var ChQSHP = make(chan queueing.QSHP)
var ThisServiceName = "serve"
var JDB *postgres.JemisonDB

func main() {
	env.InitGlobalEnv(ThisServiceName)
	//s3 := kv.NewS3(ThisServiceName)
	InitializeQueues()

	go queueing.Enqueue(ChQSHP)

	s, _ := env.Env.GetUserService(ThisServiceName)
	template_files_path := s.GetParamString("template_files_path")
	static_files_path := s.GetParamString("static_files_path")

	external_host := s.GetParamString("external_host")
	external_port := s.GetParamInt64("external_port")

	JDB = postgres.NewJemisonDB()

	log.Println("environment initialized")

	zap.L().Info("serve environment",
		zap.String("template_files_path", template_files_path),
		zap.String("external_host", external_host),
		zap.Int64("external_port", external_port),
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
	engine.StaticFS("/static", gin.Dir(static_files_path, true))
	// engine.GET("/search/:host", ServeHost)

	engine.LoadHTMLGlob(template_files_path + "/*")

	base_params := gin.H{
		"scheme":      "http",
		"search_host": "localhost",
		"search_port": "10000",
	}

	engine.GET("/:tld", func(c *gin.Context) {
		tld := config.GetTLD(c.Param("tld"))
		d64_start, _ := strconv.ParseInt(fmt.Sprintf("%02x00000000000000", tld), 16, 64)
		d64_end, _ := strconv.ParseInt(fmt.Sprintf("%02xFFFFFFFFFFFF00", tld), 16, 64)
		base_params["tld"] = c.Param("tld")
		base_params["fqdn"] = c.Param("tld")
		base_params["d64_start"] = d64_start
		base_params["d64_end"] = d64_end
		c.HTML(http.StatusOK, "index.tmpl", base_params)
	})

	engine.GET("/:tld/:domain", func(c *gin.Context) {
		tld := c.Param("tld")
		domain := c.Param("domain")
		start := config.RDomainToDomain64(fmt.Sprintf("%s.%s", tld, domain))
		zap.L().Debug("rdomain", zap.String("start", start))

		d64_start, _ := strconv.ParseInt(fmt.Sprintf("%s00000000", start), 16, 64)
		d64_end, _ := strconv.ParseInt(fmt.Sprintf("%sFFFFFF00", start), 16, 64)

		base_params["tld"] = tld
		base_params["domain"] = domain
		base_params["fqdn"] = fmt.Sprintf("%s.%s", domain, tld)
		base_params["d64_start"] = d64_start
		base_params["d64_end"] = d64_end
		c.HTML(http.StatusOK, "index.tmpl", base_params)
	})

	engine.GET("/:tld/:domain/:subdomain", func(c *gin.Context) {
		tld := c.Param("tld")
		domain := c.Param("domain")
		subdomain := c.Param("subdomain")
		fqdn := fmt.Sprintf("%s.%s.%s", subdomain, domain, tld)
		start, _ := config.FQDNToDomain64(fqdn)
		d64_start := start
		d64_end := start + 1

		base_params["tld"] = tld
		base_params["domain"] = domain
		base_params["subdomain"] = subdomain
		base_params["fqdn"] = fqdn
		base_params["d64_start"] = d64_start
		base_params["d64_end"] = d64_end
		c.HTML(http.StatusOK, "index.tmpl", base_params)
	})

	v1 := engine.Group("/api")
	{
		v1.GET("/heartbeat", common.Heartbeat)
		v1.POST("/search", SearchHandler)
	}

	zap.L().Info("listening to the music of the spheres",
		zap.String("port", env.Env.Port))
	// Local and Cloud should both get this from the environment.
	http.ListenAndServe(":"+env.Env.Port, engine)

}
