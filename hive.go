package mgconfig

import (
	"github.com/beltran/gohive"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/levigross/grequests"
)

var Hive *gohive.Connection

func hiveInit() {
	if Hive == nil {
		hbaseConfigUrl := getConfigUrl(conf.String("go.config.prefix.hive"))
		resp, _ := grequests.Get(hbaseConfigUrl, nil)
		cfg := koanf.New(".")
		cfg.Load(rawbytes.Provider([]byte(resp.String())), yaml.Parser())
		host := cfg.String("go.data.hive.host")
		port := cfg.Int("go.data.hive.port")
		auth := cfg.String("go.data.hive.auth")
		user := cfg.String("go.data.hive.user")
		password := cfg.String("go.data.hive.password")
		db := cfg.String("go.data.hive.database")
		service := cfg.String("go.data.hive.service")
		transport := cfg.String("go.data.hive.transport")
		config := &gohive.ConnectConfiguration{Username: user,
			Password:             password,
			Service:              service,
			TransportMode:        transport,
			Database:             db,
			PollIntervalInMillis: 200,
			HiveConfiguration:    nil,
			FetchSize:            gohive.DEFAULT_FETCH_SIZE,
			TLSConfig:            nil,
			HTTPPath:             "cliservice",
			ZookeeperNamespace:   gohive.ZOOKEEPER_DEFAULT_NAMESPACE,
		}
		var err error
		Hive, err = gohive.Connect(host, port, auth, config)
		if err != nil {
			logger.Error("Hive连接错误:" + err.Error())
			Hive = nil
			return
		} else {
			logger.Info("Hive连接成功")
		}
	}
}

func hiveClose() {
	if Hive != nil {
		err := Hive.Close()
		if err != nil {
			logger.Error("Hive断开链接错误:" + err.Error())
		}
		Hive = nil
	}
}
