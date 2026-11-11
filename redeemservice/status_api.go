package api

import (
	"fulfillmentd/server"
	"fulfillmentd/version"
	"github.com/gin-gonic/gin"
	"net/http"
)

// addStatusRoutes adds /version and /health
func addStatusRoutes(s *server.Server) {
	s.Router.GET("/version", Version)
	s.Router.GET("/health", Health(s))
}

// Version godoc
// @ID version
// @Summary Version info of the running build
// @Produce  json
// @Router /version [GET]
func Version(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"name":     "fulfillmentd",
		"version":  version.BestVersion(),
		"revision": version.Revision(),
		"branch":   version.Branch(),
		"date":     version.Date(),
	})
}

// Health godoc
// @ID health
// @Summary 200 if the service can reach its database, 503 if not
// @Produce  json
// @Router /health [GET]
func Health(s *server.Server) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, err := s.ConnectionManager.GetConn().Exec("SELECT 1")
		if err != nil {
			log.Warn("health check: database unreachable", "err", err)
			ctx.JSON(http.StatusServiceUnavailable, gin.H{"result": "Failure", "reason": "database unreachable"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"result": "Success"})
	}
}
