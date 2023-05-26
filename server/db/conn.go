package db

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"fulfillmentd/server/config"
	elog "github.com/eluv-io/log-go"
	"io/ioutil"
	"time"

	"github.com/jackc/pgx"
)

var log = elog.Get("/fs/conn")

type ConnectionManager struct {
	cfg  config.DbConfig
	pool *pgx.ConnPool
}

func NewConnectionManager(cfg config.DbConfig) (m *ConnectionManager, err error) {
	connConfig := pgx.ConnConfig{
		Host:     cfg.Host,
		Port:     cfg.Port,
		Database: cfg.DefaultDb,
		User:     cfg.Username,
		Password: cfg.Password,
	}

	switch cfg.SSLMode {
	case "", "disable":
		connConfig.TLSConfig = nil
	case "verify-full":
		var cert tls.Certificate
		cert, err = tls.LoadX509KeyPair(cfg.SSLCert, cfg.SSLKey)
		if err != nil {
			return
		}
		var caCert []byte
		caCert, err = ioutil.ReadFile(cfg.SSLRootCert)
		if err != nil {
			return
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		var tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
		}
		tlsConfig.ServerName = cfg.Host
		tlsConfig.BuildNameToCertificate()

		connConfig.TLSConfig = tlsConfig
	default:
		err = errors.New(fmt.Sprintf("invalid sslmode for database connection '%v'", cfg.SSLMode))
	}

	poolConfig := pgx.ConnPoolConfig{
		ConnConfig:     connConfig,
		MaxConnections: cfg.MaxConn, // suggest = (number of cores * 4) where number of cores is CPUs in cluster
		AcquireTimeout: time.Duration(cfg.ConnTimeoutMS) * time.Millisecond,
		AfterConnect: func(conn *pgx.Conn) (err error) {
			log.Info(fmt.Sprintf("DB connection established: %v\n", conn.RuntimeParams))
			return
		},
	}
	var pool *pgx.ConnPool
	if pool, err = pgx.NewConnPool(poolConfig); err != nil {
		return
	}
	m = &ConnectionManager{
		cfg:  cfg,
		pool: pool,
	}
	return
}

func (c *ConnectionManager) Close() {
	c.pool.Close()
}

func (c *ConnectionManager) GetConn() *pgx.ConnPool {
	return c.pool
}
