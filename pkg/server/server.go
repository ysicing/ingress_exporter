package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ergoapi/exgin"
	"github.com/ergoapi/util/version"
	_ "github.com/ergoapi/util/version/prometheus"
	"github.com/ergoapi/zlog"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/ysicing/ingress_exporter/pkg/cron"
	"github.com/ysicing/ingress_exporter/pkg/k8s"
	"k8s.io/client-go/informers"
)

func Serve(ctx context.Context) error {
	defer cron.Cron.Stop()
	cron.Cron.Start()
	g := exgin.Init(true)
	g.Use(exgin.ExCors())
	g.Use(exgin.ExLog())
	g.Use(exgin.ExRecovery())
	g.GET("/metrics", gin.WrapH(promhttp.Handler()))
	g.GET("/kv", func(c *gin.Context) {
		exgin.GinsData(c, map[string]string{
			"k8s_version": k8s.KV(),
		}, nil)
	})
	g.GET("/rv", func(c *gin.Context) {
		v := version.Get()
		exgin.GinsData(c, map[string]string{
			"builddate": v.BuildDate,
			"release":   version.GetShortString(),
			"gitcommit": v.GitCommit,
			"version":   v.GitVersion,
		}, nil)
	})
	g.NoMethod(func(c *gin.Context) {
		msg := fmt.Sprintf("not found: %v", c.Request.Method)
		exgin.GinsAbortWithCode(c, 404, msg)
	})
	g.NoRoute(func(c *gin.Context) {
		msg := fmt.Sprintf("not found: %v", c.Request.URL.Path)
		exgin.GinsAbortWithCode(c, 404, msg)
	})
	cron.Cron.Add("@every 60s", func() {
		zlog.Debug("cron test")
	})
	stopChan := make(chan struct{})
	factory := informers.NewSharedInformerFactory(k8s.K8SClient, time.Hour)
	controller := k8s.NewNamespaceControlller(factory)
	controller.Run(stopChan)
	addr := "0.0.0.0:65001"
	srv := &http.Server{
		Addr:    addr,
		Handler: g,
	}
	go func() {
		defer close(stopChan)
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			zlog.Error("Failed to stop server, error: %s", err)
		}
		zlog.Info("server exited.")
	}()
	zlog.Info("http listen to %v, pid is %v", addr, os.Getpid())
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		zlog.Error("Failed to start http server, error: %s", err)
		return err
	}

	<-stopChan

	return nil
}
