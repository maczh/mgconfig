package mgconfig

import (
	"errors"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/levigross/grequests"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
)

var mySQL *gorm.DB
var mysqls = make(map[string]*gorm.DB)
var multidb bool

func mySqlInit() {
	if mySQL == nil && len(mysqls) == 0 {
		mysqlConfigUrl := getConfigUrl(conf.String("go.config.prefix.mysql"))
		logger.Debug("正在获取MySQL配置: " + mysqlConfigUrl)
		resp, err := grequests.Get(mysqlConfigUrl, nil)
		if err != nil {
			logger.Error("MySQL配置下载失败! " + err.Error())
			return
		}
		cfg := koanf.New(".")
		cfg.Load(rawbytes.Provider([]byte(resp.String())), yaml.Parser())
		multidb = false
		if cfg.Exists("go.data.mysql.multidb") && cfg.Bool("go.data.mysql.multidb") {
			multidb = true
			dbNames := strings.Split(cfg.String("go.data.mysql.dbNames"), ",")
			for _, dbName := range dbNames {
				if dbName != "" && cfg.String("go.data.mysql."+dbName) != "" {
					conn, err := gorm.Open(mysql.Open(cfg.String("go.data.mysql."+dbName)), &gorm.Config{})
					if err != nil {
						logger.Error(dbName + " mysql connection error:" + err.Error())
						continue
					}
					mysqls[dbName] = conn
				}
			}
		} else {
			mySQL, _ = gorm.Open(mysql.Open(cfg.String("go.data.mysql")), &gorm.Config{})
		}
		if cfg.Bool("go.data.mysql_debug") {
			if multidb {
				for k, _ := range mysqls {
					mysqls[k] = mysqls[k].Debug()
				}
			} else {
				mySQL = mySQL.Debug()
			}
		}
		if cfg.Int("go.data.mysql_pool.max") > 1 {
			max := cfg.Int("go.data.mysql_pool.max")
			if max < 10 {
				max = 10
			}
			idle := cfg.Int("go.data.mysql_pool.total")
			if idle == 0 || idle < max {
				idle = 5 * max
			}
			idleTimeout := cfg.Int("go.data.mysql_pool.timeout")
			if idleTimeout == 0 {
				idleTimeout = 60
			}
			lifetime := cfg.Int("go.data.mysql_pool.life")
			if lifetime == 0 {
				lifetime = 60
			}
			if !multidb {
				sqldb, _ := mySQL.DB()
				sqldb.SetConnMaxIdleTime(time.Duration(idleTimeout) * time.Second)
				sqldb.SetMaxIdleConns(idle)
				sqldb.SetMaxOpenConns(max)
				sqldb.SetConnMaxLifetime(time.Duration(lifetime) * time.Minute)
			} else {
				for k, _ := range mysqls {
					sqldb, _ := mysqls[k].DB()
					sqldb.SetConnMaxIdleTime(time.Duration(idleTimeout) * time.Second)
					sqldb.SetMaxIdleConns(idle)
					sqldb.SetMaxOpenConns(max)
					sqldb.SetConnMaxLifetime(time.Duration(lifetime) * time.Minute)
				}
			}
		}
	}
}

func mySqlClose() {
	if multidb {
		for k, _ := range mysqls {
			sqldb, _ := mysqls[k].DB()
			sqldb.Close()
			delete(mysqls, k)
		}
	} else {
		sqldb, _ := mySQL.DB()
		sqldb.Close()
		mySQL = nil
	}
}

func mySqlsCheck() error {
	if !multidb {
		return errors.New("Not multidb mysql connections setting")
	}
	if len(mysqls) == 0 {
		mySqlInit()
		if len(mysqls) == 0 {
			return errors.New("mySQL connection error")
		}
	}
	for k, _ := range mysqls {
		sqldb, _ := mysqls[k].DB()
		err := sqldb.Ping()
		if err != nil {
			mySqlClose()
			mySqlInit()
			if len(mysqls) == 0 {
				return errors.New("mySQL connection error")
			}
		}
	}
	return nil
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
	if multidb {
		err := mySqlsCheck()
		if err != nil {
			logger.Error(err.Error())
		}
	} else {
		_, err := mySqlCheck()
		if err != nil {
			logger.Error(err.Error())
		}
	}
}

func GetMySQLConnection(dbName ...string) (*gorm.DB, error) {
	if len(dbName) == 0 {
		if multidb {
			return nil, errors.New("multidb get connection must specify a database name")
		}
		return mySqlCheck()
	}
	if len(dbName) > 1 {
		return nil, errors.New("Multidb can only get one connection")
	}
	if !multidb {
		return mySqlCheck()
	}
	conn := mysqls[dbName[0]]
	if conn == nil {
		return nil, errors.New(dbName[0] + " mysql connection not found or failed")
	}
	return conn, nil
}
