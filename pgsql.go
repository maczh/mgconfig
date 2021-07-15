package mgconfig

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/levigross/grequests"
	"strings"
)

var Postgres *gorm.DB

func pgsqlInit() {
	if Postgres == nil {
		pgsqlConfigUrl := getConfigUrl(conf.String("go.config.prefix.pgsql"))
		resp, _ := grequests.Get(pgsqlConfigUrl, nil)
		cfg := koanf.New(".")
		cfg.Load(rawbytes.Provider([]byte(resp.String())), yaml.Parser())
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
			cfg.String("go.data.postgres.host"),
			cfg.Int("go.data.postgres.port"),
			cfg.String("go.data.postgres.user"),
			cfg.String("go.data.postgres.password"),
			cfg.String("go.data.postgres.db"))
		Postgres, _ = gorm.Open("postgres", dsn)
		Postgres.LogMode(true)
		if strings.Contains(conf.String("go.config.used"), "redis") {
			Cacheable = true
		} else {
			Cacheable = false
		}
	}
}

func pgsqlClose() {
	Postgres.Close()
	Postgres = nil
}

func pgsqlCheck() *gorm.DB {
	_, err := Postgres.Rows()
	if err != nil {
		pgsqlClose()
		pgsqlInit()
	}
	return Postgres
}

func CheckPqsql() {
	pgsqlInit()
}
