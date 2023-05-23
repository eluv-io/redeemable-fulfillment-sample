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

const (
	DaemonName = "fulfillmentd"
)

var (
	cfgState = ConfigState{}
	log      = elog.Get("/fs")
)

func main() {
	if len(os.Args) < 3 {
		fmt.Sprintf("Usage: %s --config <config.toml>\n", DaemonName)
		return
	}

	if _, err := startServer(os.Args[2]); err != nil {
		fmt.Println("cannot launch", "Error", err)
		os.Exit(1)
	}
}

func startServer(configFile string) (s *authority.Server, err error) {
	cfg, err := loadConfig(configFile)
	if err != nil {
		return
	}

	s, err = authority.ConnectDb(cfg)
	if err != nil {
		return
	}

	err = fulfillmentd.Init(s)
	if err != nil {
		return
	}

	return
}

func loadConfig(configFile string) (cfg *config.AuthorityConfig, err error) {
	log.Debug("config", "file", configFile)
	viper.SetDefault(DaemonName+".service_port", 2023)
	viper.SetDefault(DaemonName+".log_file", DaemonName)
	viper.SetDefault(DaemonName+".log_handler", "console")
	viper.SetDefault(DaemonName+".verbosity", 3)
	viper.SetConfigFile(configFile)

	cfg = &config.AuthorityConfig{}
	err = getBaseConfig(cfg)
	if err != nil {
		return nil, err
	}

	var trueVal = true
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

	log.Debug("loadConfig", "service_port", cfg.Port)

	return cfg, nil
}

func getBaseConfig(cfg *config.AuthorityConfig) (err error) {
	if err = loadConfigState(&cfgState, DaemonName); err != nil {
		log.Error("error parsing", "config file", err)
		return
	}

	if cfg.DbConfig, err = getDbConfig(); err != nil {
		log.Error("getDbConfig error", err)
		return
	}

	cfg.Port = viper.GetInt(DaemonName + ".service_port")

	return
}

func loadConfigState(configState *ConfigState, prefix string) error {
	var filename string
	var err error

	filename = filepath.Base(configState.Config)
	viper.SetConfigName(filename[:len(filename)-len(filepath.Ext(filename))])
	viper.AddConfigPath(filepath.Dir(configState.Config))

	if err = viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file - %v", err)
	}

	configState.LogFile = viper.GetString(prefix + ".log_file")
	configState.LogHandler = viper.GetString(prefix + ".log_handler")
	configState.Verbosity = toVerbosity(viper.GetInt(prefix + ".verbosity"))

	return nil
}

func getDbConfig() (dbCfg config.DbConfig, err error) {
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
