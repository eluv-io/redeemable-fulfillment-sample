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

func ConnectDb(cfg *config.AuthorityConfig) (s *Server, err error) {
	log.Info("StartServer", "DbConfig", cfg.DbConfig)
	s = &Server{Cfg: cfg}

	if s.ConnectionManager, err = db.NewConnectionManager(cfg.DbConfig); err != nil {
		log.Error("error connecting", err)
		return
	}

	return s, nil
}
