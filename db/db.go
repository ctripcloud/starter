package db

import (
	"github.com/ctripcloud/starter/config"
	"github.com/go-xorm/xorm"
	// _ "github.com/go-sql-driver/mysql"
)

var (
	Engine *xorm.Engine
)

func Init() error {
	var err error
	cfg := config.GetConfig()
	Engine, err = xorm.NewEngine(cfg.DB.Driver, cfg.DB.DataSource)
	if err != nil {
		return err
	}

	Engine.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	Engine.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	Engine.ShowSQL(cfg.DB.ShowSQL)
	Engine.ShowExecTime(cfg.DB.ShowExecTime)
	return nil
}

func Final() {
	Engine.Close()
	Engine = nil
}
