package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/GSA-TTS/jemison/internal/common"
	"github.com/GSA-TTS/jemison/internal/env"
	"github.com/GSA-TTS/jemison/internal/kv"
	"github.com/GSA-TTS/jemison/internal/queueing"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var Databases sync.Map //map[string]*sql.DB
var ChQSHP = make(chan queueing.QSHP)
var ThisServiceName = "serve"

func ServeHost(c *gin.Context) {
	s, _ := env.Env.GetUserService("serve")
	external_scheme := s.GetParamString("external_scheme")
	external_host := s.GetParamString("external_host")
	external_port := s.GetParamInt64("external_port")
	static_files_path := s.GetParamString("static_files_path")

	zap.L().Debug("serving up a host", zap.String("external_host", external_host))

	host := c.Param("host")
	data, err := os.ReadFile(static_files_path + "/index.html")
	if err != nil {
		log.Println("SERVE could not read index.html")
		log.Fatal(err)
	}
	data = bytes.ReplaceAll(data, []byte("{SCHEME}"), []byte(external_scheme))
	data = bytes.ReplaceAll(data, []byte("{HOST}"), []byte(external_host))
	data = bytes.ReplaceAll(data, []byte("{SEARCH_HOST}"), []byte(host))

	data = bytes.ReplaceAll(data, []byte("{PORT}"), []byte(fmt.Sprintf("%d", external_port)))

	c.Data(http.StatusOK, "text/html", data)
}

func MultiStatsHandler(c *gin.Context) {
	dbs := listHostedDomains()
	res := make(map[string]any)
	base_stats := common.NewBaseStats("serve")
	res["stats"] = base_stats.GetAll()
	for _, db := range dbs {
		st := common.NewBaseStats(db)
		all := st.GetAll()
		page_count := countPages(db)
		all["pages"] = page_count
		res[db] = all
	}
	res["hosted_domains"] = dbs

	c.JSON(http.StatusOK, res)
}

func CheckS3ForDatabases(storage *kv.S3) {
	objects, err := storage.List("")
	if err != nil {
		zap.L().Error("problem listing objects")
	}
	// For each object, queue ourselves to download the file.
	// Why? Because we have the machinery in the worker, and it
	// might as well do the work that way.
	for _, obj := range objects {
		//queueing.InsertServe(obj.Key)
		zap.L().Debug("copying in database",
			zap.String("key", obj.Key))
		ChQSHP <- queueing.QSHP{
			Queue:    "serve",
			Filename: obj.Key,
		}
	}
}

func main() {
	env.InitGlobalEnv(ThisServiceName)
	s3 := kv.NewS3(ThisServiceName)
	InitializeQueues()

	go queueing.Enqueue(ChQSHP)
	CheckS3ForDatabases(s3)

	s, _ := env.Env.GetUserService(ThisServiceName)
	static_files_path := s.GetParamString("static_files_path")
	external_host := s.GetParamString("external_host")
	external_port := s.GetParamInt64("external_port")

	// Databases = make(map[string]*sql.DB, 0)

	log.Println("environment initialized")

	zap.L().Info("serve environment",
		zap.String("static_files_path", static_files_path),
		zap.String("external_host", external_host),
		zap.Int64("external_port", external_port),
	)

	dbs := listHostedDomains()

	start := "start"
	if len(dbs) > 0 {
		start = dbs[0]
	}

	/////////////////////
	// Server/API
	engine := gin.Default()
	engine.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/search/"+start)
	})
	engine.GET("/search", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/search/"+start)
	})
	engine.StaticFS("/static", gin.Dir(static_files_path, true))
	engine.GET("/search/:host", ServeHost)
	v1 := engine.Group("/api")
	{
		v1.GET("/heartbeat", common.Heartbeat)
		v1.POST("/search", SearchHandler)
		v1.GET("/databases", DatabasesHandler)
		v1.GET("/stats", MultiStatsHandler)
	}

	zap.L().Info("listening to the music of the spheres",
		zap.String("port", env.Env.Port))
	// Local and Cloud should both get this from the environment.
	http.ListenAndServe(":"+env.Env.Port, engine)

}
