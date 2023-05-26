package main

import (
	"encoding/json"
	"fmt"
	"fulfillmentd/fulfillmentd"
	"fulfillmentd/server"
	"fulfillmentd/server/config"
	"github.com/eluv-io/errors-go"
	elog "github.com/eluv-io/log-go"
	"github.com/eluv-io/log-go/handlers/console"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
	"io/ioutil"
	"net/http"
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
	ElvSection = "elv"
	Main       = "main"
	Demov3     = "demov3"
)

var (
	cfgState = ConfigState{}
	log      = elog.Get("/fs")
)

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s --config <config.toml>\n", DaemonName)
		return
	}

	if _, err := startServer(os.Args[2]); err != nil {
		fmt.Println("cannot launch", "Error", err)
		os.Exit(1)
	}
}

func startServer(configFile string) (s *server.Server, err error) {
	cfg, err := loadConfig(configFile)
	if err != nil {
		return
	}

	s, err = server.ConnectDb(cfg)
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
	viper.SetDefault(ElvSection+".networks", map[string]string{
		Main:   "https://main.net955305.contentfabric.io/config",
		Demov3: "https://demov3.net955210.contentfabric.io/config",
	})

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

	gin.DefaultWriter = &lumberjack.Logger{
		Filename:  cfgState.LogFile,
		LocalTime: false,
	}

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

	nets := viper.GetStringMapString("elv.networks")
	log.Info("network configs", "nets", nets)

	cfg.EthUrlsByNetwork = make(map[string]string)
	for net, url := range nets {
		var ethUrl string
		ethUrl, err = getEthUrlFromConfigUrl(url)
		if err != nil {
			return
		}
		cfg.EthUrlsByNetwork[net] = ethUrl
	}
	log.Info("eth endpoints", "url-map", cfg.EthUrlsByNetwork)

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
	configState.Verbosity = viper.GetString(prefix + ".verbosity")

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

// getEthUrlFromConfigUrl loads the fabric config url js data and then pulls out the first eth endpoint in it.
func getEthUrlFromConfigUrl(configUrl string) (ethUrl string, err error) {
	var resp *http.Response
	var body []byte
	var js map[string]interface{}

	if resp, err = http.Get(configUrl); err != nil {
		return
	}
	defer resp.Body.Close()

	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &js); err != nil {
		return
	}

	if js["network"] == nil {
		err = errors.NoTrace("no network in config")
		return
	}

	if js["network"].(map[string]interface{})["services"] == nil {
		err = errors.NoTrace("no services in config")
		return
	}

	if js["network"].(map[string]interface{})["services"].(map[string]interface{})["ethereum_api"] == nil {
		err = errors.NoTrace("no ethereum_api in config")
		return
	}

	ethUrl = js["network"].(map[string]interface{})["services"].(map[string]interface{})["ethereum_api"].([]interface{})[0].(string)

	return
}
