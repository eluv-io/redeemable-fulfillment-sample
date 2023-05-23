package config

import (
	"time"
)

type DbConfig struct {
	Username      string
	Password      string
	Host          string
	Port          uint16
	DefaultDb     string
	MaxConn       int
	ConnTimeoutMS int
	RunMigrations bool
	SSLMode       string // disable, verify-full
	SSLCert       string
	SSLKey        string
	SSLRootCert   string
}

type AuthorityConfig struct {
	DevMode             bool
	ShutdownGracePeriod time.Duration
	DbConfig            DbConfig
	Port                int
}
