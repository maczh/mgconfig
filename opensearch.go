package mgconfig

import (
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/levigross/grequests"
	"ququ.im/opensearch"
)

var OpenSearch *opensearch.OpenSearchClient

func openSearchInit() {
	if OpenSearch == nil {
		opensearchConfigUrl := getConfigUrl(conf.String("go.config.prefix.opensearch"))
		logger.Debug("正在读取OpenSearch配置文件:" + opensearchConfigUrl)
		resp, err := grequests.Get(opensearchConfigUrl, nil)
		if err != nil {
			logger.Error("获取OpenSearch配置错误:" + err.Error())
		}
		cfg := koanf.New(".")
		cfg.Load(rawbytes.Provider([]byte(resp.String())), yaml.Parser())
		//logger.Debug("Elastic地址:" + cfg.String("go.elasticsearch.uri"))
		var cf opensearch.Config
		cf.OS_HOST = cfg.String("go.opensearch.host")
		cf.OS_APPNAME = cfg.String("go.opensearch.appname")
		cf.OS_ACCESS_KEY = cfg.String("go.opensearch.accesskey")
		cf.OS_SECRET_KEY = cfg.String("go.opensearch.secret")
		OpenSearch = opensearch.NewOpenSearchClient(cf)
	}
}

func openSearchClose() {
	OpenSearch = nil
}

func openSearchCheck() {
	if OpenSearch == nil {
		openSearchInit()
	}
}
