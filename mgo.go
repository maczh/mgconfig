package mgconfig

import (
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/levigross/grequests"
	"github.com/silenceper/pool"
	"gopkg.in/mgo.v2"
	"time"
)

var Mgo *mgo.Database
var mgoPool pool.Pool

func mgoInit() {
	if Mgo == nil {
		mongodbConfigUrl := getConfigUrl(conf.String("go.config.prefix.mongodb"))
		resp, _ := grequests.Get(mongodbConfigUrl, nil)
		cfg := koanf.New(".")
		cfg.Load(rawbytes.Provider([]byte(resp.String())), yaml.Parser())
		//mgo.SetDebug(true)
		//var mgoLogger *log.Logger
		//mgoLogger = log.New(os.Stderr, "", log.LstdFlags)
		//mgo.SetLogger(mgoLogger)
		mongo, err := mgo.Dial(cfg.String("go.data.mongodb.uri"))
		if err != nil {
			logger.Error("MongoDB连接错误:" + err.Error())
		} else {
			Mgo = mongo.DB(cfg.String("go.data.mongodb.db"))
		}
		if cfg.Int("go.data.mongo_pool.max") > 1 && (mgoPool == nil || mgoPool.Len() == 0) {
			factory := func() (interface{}, error) { return mgo.Dial(cfg.String("go.data.mongodb.uri")) }
			close := func(v interface{}) error { v.(*mgo.Database).Session.Close(); return nil }
			ping := func(v interface{}) error { return v.(*mgo.Database).Session.Ping() }
			min := cfg.Int("go.data.mongo_pool.min")
			if min == 0 {
				min = 2
			}
			max := cfg.Int("go.data.mongo_pool.max")
			if max < min {
				max = 10
			}
			idle := cfg.Int("go.data.mongo_pool.idle")
			if idle == 0 || idle > max {
				idle = max / 2
			}
			idleTimeout := cfg.Int("go.data.mongo_pool.timeout")
			if idleTimeout == 0 {
				idleTimeout = 60
			}
			poolConfig := &pool.Config{
				InitialCap:  min,
				MaxCap:      max,
				MaxIdle:     idle,
				Factory:     factory,
				Close:       close,
				Ping:        ping,
				IdleTimeout: time.Duration(idleTimeout) * time.Second,
			}
			var err error
			mgoPool, err = pool.NewChannelPool(poolConfig)
			if err != nil {
				logger.Error("MongoDB连接池初始化错误")
			}
		}
	}
}

func mgoClose() {
	Mgo.Session.Close()
	Mgo = nil
	if mgoPool != nil && mgoPool.Len() > 0 {
		mgoPool.Release()
	}
}

func MgoCheck() {
	if Mgo == nil {
		mgoInit()
		return
	}
	if Mgo.Session.Ping() != nil {
		mgoClose()
		mgoInit()
	}
}

func GetMongoConnection() *mgo.Database {
	if mgoPool == nil || mgoPool.Len() == 0 {
		logger.Error("MongoDB连接池未初始化")
		return Mgo
	}
	conn, err := mgoPool.Get()
	if err != nil {
		logger.Error("MongoDB连接池错误:" + err.Error())
		return nil
	}
	if conn == nil {
		return nil
	}
	return conn.(*mgo.Database)
}

func ReturnMongoConnection(conn *mgo.Database) {
	if mgoPool == nil || mgoPool.Len() == 0 || conn == nil {
		return
	}
	err := mgoPool.Put(conn)
	if err != nil {
		logger.Error("归还MongoDB连接给连接池错误:" + err.Error())
	}
}
