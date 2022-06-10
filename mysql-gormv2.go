package mgconfig

import (
	"errors"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/levigross/grequests"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

var mySQL *gorm.DB

func mySqlInit() {
	if mySQL == nil {
		mysqlConfigUrl := getConfigUrl(conf.String("go.config.prefix.mysql"))
		logger.Debug("正在获取MySQL配置: " + mysqlConfigUrl)
		resp, err := grequests.Get(mysqlConfigUrl, nil)
		if err != nil {
			logger.Error("MySQL配置下载失败! " + err.Error())
			return
		}
		cfg := koanf.New(".")
		cfg.Load(rawbytes.Provider([]byte(resp.String())), yaml.Parser())
		mySQL, _ = gorm.Open(mysql.Open(cfg.String("go.data.mysql")), &gorm.Config{})
		if cfg.Bool("go.data.mysql_debug") {
			mySQL = mySQL.Debug()
		}
		if cfg.Int("go.data.mysql_pool.max") > 1 {
			sqldb, _ := mySQL.DB()
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

func mySqlClose() {
	sqldb, _ := mySQL.DB()
	sqldb.Close()
	mySQL = nil
}

func mySqlCheck() (*gorm.DB, error) {
	if mySQL == nil {
		mySqlInit()
		if mySQL == nil {
			return mySQL, errors.New("mySQL connection error")
		}
	}
	sqldb, _ := mySQL.DB()
	err := sqldb.Ping()
	if err != nil {
		mySqlClose()
		mySqlInit()
		if mySQL == nil {
			return mySQL, errors.New("mySQL connection error")
		}
	}
	return mySQL, nil
}

func CheckMySQL() {
	_, err := mySqlCheck()
	if err != nil {
		logger.Error(err.Error())
	}
}

func GetMySQLConnection() (*gorm.DB, error) {
	return mySqlCheck()
}
