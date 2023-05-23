package authority

import (
	"github.com/gin-gonic/gin"
	"reflect"
	"runtime"
)

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
