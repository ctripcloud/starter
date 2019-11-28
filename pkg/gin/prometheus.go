package gin

import (
	"strconv"
	"time"

	gincore "github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

// PromObserve observes prometheus metrics for request count and lantency
func PromObserve(nameReqCnt, nameReqDuration string) gincore.HandlerFunc {
	counterReq := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: nameReqCnt,
			Help: "How many HTTP requests processed, partitioned by status code and request path.",
		},
		[]string{"code", "path"},
	)
	durationReq := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    nameReqDuration,
			Help:    "Latency of requests processed, partitioned by status code and request path.",
			Buckets: prometheus.ExponentialBuckets(2, 3.5, 10),
		},
		[]string{"code", "path"})
	return func(c *gincore.Context) {
		start := time.Now()
		defer func() {
			path := c.Request.URL.Path
			code := strconv.Itoa(c.Writer.Status())
			counterReq.WithLabelValues(code, path).Inc()
			durationReq.WithLabelValues(code, path).Observe(float64(time.Now().Sub(start) / time.Millisecond))
		}()
		c.Next()
	}
}
