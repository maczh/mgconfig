package mgconfig

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/levigross/grequests"
	"time"
)

var Mysql *gorm.DB

func mysqlInit() {
	if Mysql == nil {
		mysqlConfigUrl := getConfigUrl(conf.String("go.config.prefix.mysql"))
		logger.Debug("正在获取MySQL配置: " + mysqlConfigUrl)
		resp, err := grequests.Get(mysqlConfigUrl, nil)
		if err != nil {
			logger.Error("MySQL配置下载失败! " + err.Error())
			return
		}
		cfg := koanf.New(".")
		cfg.Load(rawbytes.Provider([]byte(resp.String())), yaml.Parser())
		Mysql, _ = gorm.Open("mysql", cfg.String("go.data.mysql"))
		if cfg.Bool("go.data.mysql_debug") {
			Mysql = Mysql.Debug()
		}
		Mysql.LogMode(true)
		if cfg.Int("go.data.mysql_pool.max") > 1 {
			if cfg.Int("go.data.mysql_pool.max") > 1 {
				sqldb := Mysql.DB()
				max := cfg.Int("go.data.mysql_pool.max")
				if max < 10 {
					max = 10
				}
				sqldb.SetMaxOpenConns(max)
				idle := cfg.Int("go.data.mysql_pool.total")
				if idle == 0 || idle < max {
					idle = 5 * max
				}
				sqldb.SetMaxIdleConns(idle)
				idleTimeout := cfg.Int("go.data.mysql_pool.timeout")
				if idleTimeout == 0 {
					idleTimeout = 60
				}
				sqldb.SetConnMaxIdleTime(time.Duration(idleTimeout) * time.Second)
				lifetime := cfg.Int("go.data.mysql_pool.life")
				if lifetime == 0 {
					lifetime = 60
				}
				sqldb.SetConnMaxLifetime(time.Duration(lifetime) * time.Minute)
			}
		}
	}
}

func mysqlClose() {
	Mysql.Close()
	Mysql = nil
}

func mysqlCheck() *gorm.DB {
	if Mysql == nil {
		mysqlInit()
		return Mysql
	}
	err := Mysql.DB().Ping()
	if err != nil {
		mysqlClose()
		mysqlInit()
	}
	return Mysql
}

func CheckMySql() {
	mysqlCheck()
}

func GetMysqlConnection() *gorm.DB {
	return mysqlCheck()
}

func ReturnMysqlConnection(conn *gorm.DB) {
}
