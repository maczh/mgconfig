package mgconfig

import (
	"errors"
	"fmt"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/levigross/grequests"
	"gopkg.in/mgo.v2"
	"log"
	"os"
	"strings"
)

var conn *mgo.Database
var mongo *mgo.Session
var mgodb string
var multimgo bool = false
var mongos = make(map[string]*mgo.Session)
var mgoDbNames = make(map[string]string)
var mongoUrls = make(map[string]string)
var max int

func mgoInit() {
	if conn == nil && len(mongos) == 0 {
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
		multimgo = cfg.Bool("go.data.mongodb.multidb")
		if multimgo {
			dbNames := strings.Split(cfg.String("go.data.mongodb.dbNames"), ",")
			for _, dbName := range dbNames {
				if dbName != "" && cfg.Exists(fmt.Sprintf("go.data.mongodb.%s.uri", dbName)) {
					mongoUrls[dbName] = cfg.String(fmt.Sprintf("go.data.mongodb.%s.uri", dbName))
					session, err := mgo.Dial(mongoUrls[dbName])
					if err != nil {
						logger.Error(dbName + " MongoDB连接错误:" + err.Error())
						continue
					}
					mongos[dbName] = session
					mgoDbNames[dbName] = cfg.String(fmt.Sprintf("go.data.mongodb.%s.db", dbName))
					if cfg.Int("go.data.mongo_pool.max") > 1 {
						max = cfg.Int("go.data.mongo_pool.max")
						if max < 10 {
							max = 10
						}
						session.SetPoolLimit(max)
						session.SetMode(mgo.Monotonic, true)
					}
				}
			}
		} else {
			mongo, err = mgo.Dial(cfg.String("go.data.mongodb.uri"))
			if err != nil {
				logger.Error("MongoDB连接错误:" + err.Error())
				return
			}
			if cfg.Int("go.data.mongo_pool.max") > 1 {
				max = cfg.Int("go.data.mongo_pool.max")
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
}

func mgoClose() {
	if multimgo {
		for k, _ := range mongos {
			mongos[k].Close()
			delete(mongos, k)
		}
	} else {
		mongo.Close()
		conn = nil
		mongo = nil
	}
}

func mgoCheck(dbName string) error {
	if mongos[dbName].Ping() != nil {
		mongos[dbName].Close()
		session, err := mgo.Dial(mongoUrls[dbName])
		if err != nil {
			logger.Error(dbName + " MongoDB连接错误:" + err.Error())
			return err
		}
		mongos[dbName] = session
		session.SetPoolLimit(max)
		session.SetMode(mgo.Monotonic, true)
	}
	return nil
}

func MgoCheck() {
	if (conn == nil || mongo == nil) && len(mongos) == 0 {
		mgoInit()
		return
	}
	if multimgo {
		for dbName, _ := range mongos {
			err := mgoCheck(dbName)
			if err != nil {
				continue
			}
		}
	} else {
		if mongo.Ping() != nil {
			mgoClose()
			mgoInit()
		}
	}
}

func GetMongoConnection(dbName ...string) (*mgo.Database, error) {
	if multimgo {
		if len(dbName) > 1 || len(dbName) == 0 {
			return nil, errors.New("Multidb MongoDB get connection must be specified one dbName")
		}
		err := mgoCheck(dbName[0])
		if err != nil {
			return nil, err
		}
		return mongos[dbName[0]].Copy().DB(mgoDbNames[dbName[0]]), nil
	} else {
		MgoCheck()
		if mongo == nil {
			return nil, errors.New("MongoDB connection failed")
		}
		return mongo.Copy().DB(mgodb), nil
	}
}

func ReturnMongoConnection(conn *mgo.Database) {
	conn.Session.Close()
}
