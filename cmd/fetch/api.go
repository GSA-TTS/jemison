package main

import (
	"net/http"
	"os"

	"github.com/GSA-TTS/jemison/internal/queueing"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var FETCH_API_VERSION = "1.0.0"

type FetchRequestInput struct {
	Scheme string `json:"scheme" maxLength:"10" doc:"Resource scheme"`
	Host   string `json:"host" maxLength:"500" doc:"Host of resource"`
	Path   string `json:"path" maxLength:"1500" doc:"Path to resource"`
	ApiKey string `json:"api-key"`
}

// https://dev.to/kashifsoofi/rest-api-with-go-chi-and-inmemory-store-43ag
func FetchRequestHandler(c *gin.Context) {
	var fri FetchRequestInput
	if err := c.BindJSON(&fri); err != nil {
		return
	}

	if fri.ApiKey == os.Getenv("API_KEY") {
		zap.L().Debug("api enqueue", zap.String("host", fri.Host), zap.String("path", fri.Path))

		// if fetchClient == nil {
		// 	zap.L().Error("fetchClient IS NIL")
		// }

		// fetchClient.Insert(context.Background(), common.FetchArgs{
		// 	Scheme: fri.Scheme,
		// 	Host:   fri.Host,
		// 	Path:   fri.Path,
		// }, &river.InsertOpts{
		// 	Queue: "fetch",
		// })
		ChQSHP <- queueing.QSHP{
			Queue:  "fetch",
			Scheme: fri.Scheme,
			Host:   fri.Host,
			Path:   fri.Path,
		}

		c.IndentedJSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	}
}

func SitemapRequestHandler(c *gin.Context) {
	// pass
}

func ExtendApi(r *gin.Engine) {
	r.PUT("/fetch", FetchRequestHandler)
	r.PUT("/sitemap", SitemapRequestHandler)
}
