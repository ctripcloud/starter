package handler

import (
	"github.com/ctripcloud/starter/logger"
	pkgin "github.com/ctripcloud/starter/pkg/gin"
	"github.com/gin-gonic/gin"
)

//NewHTTPHandler create a handler to serve by http server
func NewHTTPHandler() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), pkgin.PromObserve("httpserver_request_count", "httpserver_request_latency"), logger.AccessLoggerForGin())
	//	apis := r.Group("/apis/")
	self := r.Group("/_self")
	self.GET("/version", SelfVersion)
	self.GET("/prometheus", SelfPrometheusMetrics)
	return r
}
