package api

import (
	"fulfillmentd/server"
	"github.com/gin-gonic/gin"
	"net/http"
)

// addStatusRoutes adds /health (/version is in fulfillmentd/fulfillmentd.go addBaseRoutes)
func addStatusRoutes(s *server.Server) {
	s.Router.GET("/health", Health(s))
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
