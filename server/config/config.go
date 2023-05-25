package config

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
	DbConfig               DbConfig
	Port                   int
	EthUrl                 string
	ContentFabricConfigUrl string
}
