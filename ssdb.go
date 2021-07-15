package config

import (
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/levigross/grequests"
	"github.com/seefan/gossdb"
	ssdbconf "github.com/seefan/gossdb/conf"
	"github.com/seefan/gossdb/pool"
)

var Ssdb *pool.Client

func ssdbInit() {
	if Ssdb == nil {
		ssdbConfigUrl := getConfigUrl(conf.String("go.config.prefix.ssdb"))
		resp, _ := grequests.Get(ssdbConfigUrl, nil)
		cfg := koanf.New(".")
		cfg.Load(rawbytes.Provider([]byte(resp.String())), yaml.Parser())
		err := gossdb.Start(&ssdbconf.Config{
			Host:           cfg.String("go.data.ssdb.host"),
			Port:           cfg.Int("go.data.ssdb.port"),
			Password:       cfg.String("go.data.ssdb.password"),
			ConnectTimeout: cfg.Int("go.data.ssdb.timeout"),
		})
		if err != nil {
			Ssdb = nil
			logger.Error("SSDB连接错误:" + err.Error())
		} else {
			Ssdb, err = gossdb.NewClient()
			if err != nil {
				logger.Error("SSDB连接错误:" + err.Error())
				Ssdb = nil
			}
		}
	}
}

func ssdbClose() {
	Ssdb.Close()
	Ssdb = nil
}

func ssdbCheck() {
	if !Ssdb.Ping() {
		ssdbClose()
		ssdbInit()
	}
}
