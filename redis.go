package mgconfig

import (
	"errors"
	"github.com/go-redis/redis"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/levigross/grequests"
	"net"
	"time"
)

var Redis *redis.Client

func redisInit() {
	if Redis == nil {
		redisConfigUrl := getConfigUrl(conf.String("go.config.prefix.redis"))
		logger.Debug("正在获取Redis配置: " + redisConfigUrl)
		resp, err := grequests.Get(redisConfigUrl, nil)
		if err != nil {
			logger.Error("Redis配置下载失败! " + err.Error())
			return
		}
		cfg := koanf.New(".")
		cfg.Load(rawbytes.Provider([]byte(resp.String())), yaml.Parser())
		ro := redis.Options{
			Addr:     cfg.String("go.data.redis.host") + ":" + cfg.String("go.data.redis.port"),
			Password: cfg.String("go.data.redis.password"),
			DB:       cfg.Int("go.data.redis.database"),
			Dialer: func() (net.Conn, error) {
				netDialer := &net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 5 * time.Minute,
				}
				return netDialer.Dial("tcp", cfg.String("go.data.redis.host")+":"+cfg.String("go.data.redis.port"))
			},
		}
		if cfg.Int("go.data.redis_pool.max") > 1 {
			min := cfg.Int("go.data.redis_pool.min")
			if min == 0 {
				min = 2
			}
			ro.MinIdleConns = min
			max := cfg.Int("go.data.redis_pool.max")
			if max < min {
				max = 10
			}
			ro.PoolSize = max
			idleTimeout := cfg.Int("go.data.redis_pool.idleTimeout")
			if idleTimeout == 0 {
				idleTimeout = 5
			}
			ro.IdleTimeout = time.Duration(idleTimeout) * time.Minute
			connectTimeout := cfg.Int("go.data.redis_pool.timeout")
			if connectTimeout == 0 {
				connectTimeout = 60
			}
			ro.DialTimeout = time.Duration(connectTimeout) * time.Second
		}
		Redis = redis.NewClient(&ro)
		if err := Redis.Ping().Err(); err != nil {
			logger.Error("Redis连接失败:" + err.Error())
		}
	}
}

func redisClose() {
	Redis.Close()
	Redis = nil
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

func GetRedisConnection() (*redis.Client, error) {
	RedisCheck()
	if Redis == nil {
		return nil, errors.New("redis connection failed")
	}
	return Redis, nil
}

func ReturnRedisConnection(conn *redis.Client) {
}
