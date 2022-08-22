package mgconfig

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/levigross/grequests"
	"net"
	"strings"
	"time"
)

var redisClient *redis.Client

var multiRedis bool
var redisClients = make(map[string]*redis.Client)
var redisCfgs = make(map[string]*redis.Options)

func redisInit() {
	if redisClient == nil && len(redisClients) == 0 {
		redisConfigUrl := getConfigUrl(conf.String("go.config.prefix.redis"))
		logger.Debug("正在获取Redis配置: " + redisConfigUrl)
		resp, err := grequests.Get(redisConfigUrl, nil)
		if err != nil {
			logger.Error("Redis配置下载失败! " + err.Error())
			return
		}
		cfg := koanf.New(".")
		cfg.Load(rawbytes.Provider([]byte(resp.String())), yaml.Parser())
		multiRedis = cfg.Bool("go.data.redis.multidb")
		fmt.Printf("multiRedis = %v\n", multiRedis)
		var ro redis.Options
		if multiRedis {
			dbNames := strings.Split(cfg.String("go.data.redis.dbNames"), ",")
			for _, dbName := range dbNames {
				if dbName != "" && cfg.Exists(fmt.Sprintf("go.data.redis.%s.host", dbName)) {
					ropt := redis.Options{
						Addr:     cfg.String(fmt.Sprintf("go.data.redis.%s.host", dbName)) + ":" + cfg.String(fmt.Sprintf("go.data.redis.%s.port", dbName)),
						Password: cfg.String(fmt.Sprintf("go.data.redis.%s.password", dbName)),
						DB:       cfg.Int(fmt.Sprintf("go.data.redis.%s.database", dbName)),
						Dialer: func() (net.Conn, error) {
							netDialer := &net.Dialer{
								Timeout:   5 * time.Second,
								KeepAlive: 5 * time.Minute,
							}
							return netDialer.Dial("tcp", cfg.String(fmt.Sprintf("go.data.redis.%s.host", dbName))+":"+cfg.String(fmt.Sprintf("go.data.redis.%s.port", dbName)))
						},
					}
					redisCfgs[dbName] = &ropt
				}
			}
			fmt.Printf("opts : %v\n", redisCfgs)
		} else {
			ro = redis.Options{
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
		}
		if cfg.Int("go.data.redis_pool.max") > 1 {
			min := cfg.Int("go.data.redis_pool.min")
			if min == 0 {
				min = 2
			}
			max := cfg.Int("go.data.redis_pool.max")
			if max < 10 {
				max = 10
			}
			idleTimeout := cfg.Int("go.data.redis_pool.idleTimeout")
			if idleTimeout == 0 {
				idleTimeout = 5
			}
			connectTimeout := cfg.Int("go.data.redis_pool.timeout")
			if connectTimeout == 0 {
				connectTimeout = 60
			}
			if multiRedis {
				for k, r := range redisCfgs {
					r.PoolSize = max
					r.MinIdleConns = min
					r.IdleTimeout = time.Duration(idleTimeout) * time.Minute
					r.DialTimeout = time.Duration(connectTimeout) * time.Second
					redisCfgs[k] = r
				}
			} else {
				ro.PoolSize = max
				ro.MinIdleConns = min
				ro.IdleTimeout = time.Duration(idleTimeout) * time.Minute
				ro.DialTimeout = time.Duration(connectTimeout) * time.Second
			}
		}
		if multiRedis {
			for dbName, r := range redisCfgs {
				fmt.Printf("正在连接%s,参数\v\n", dbName, *r)
				rc := redis.NewClient(r)
				if err := rc.Ping().Err(); err != nil {
					logger.Error(dbName + " Redis连接失败:" + err.Error())
					continue
				}
				fmt.Printf("%s 连接成功\n", dbName)
				redisClients[dbName] = rc
			}
		} else {
			redisClient = redis.NewClient(&ro)
			if err := redisClient.Ping().Err(); err != nil {
				logger.Error("Redis连接失败:" + err.Error())
			}
		}
	}
}

func redisClose() {
	if multiRedis {
		for dbName, rc := range redisClients {
			rc.Close()
			delete(redisClients, dbName)
		}
	} else {
		redisClient.Close()
		redisClient = nil
	}
}

func redisCheck(dbName string) error {
	fmt.Printf("正在检查%s连接\n", dbName)
	if err := redisClients[dbName].Ping().Err(); err != nil {
		logger.Error("Redis连接故障:" + err.Error())
		ropt := redisCfgs[dbName]
		rc := redis.NewClient(ropt)
		if err := rc.Ping().Err(); err != nil {
			logger.Error(dbName + " Redis连接失败:" + err.Error())
			return err
		}
		redisClients[dbName] = rc
	}
	return nil
}

func RedisCheck() {
	if redisClient == nil && len(redisClients) == 0 {
		redisInit()
		return
	}
	if multiRedis {
		fmt.Printf("redisClients: %v\n", redisClients)
		for dbName, _ := range redisCfgs {
			_ = redisCheck(dbName)
		}
	} else {
		if err := redisClient.Ping().Err(); err != nil {
			logger.Error("Redis连接故障:" + err.Error())
			redisClose()
			redisInit()
		}
	}
}

func GetRedisConnection(dbName ...string) (*redis.Client, error) {
	if multiRedis {
		if len(dbName) == 0 || len(dbName) > 1 {
			return nil, errors.New("Multidb Get Redis connection must specify one database name")
		}
		err := redisCheck(dbName[0])
		if err != nil {
			return nil, err
		}
		return redisClients[dbName[0]], nil
	} else {
		RedisCheck()
		if redisClient == nil {
			return nil, errors.New("redis connection failed")
		}
		return redisClient, nil
	}
}
