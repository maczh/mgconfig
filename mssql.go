package config

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mssql"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/levigross/grequests"
	"strings"
)

var Mssql *gorm.DB

func mssqlInit() {
	if Mssql == nil {
		mssqlConfigUrl := getConfigUrl(conf.String("go.config.prefix.mssql"))
		resp, _ := grequests.Get(mssqlConfigUrl, nil)
		cfg := koanf.New(".")
		cfg.Load(rawbytes.Provider([]byte(resp.String())), yaml.Parser())
		dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
			cfg.String("go.data.postgres.user"),
			cfg.String("go.data.postgres.password"),
			cfg.String("go.data.postgres.host"),
			cfg.Int("go.data.postgres.port"),
			cfg.String("go.data.postgres.db"))
		Mssql, _ = gorm.Open("mssql", dsn)
		Mssql.LogMode(true)
		if strings.Contains(conf.String("go.config.used"), "redis") {
			Cacheable = true
		} else {
			Cacheable = false
		}
	}
}

func mssqlClose() {
	Mssql.Close()
	Mssql = nil
}

func mssqlCheck() *gorm.DB {
	_, err := Mssql.Rows()
	if err != nil {
		mssqlClose()
		mssqlInit()
	}
	return Mssql
}

func CheckMsSql() {
	mssqlInit()
}
