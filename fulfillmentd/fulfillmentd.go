package fulfillmentd

import (
	"fmt"
	"fulfillmentd/constants"
	api "fulfillmentd/redeemservice"
	"fulfillmentd/server"
	"fulfillmentd/version"
	"github.com/eluv-io/errors-go"
	elog "github.com/eluv-io/log-go"
	"github.com/gin-gonic/gin"
	"net/http"
)

var log = elog.Get("/fs")

func Init(s *server.Server) error {
	s.Router = gin.Default()
	s.Router.Use(defaultCORS)

	s.FulfillmentService = server.NewFulfillmentService(s)
	log.Info("Init", "service", s.FulfillmentService)

	addBaseRoutes(s.Router)
	api.AddRoutes(s)
	log.Info("registered routes")

	err := s.Router.Run(fmt.Sprintf(":%d", s.Cfg.Port))
	if err != nil {
		return errors.E("error in service Run()", errors.K.Cancelled, "err", err)
	}

	return nil
}

func addBaseRoutes(engine *gin.Engine) {
	defaultRoutes := []*server.Route{
		GET("", func(ctx *gin.Context) { Version(ctx) }),
		GET("/version", func(ctx *gin.Context) { Version(ctx) }),
		GET("/:network/version", func(ctx *gin.Context) { Version(ctx) }),
	}
	routeGroup := server.NewGroup(defaultRoutes...)
	routeGroup.HandleAllRoutes(engine)
}

func Version(ctx *gin.Context) {
	resp := gin.H{
		"name":     constants.DaemonName,
		"version":  "v" + version.BestVersion(),
		"revision": version.Revision(),
		"branch":   version.Branch(),
		"date":     version.Date(),
	}
	ctx.JSON(http.StatusOK, resp)
}

func GET(path string, handler gin.HandlerFunc) *server.Route {
	return server.NewRoute("GET", path, handler)
}

func defaultCORS(ctx *gin.Context) {
	ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	ctx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	ctx.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

	if ctx.Request.Method == "OPTIONS" {
		ctx.AbortWithStatus(http.StatusNoContent)
		return
	}
	ctx.Next()
}
