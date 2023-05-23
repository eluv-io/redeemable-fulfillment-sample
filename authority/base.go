package authority

import (
	"fulfillmentd/authority/config"
	"fulfillmentd/authority/db"
	lg "github.com/eluv-io/log-go"
	"github.com/gin-gonic/gin"
	"net/http"
)

var log = lg.Get("/fs/authority")

type Server struct {
	http        *http.Server
	Router      *gin.Engine
	AdminRouter *gin.Engine
	middleware  struct {
		clientToken gin.HandlerFunc
		metrics     gin.HandlerFunc
	}

	Cfg               *config.AuthorityConfig
	ConnectionManager *db.ConnectionManager

	FulfillmentService *FulfillmentService
}

type Route struct {
	verb    string
	path    string
	handler gin.HandlerFunc
}

type group struct {
	basePath   string
	middleware []interface{}
	routes     []*Route
	subs       []*group
}

func ConnectDb(cfg *config.AuthorityConfig) (s *Server, err error) {
	log.Info("StartServer", "DbConfig", cfg.DbConfig)
	s = &Server{Cfg: cfg}

	if s.ConnectionManager, err = db.NewConnectionManager(cfg.DbConfig); err != nil {
		log.Error("error connecting", err)
		return
	}

	return s, nil
}

func NewGroup(routes ...*Route) *group {
	g := &group{basePath: ""}
	g.routes = append(g.routes, routes...)
	return g
}

func NewRoute(verb string, path string, handler gin.HandlerFunc) *Route {
	return &Route{
		verb:    verb,
		path:    path,
		handler: handler,
	}
}

func (g *group) HandleAllRoutes(engine *gin.Engine) {
	for _, rt := range g.routes {
		engine.Handle(rt.verb, rt.path, rt.handler)
	}
}
