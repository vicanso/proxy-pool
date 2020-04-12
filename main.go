package main

import (
	"github.com/vicanso/elton"
	"github.com/vicanso/elton/middleware"
	"github.com/vicanso/proxy-pool/config"
	_ "github.com/vicanso/proxy-pool/controller"
	"github.com/vicanso/proxy-pool/log"
	"github.com/vicanso/proxy-pool/router"
	"go.uber.org/zap"
)

func main() {
	logger := log.Default()
	e := elton.New()

	e.Use(func(c *elton.Context) error {
		c.NoCache()
		return c.Next()
	})
	e.Use(middleware.NewDefaultResponder())

	router.Init(e)
	addr := config.GetListenAddr()
	logger.Info("start to linstening...",
		zap.String("listen", addr),
	)
	err := e.ListenAndServe(addr)
	if err != nil {
		panic(err)
	}
}
