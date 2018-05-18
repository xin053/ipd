package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/getsentry/raven-go"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/xin053/ipd/api"
	"github.com/xin053/ipd/config"
	"github.com/xin053/ipd/es"
	"github.com/xin053/ipd/geolite2"
	"github.com/xin053/ipd/ip2region"
	"github.com/xin053/ipd/middleware"
	"github.com/xin053/ipd/qqwry"
	"github.com/xin053/ipd/server"
)

var router *gin.Engine

func main() {
	gin.SetMode(gin.ReleaseMode)
	router = gin.Default()

	router.Use(middleware.JSONLogMiddleware())
	router.Use(middleware.AuthRequired())
	router.Use(middleware.CORS(middleware.CORSOptions{}))
	if config.UseSentry {
		router.Use(middleware.Sentry(raven.DefaultClient, false))
	}

	initRouters()

	log.Info("Service starting on port " + config.Port)

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT)
	go func() {
		<-signals
		geolite2.DB2.Close()
		ip2region.DB3.Close()
		es.Client.Stop()
		log.Info("closing databases, preparing exit...bye")
		os.Exit(0)
	}()

	router.Run(":" + config.Port)
}

func initRouters() {
	v1 := router.Group("v1")
	{
		v1.POST("/db", qqwry.FromDb)
		v1.POST("/db2", geolite2.FromDb2)
		v1.POST("/db3", ip2region.FromDb3)
		v1.POST("/api", api.FromAPI)
		v1.POST("/ip", server.GetIP)
	}

	serverList := map[string]server.Server{
		"ip2region": &ip2region.IPdDb3{},
		"chunzhen":  &qqwry.IPdDb{},
		"geolite2":  &geolite2.IPdDb2{},
		"api":       &api.IPdAPI{},
	}
	// register server
	if len(config.RequestOrder) > 0 {
		serverList[config.RequestOrder[0]].Register(true)
		for _, server := range config.RequestOrder[1:len(config.RequestOrder)] {
			serverList[server].Register(false)
		}
	}
}

func init() {
	if config.UseSentry {
		raven.SetDSN(config.SentryDSN)
	}
}
