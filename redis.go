package mgconfig

import (
	"errors"
	"github.com/go-redis/redis"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/levigross/grequests"
	"github.com/silenceper/pool"
	"time"
)

var Redis *redis.Client
var redisPool pool.Pool

func redisInit() {
	if Redis == nil {
		redisConfigUrl := getConfigUrl(conf.String("go.config.prefix.redis"))
		resp, _ := grequests.Get(redisConfigUrl, nil)
		cfg := koanf.New(".")
		cfg.Load(rawbytes.Provider([]byte(resp.String())), yaml.Parser())
		ro := redis.Options{
			Addr:     cfg.String("go.data.redis.host") + ":" + cfg.String("go.data.redis.port"),
			Password: cfg.String("go.data.redis.password"),
			DB:       cfg.Int("go.data.redis.database"),
		}
		Redis = redis.NewClient(&ro)
		if err := Redis.Ping().Err(); err != nil {
			logger.Error("Redis连接失败:" + err.Error())
		}
		if cfg.Int("go.data.redis_pool.max") > 1 && (redisPool == nil || redisPool.Len() == 0) {
			factory := func() (interface{}, error) { return redisFactory(ro) }
			close := func(v interface{}) error { return v.(*redis.Client).Close() }
			ping := func(v interface{}) error { return v.(*redis.Client).Ping().Err() }
			min := cfg.Int("go.data.redis_pool.min")
			if min == 0 {
				min = 2
			}
			max := cfg.Int("go.data.redis_pool.max")
			if max < min {
				max = 10
			}
			idle := cfg.Int("go.data.redis_pool.idle")
			if idle == 0 || idle > max {
				idle = max / 2
			}
			idleTimeout := cfg.Int("go.data.redis_pool.timeout")
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
			redisPool, err = pool.NewChannelPool(poolConfig)
			if err != nil {
				logger.Error("Redis连接池初始化错误")
			}
		}
	}
}

func redisClose() {
	Redis.Close()
	Redis = nil
	if redisPool != nil && redisPool.Len() > 0 {
		redisPool.Release()
	}
}

func RedisCheck() {
	if Redis == nil {
		redisInit()
		return
	}
	if err := Redis.Ping().Err(); err != nil {
		logger.Error("Redis连接故障:" + err.Error())
		redisClose()
		redisInit()
	}
}

func GetRedisConnection() *redis.Client {
	if redisPool == nil {
		logger.Error("未初始化Redis连接池")
		return Redis
	}
	conn, err := redisPool.Get()
	if err != nil {
		logger.Error("获取Redis连接池中的连接失败:" + err.Error())
		return nil
	}
	if conn == nil {
		return nil
	}
	return conn.(*redis.Client)
}

func ReturnRedisConnection(conn *redis.Client) {
	if redisPool == nil || conn == nil {
		return
	}
	err := redisPool.Put(conn)
	if err != nil {
		logger.Error("归还Redis连接给连接池错误:" + err.Error())
	}
}

func redisFactory(ro redis.Options) (*redis.Client, error) {
	goredis := redis.NewClient(&ro)
	if goredis == nil {
		return nil, errors.New("Redis Connection failed")
	}
	return goredis, goredis.Ping().Err()
}
