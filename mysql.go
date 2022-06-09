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
			min := cfg.Int("go.data.mysql_pool.min")
			if min == 0 {
				min = 2
			}
			Mysql.DB().SetMaxIdleConns(min)
			max := cfg.Int("go.data.mysql_pool.max")
			if max < 10 {
				max = 10
			}
			Mysql.DB().SetMaxOpenConns(max)
			//idle := cfg.Int("go.data.mysql_pool.idle")
			//if idle == 0 || idle > max {
			//	idle = max / 2
			//}
			idleTimeout := cfg.Int("go.data.mysql_pool.timeout")
			if idleTimeout == 0 {
				idleTimeout = 60
			}
			Mysql.DB().SetConnMaxIdleTime(time.Duration(idleTimeout) * time.Second)
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
