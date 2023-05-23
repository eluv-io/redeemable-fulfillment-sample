package authority

import "github.com/gin-gonic/gin"

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
