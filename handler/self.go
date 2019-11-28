package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ctripcloud/starter/pkg"
	"github.com/ctripcloud/starter/pkg/dto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//SelfVersion handler for describe version of current service
func SelfVersion(c *gin.Context) {
	c.JSON(http.StatusOK, dto.NewSuccessOK(dto.Version{
		Version:   pkg.Version,
		GoVersion: pkg.GoVersion,
		BuildTime: pkg.BuildTime,
	}))
}

//SelfPrometheusMetrics handler to export prometheus metrics
func SelfPrometheusMetrics(c *gin.Context) {
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}
