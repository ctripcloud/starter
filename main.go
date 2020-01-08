package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"runtime"

	"github.com/ctripcloud/starter/config"
	"github.com/ctripcloud/starter/db"
	"github.com/ctripcloud/starter/handler"
	"github.com/ctripcloud/starter/logger"
	"github.com/ctripcloud/starter/pkg"
	"github.com/ctripcloud/starter/pkg/signals"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	shortVersion         bool
	verboseVersion       bool
	configFile           string
	secretFile           string
	startTime, readyTime prometheus.Gauge
)

func init() {
	flag.BoolVar(&shortVersion, "version", false, "print short version and then exit")
	flag.BoolVar(&verboseVersion, "Version", false, "print verbose version and then exit")
	flag.StringVar(&configFile, "configFile", "", "filepath of configuration.")
	flag.StringVar(&secretFile, "secretFile", "", "filepath of secret configuration.")

	startTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "starttime",
		Help: "start timestamp of current service",
	})
	prometheus.MustRegister(startTime)

	readyTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "readytime",
		Help: "ready timestamp of current service",
	})
	prometheus.MustRegister(readyTime)
}

func main() {
	flag.Parse()

	if shortVersion {
		fmt.Println(pkg.Version)
		return
	}
	if verboseVersion {
		fmt.Printf("version: %s\ngo version: %s\nbuild time: %s\n", pkg.Version, runtime.Version(), pkg.BuildTime)
		return
	}
	startTime.SetToCurrentTime()
	stopCh := signals.SetupSignalHandler()
	var err error
	if err = config.Init(configFile, secretFile, logger.RuntimeSugar); err != nil {
		logger.RuntimeSugar.Fatalf("config init failed: %v", err)
	}
	defer config.Final()

	cfg := config.GetConfig()
	if err = logger.Init(cfg.Logger.RuntimePath, cfg.Logger.RuntimeRemainDays, cfg.Logger.AccessPath, cfg.Logger.AccessRemainDays); err != nil {
		logger.RuntimeSugar.Fatalf("logger init failed: %v", err)
	}
	defer logger.Final()

	if err = db.Init(); err != nil {
		logger.RuntimeSugar.Fatalf("db init failed: %v", err)
	}
	defer db.Final()

	srv := &http.Server{
		Addr:    cfg.HTTPServer.ListenAddr,
		Handler: handler.NewHTTPHandler(),
	}
	readyTime.SetToCurrentTime()
	go func() {
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.RuntimeSugar.Errorf("Failed to listen and serve http server @%s: %v", cfg.HTTPServer.ListenAddr, err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout (graceShutdownPeriod)
	<-stopCh
	cfg = config.GetConfig()
	logger.RuntimeSugar.Infof("Shutdown Server (within %v seconds)...", cfg.HTTPServer.GraceShutdownPeriod)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTPServer.GraceShutdownPeriod.Duration)
	defer cancel()
	if err = srv.Shutdown(ctx); err != nil {
		logger.RuntimeSugar.Fatalf("Server gracefully shutdown failed: %v", err)
	}
	logger.RuntimeSugar.Info("Server shutdown successfully")
}
