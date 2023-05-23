package authority

import (
	"fulfillmentd/authority/config"
	"fulfillmentd/authority/db"
	lg "github.com/eluv-io/log-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
	"runtime"
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
	g := newGroup("")
	g.AddRoutes(routes...)
	return g
}

func newGroup(basePath string) *group {
	return &group{basePath: basePath}
}

type group struct {
	basePath   string
	middleware []interface{}
	routes     []*Route
	subs       []*group
}

func (g *group) WithBasePath(p string) *group {
	g.basePath = p
	return g
}

func (g *group) AddRoutes(r ...*Route) *group {
	g.routes = append(g.routes, r...)
	return g
}

func (g *group) AddSubs(s ...*group) *group {
	g.subs = append(g.subs, s...)
	return g
}

func (g *group) HandleAllRoutes(engine *gin.Engine) {
	for _, rt := range g.routes {
		//log.Trace("handle", "rt", rt, "handler", HandlerName(rt.handler))
		engine.Handle(rt.verb, rt.path, rt.handler)
	}
}

func HandlerName(handler gin.HandlerFunc) string {
	return runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
}

func AddMiddleware(s *Server) {
	s.Router.Use(defaultCORS)
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

type Route struct {
	verb    string
	path    string
	handler gin.HandlerFunc
}

func NewRoute(verb string, path string, handler gin.HandlerFunc) *Route {
	return &Route{
		verb:    verb,
		path:    path,
		handler: handler,
	}
}
