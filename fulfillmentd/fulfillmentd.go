package fulfillmentd

import (
	"fmt"
	"fulfillmentd/authority"
	api "fulfillmentd/redeemservice"
	"github.com/eluv-io/errors-go"
	elog "github.com/eluv-io/log-go"
	"github.com/gin-gonic/gin"
	"net/http"
)

var log = elog.Get("/fs")

func Init(s *authority.Server) error {
	s.Router = gin.Default()

	s.FulfillmentService = authority.NewFulfillmentService(s)
	log.Info("Init", "service", s.FulfillmentService)

	AddBaseRoutes(s.Router)
	api.AddRoutes(s)
	log.Info("registered routes")

	err := s.Router.Run(fmt.Sprintf(":%d", s.Cfg.Port))
	if err != nil {
		return errors.E("error in service Run()", errors.K.Cancelled, "err", err)
	}

	return nil
}

func AddBaseRoutes(engine *gin.Engine) {
	defaultRoutes := []*authority.Route{
		GET("", func(ctx *gin.Context) { Version(ctx) }),
		GET("/version", func(ctx *gin.Context) { Version(ctx) }),
	}
	routeGroup := authority.NewGroup(defaultRoutes...)
	routeGroup.HandleAllRoutes(engine)
}

func Version(ctx *gin.Context) {
	resp := gin.H{
		"version": "v.0.0.1",
	}
	ctx.JSON(http.StatusOK, resp)
}

func GET(path string, handler gin.HandlerFunc) *authority.Route {
	return authority.NewRoute("GET", path, handler)
}

func POST(path string, handler gin.HandlerFunc) *authority.Route {
	return authority.NewRoute("POST", path, handler)
}
