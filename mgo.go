package mgconfig

import (
	"errors"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/levigross/grequests"
	"gopkg.in/mgo.v2"
	"log"
	"os"
)

var conn *mgo.Database
var mongo *mgo.Session
var mgodb string

func mgoInit() {
	if conn == nil {
		mongodbConfigUrl := getConfigUrl(conf.String("go.config.prefix.mongodb"))
		logger.Debug("正在获取MongoDB配置: " + mongodbConfigUrl)
		resp, err := grequests.Get(mongodbConfigUrl, nil)
		if err != nil {
			logger.Error("MongoDB配置下载失败! " + err.Error())
			return
		}
		cfg := koanf.New(".")
		cfg.Load(rawbytes.Provider([]byte(resp.String())), yaml.Parser())
		if cfg.Bool("go.data.mongodb.debug") {
			mgo.SetDebug(true)
			var mgoLogger *log.Logger
			mgoLogger = log.New(os.Stderr, "", log.LstdFlags)
			mgo.SetLogger(mgoLogger)
		}
		mongo, err = mgo.Dial(cfg.String("go.data.mongodb.uri"))
		if err != nil {
			logger.Error("MongoDB连接错误:" + err.Error())
			return
		}
		if cfg.Int("go.data.mongo_pool.max") > 1 {
			max := cfg.Int("go.data.mongo_pool.max")
			if max < 10 {
				max = 10
			}
			mongo.SetPoolLimit(max)
			mongo.SetMode(mgo.Monotonic, true)
		}
		mgodb = cfg.String("go.data.mongodb.db")
		conn = mongo.Copy().DB(mgodb)
	}
}

func mgoClose() {
	mongo.Close()
	conn = nil
	mongo = nil
}

func MgoCheck() {
	if conn == nil || mongo == nil {
		mgoInit()
		return
	}
	if mongo.Ping() != nil {
		mgoClose()
		mgoInit()
	}
}

func GetMongoConnection() (*mgo.Database, error) {
	MgoCheck()
	if mongo == nil {
		return nil, errors.New("MongoDB connection failed")
	}
	return mongo.Copy().DB(mgodb), nil
}

func ReturnMongoConnection(conn *mgo.Database) {
	conn.Session.Close()
}
