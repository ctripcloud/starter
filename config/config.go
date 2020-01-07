package config

import (
	"github.com/gokits/cfg"
	"github.com/gokits/cfg/source/file"
	"github.com/gokits/gotools"
	"github.com/gokits/stdlogger"
	validator "gopkg.in/go-playground/validator.v9"
)

// Secret defines sensitive configuration
type Secret struct {
}

// Logger configuration for logger
type Logger struct {
	RuntimePath       string `validate:"required"`
	RuntimeRemainDays int    `validate:"required"`
	AccessPath        string `validate:"required"`
	AccessRemainDays  int    `validate:"required"`
}

type DB struct {
	Driver       string `validate:"required,oneof=mysql sqlserver postgres sqlite3 oci8"`
	DataSource   string `validate:"required"`
	MaxIdleConns int    `validate:"min=0"`
	MaxOpenConns int    `validate:"min=1"`
	ShowSQL      bool
	ShowExecTime bool
}

// HTTPServer configuration for HttpServer
type HTTPServer struct {
	// ListenAddr addr to listen, ":8080" for example
	ListenAddr string `validate:"required"`
	// GraceShutdownPeriod time to wait before shutting down the server forcely
	GraceShutdownPeriod gotools.Duration `validate:"required"`
}

// Config defines configuration for perf-agent
type Config struct {
	Logger     Logger
	HTTPServer HTTPServer
	DB         DB
}

// PostDecode hook to validate new config. New config reload will be canceled if `error != nil`.
// `oldptr` will hold pointer to current version config struct
// `c` will hold pointer to new config struct
func (c *Config) PostDecode(oldptr interface{}) error {
	err := gvalidator.Struct(c)
	if err != nil {
		glogger.Errorf("Config not validate: %v", err)
	}
	return err
}

// PostSwap log out config
func (c *Config) PostSwap(oldptr interface{}) {
	glogger.Infof("Latest config: %+v", c)
}

// PostDecode hook to validate new config. New config reload will be canceled if `error != nil`.
// `oldptr` will hold pointer to current version config struct
// `c` will hold pointer to new config struct
func (c *Secret) PostDecode(oldptr interface{}) error {
	err := gvalidator.Struct(c)
	if err != nil {
		//warning do not log secret info
		glogger.Errorf("Secret not validate: %v", err)
	}
	return err
}

// PostSwap log out config
func (c *Secret) PostSwap(oldptr interface{}) {
	glogger.Infof("Latest Secret: %+v", c)
}

var (
	gvalidator   *validator.Validate
	cfgsource    *file.File
	secretsource *file.File
	glogger      stdlogger.LeveledLogger
	metacfg      *cfg.ConfigMeta
	metasecret   *cfg.ConfigMeta
)

// GetConfig returns latest snapshot of configuration, should be the only entrace to get config
func GetConfig() *Config {
	return metacfg.Get().(*Config)
}

// GetSecret returns latest snapshot of sensitive configuration, should be the only entrace to get config
func GetSecret() *Secret {
	return metasecret.Get().(*Secret)
}

// Init sets up hot reloader for config file, panic if fail
func Init(cfgpath, secretpath string, logger stdlogger.LeveledLogger) error {
	glogger = logger
	gvalidator = validator.New()
	var err error
	if cfgsource, err = file.NewFileSource(cfgpath, file.WithLogger(logger)); err != nil {
		return err
	}
	metacfg = cfg.NewConfigMeta(Config{}, cfgsource, cfg.WithLogger(logger))
	go metacfg.Run()
	logger.Infof("before config waitsynced")
	if err = metacfg.WaitSynced(); err != nil {
		Final()
		return err
	}
	logger.Infof("after config waitsynced")
	if secretpath != "" {
		if secretsource, err = file.NewFileSource(secretpath); err != nil {
			Final()
			return err
		}
		metasecret = cfg.NewConfigMeta(Secret{}, secretsource)
		go metasecret.Run()
		if err = metasecret.WaitSynced(); err != nil {
			Final()
			return err
		}
	}
	return nil
}

// Final stops reloader
func Final() {
	if metacfg != nil {
		metacfg.Stop()
	}
	if cfgsource != nil {
		cfgsource.Close()
	}
	if metasecret != nil {
		metasecret.Stop()
	}
	if secretsource != nil {
		secretsource.Close()
	}
}
