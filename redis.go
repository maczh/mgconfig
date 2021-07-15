package mgconfig

import (
	"github.com/go-redis/redis"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/levigross/grequests"
)

var Redis *redis.Client

func redisInit() {
	if Redis == nil {
		redisConfigUrl := getConfigUrl(conf.String("go.config.prefix.redis"))
		resp, _ := grequests.Get(redisConfigUrl, nil)
		cfg := koanf.New(".")
		cfg.Load(rawbytes.Provider([]byte(resp.String())), yaml.Parser())
		Redis = redis.NewClient(&redis.Options{
			Addr:     cfg.String("go.data.redis.host") + ":" + cfg.String("go.data.redis.port"),
			Password: cfg.String("go.data.redis.password"),
			DB:       cfg.Int("go.data.redis.database"),
		})
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
