package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gokits/rfw"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type accessEntity struct {
	Time      string
	Method    string
	Path      string
	Status    int
	Client    string
	UserAgent string
	ReqSize   int64
	RspSize   int
	Latency   int
}

func (ae *accessEntity) Reset() {
	ae.Status = 0
	ae.ReqSize = -1
	ae.RspSize = -1
}

var (
	runtimeLw     *rfw.Rfw
	runtimeSyncer *MutableWriteSyncer

	// Runtime logger for runtime logging
	Runtime *zap.Logger
	// RuntimeSugar sugar logger for runtime logging
	RuntimeSugar *zap.SugaredLogger

	accessWriter *rfw.Rfw
	accessPool   = sync.Pool{
		New: func() interface{} {
			return &accessEntity{
				Status:  0,
				ReqSize: -1,
				Latency: -1,
			}
		},
	}
)

// MutableWriteSyncer a WriteSyncer implementation support change inner WriteSyncer on the fly
type MutableWriteSyncer struct {
	syncer atomic.Value
}

func NewMutableWriteSyncer(defaultSyncer zapcore.WriteSyncer) *MutableWriteSyncer {
	mws := &MutableWriteSyncer{}
	mws.syncer.Store(&defaultSyncer)
	return mws
}

func (mws *MutableWriteSyncer) get() zapcore.WriteSyncer {
	return *(mws.syncer.Load().(*zapcore.WriteSyncer))
}

func (mws *MutableWriteSyncer) SetWriteSyncer(newSyncer zapcore.WriteSyncer) {
	mws.syncer.Store(&newSyncer)
}

func (mws *MutableWriteSyncer) Write(p []byte) (n int, err error) {
	return mws.get().Write(p)
}

func (mws *MutableWriteSyncer) Sync() error {
	return mws.get().Sync()
}

func init() {
	runtimeSyncer = NewMutableWriteSyncer(zapcore.Lock(zapcore.AddSync(os.Stdout)))
	runtimeCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
		}),
		runtimeSyncer,
		zapcore.InfoLevel,
	)
	Runtime = zap.New(runtimeCore)
	RuntimeSugar = Runtime.Sugar()
}

func AccessLoggerForGin() gin.HandlerFunc {
	encoder := json.NewEncoder(accessWriter)
	return func(c *gin.Context) {
		ae := accessPool.Get().(*accessEntity)
		defer func() {
			accessPool.Put(ae)
		}()
		ae.Reset()
		ae.Method = c.Request.Method
		ae.Path = c.Request.URL.Path
		ae.Client = c.ClientIP()
		if c.Request.URL.RawQuery != "" {
			ae.Path += "?" + c.Request.URL.RawQuery
		}
		ae.ReqSize = c.Request.ContentLength
		start := time.Now()
		defer func() {
			end := time.Now()
			ae.Latency = int(end.Sub(start) / time.Millisecond)
			ae.Status = c.Writer.Status()
			ae.RspSize = c.Writer.Size()
			ae.Time = end.Format("2006/01/02 - 15:04:05.000")
			encoder.Encode(ae)
		}()
		// Process request
		c.Next()
	}
}

//Init initializer of this module
func Init(runtimeLog string, runtimeRemaindays int, accessLog string, accessRemaindays int) (err error) {
	runtimeLw, err = rfw.NewWithOptions(runtimeLog, rfw.WithCleanUp(runtimeRemaindays))
	if err != nil {
		return fmt.Errorf("open rfw for path %s failed: %v", runtimeLog, err)
	}
	runtimeSyncer.SetWriteSyncer(zapcore.AddSync(runtimeLw))

	accessWriter, err = rfw.NewWithOptions(accessLog, rfw.WithCleanUp(accessRemaindays))
	if err != nil {
		runtimeSyncer.SetWriteSyncer(zapcore.AddSync(os.Stdout))
		runtimeLw.Close()
		return fmt.Errorf("open rfw for path %s failed: %v", accessLog, err)
	}
	return nil
}

//Final finalizer of this module
func Final() {
	Runtime.Sync()
	runtimeLw.Close()

	accessWriter.Close()
}
