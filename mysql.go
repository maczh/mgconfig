package mgconfig

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/levigross/grequests"
	"github.com/silenceper/pool"
	"strings"
	"time"
)

var Mysql *gorm.DB
var Cacheable bool
var mysqlPool pool.Pool

func mysqlInit() {
	if Mysql == nil {
		mysqlConfigUrl := getConfigUrl(conf.String("go.config.prefix.mysql"))
		resp, _ := grequests.Get(mysqlConfigUrl, nil)
		cfg := koanf.New(".")
		cfg.Load(rawbytes.Provider([]byte(resp.String())), yaml.Parser())
		Mysql, _ = gorm.Open("mysql", cfg.String("go.data.mysql"))
		Mysql.LogMode(true)
		if strings.Contains(conf.String("go.config.used"), "redis") {
			Cacheable = true
		} else {
			Cacheable = false
		}
		if cfg.Int("go.data.mysql_pool.max") > 1 && (mysqlPool == nil || mysqlPool.Len() == 0) {
			factory := func() (interface{}, error) { return gorm.Open("mysql", cfg.String("go.data.mysql")) }
			close := func(v interface{}) error { return v.(*gorm.DB).Close() }
			min := cfg.Int("go.data.mysql_pool.min")
			if min == 0 {
				min = 2
			}
			max := cfg.Int("go.data.mysql_pool.max")
			if max < min {
				max = 10
			}
			idle := cfg.Int("go.data.mysql_pool.idle")
			if idle == 0 || idle > max {
				idle = max / 2
			}
			idleTimeout := cfg.Int("go.data.mysql_pool.timeout")
			if idleTimeout == 0 {
				idleTimeout = 60
			}
			poolConfig := &pool.Config{
				InitialCap:  min,
				MaxCap:      max,
				MaxIdle:     idle,
				Factory:     factory,
				Close:       close,
				IdleTimeout: time.Duration(idleTimeout) * time.Second,
			}
			var err error
			mysqlPool, err = pool.NewChannelPool(poolConfig)
			if err != nil {
				logger.Error("MySQL连接池初始化错误")
			}
		}
	}
}

func mysqlClose() {
	Mysql.Close()
	Mysql = nil
	if mysqlPool != nil && mysqlPool.Len() > 0 {
		mysqlPool.Release()
	}
}

func mysqlCheck() *gorm.DB {
	_, err := Mysql.Rows()
	if err != nil {
		mysqlClose()
		mysqlInit()
	}
	return Mysql
}

func CheckMySql() {
	mysqlInit()
}

func GetMysqlConnection() *gorm.DB {
	if mysqlPool == nil || mysqlPool.Len() == 0 {
		logger.Error("未初始化MySQL连接池")
		return Mysql
	}
	conn, err := mysqlPool.Get()
	if err != nil {
		logger.Error("获取MySQL连接池中的连接失败:" + err.Error())
		return nil
	}
	if conn == nil {
		return nil
	}
	return conn.(*gorm.DB)
}

func ReturnMysqlConnection(conn *gorm.DB) {
	if mysqlPool == nil || mysqlPool.Len() == 0 || conn == nil {
		return
	}
	err := mysqlPool.Put(conn)
	if err != nil {
		logger.Error("归还MySQL连接给连接池错误:" + err.Error())
	}
}
