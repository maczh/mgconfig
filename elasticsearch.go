package config

import (
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/levigross/grequests"
	"github.com/olivere/elastic"
	"log"
	"os"
)

var Elastic *elastic.Client

func elasticInit() {
	if Elastic == nil {
		elasticConfigUrl := getConfigUrl(conf.String("go.config.prefix.elasticsearch"))
		resp, err := grequests.Get(elasticConfigUrl, nil)
		cfg := koanf.New(".")
		cfg.Load(rawbytes.Provider([]byte(resp.String())), yaml.Parser())
		//logger.Debug("Elastic地址:" + cfg.String("go.elasticsearch.uri"))
		user := cfg.String("go.elasticsearch.user")
		password := cfg.String("go.elasticsearch.password")
		if user != "" && password != "" {
			//logger.Debug("user:"+user+"   password:"+password)
			Elastic, err = elastic.NewClient(elastic.SetURL(cfg.String("go.elasticsearch.uri")), elastic.SetBasicAuth(user, password), elastic.SetInfoLog(log.New(os.Stdout, "Elasticsearch", log.LstdFlags)), elastic.SetSniff(false))
		} else {
			Elastic, err = elastic.NewClient(elastic.SetURL(cfg.String("go.elasticsearch.uri")), elastic.SetInfoLog(log.New(os.Stdout, "Elasticsearch", log.LstdFlags)), elastic.SetSniff(false))
		}
		if err != nil {
			logger.Error("Elasticsearch连接错误:" + err.Error())
		}
	}
}

func elasticClose() {
	Elastic = nil
}

func elasticCheck() {
	if Elastic == nil || !Elastic.IsRunning() {
		elasticInit()
	}
}
