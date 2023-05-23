package main

import (
	"fmt"
	"fulfillmentd/authority"
	"fulfillmentd/authority/config"
	"fulfillmentd/fulfillmentd"
	"github.com/eluv-io/errors-go"
	elog "github.com/eluv-io/log-go"
	"github.com/eluv-io/log-go/handlers/console"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

type ConfigState struct {
	Config     string
	LogFile    string
	LogHandler string
	Verbosity  string
}

func toVerbosity(verbosity int) string {
	switch verbosity {
	case 0:
		return "fatal"
	case 1:
		return "error"
	case 2:
		return "warn"
	case 3:
		return "info"
	case 4:
		return "debug"
	case 5:
		return "trace"
	default:
		panic("bad verbosity level")
	}
}

func HandleConfig(cfg *ConfigState, prefix string) error {
	var filename string
	var err error

	filename = filepath.Base(cfg.Config)
	viper.SetConfigName(filename[:len(filename)-len(filepath.Ext(filename))])
	viper.AddConfigPath(filepath.Dir(cfg.Config))

	if err = viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file - %v", err)
	}

	cfg.LogFile = viper.GetString(prefix + ".log_file")
	cfg.LogHandler = viper.GetString(prefix + ".log_handler")
	cfg.Verbosity = toVerbosity(viper.GetInt(prefix + ".verbosity"))

	return nil
}

var (
	trueVal = true

	cfgState = ConfigState{}

	readConfig = func() error {
		if err := HandleConfig(&cfgState, "fulfillmentd"); err != nil {
			log.Error("error parsing", "config file", err)
			return err
		}
		return nil
	}

	log = elog.Get("/fs/fulfillmentd")
)

func getBaseConfig(cfg *config.AuthorityConfig) (err error) {
	if err = readConfig(); err != nil {
		log.Error("readConfig error", err)
		return
	}

	if cfg.DbConfig, err = getDbConfig(cfg); err != nil {
		log.Error("getDbConfig error", err)
		return
	}

	cfg.Port = viper.GetInt("fulfillmentd.service_port")

	return
}

func getDbConfig(_ *config.AuthorityConfig) (dbCfg config.DbConfig, err error) {
	dbCfg = config.DbConfig{
		Username:      viper.GetString("db.username"),
		Password:      viper.GetString("db.password"),
		Host:          viper.GetString("db.host"),
		Port:          uint16(viper.GetUint("db.port")),
		DefaultDb:     viper.GetString("db.database"),
		MaxConn:       viper.GetInt("db.max_conn"),
		ConnTimeoutMS: viper.GetInt("db.conn_timeout_ms"),
		RunMigrations: viper.GetBool("db.run_migrations"),
		SSLMode:       viper.GetString("db.ssl_mode"),
	}

	switch dbCfg.SSLMode {
	case "", "disable":
		log.Warn("disabling TLS for database")
	case "verify-full":
		dbCfg.SSLCert = viper.GetString("db.ssl_cert")
		dbCfg.SSLKey = viper.GetString("db.ssl_key")
		dbCfg.SSLRootCert = viper.GetString("db.ssl_root_cert")
	default:
		err = errors.E("invalid ssl mode for database", "mode", dbCfg.SSLMode)
		return
	}

	return
}

func Config(configFile string) (cfg *config.AuthorityConfig, err error) {
	log.Debug("config", "file", configFile)
	viper.SetDefault("fulfillmentd.service_port", 2023)
	viper.SetDefault("fulfillmentd.log_file", "fulfillmentd")
	viper.SetDefault("fulfillmentd.log_handler", "console")
	viper.SetDefault("fulfillmentd.verbosity", 3)
	viper.SetConfigFile(configFile)

	cfg = &config.AuthorityConfig{}
	err = getBaseConfig(cfg)
	if err != nil {
		return nil, err
	}

	logConfig := &elog.Config{
		Level:   cfgState.Verbosity,
		Handler: cfgState.LogHandler,
		File:    &elog.LumberjackConfig{Filename: cfgState.LogFile},
		Caller:  &trueVal,
	}
	elog.SetDefault(logConfig)

	if lh, ok := log.Handler().(*console.Handler); ok {
		lh.WithTimestamps(true)
	}

	log.Debug("ports", "service_port", cfg.Port)
	log.Debug("BaseConfig", "cfg", cfg)

	return cfg, nil
}

func StartServer(configFile string) (s *authority.Server, err error) {
	cfg, e := Config(configFile)
	if e != nil {
		return nil, err
	}

	s, err = authority.ConnectDb(cfg)
	if err != nil {
		return nil, err
	}
	log.Debug("StartServer", "Server:", s)

	err = fulfillmentd.Init(s)
	if err != nil {
		return nil, err
	}

	return
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: fulfillmentd --config <config.toml>")
		return
	}

	if _, err := StartServer(os.Args[2]); err != nil {
		fmt.Println("cannot launch", "Error", err)
		log.Error("cannot launch", "Error", err)
		os.Exit(1)
	}
}
