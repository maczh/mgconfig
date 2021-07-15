package config

import (
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/levigross/grequests"
	"github.com/tsuna/gohbase"
)

var HBase gohbase.Client

func hbaseInit() {
	if HBase == nil {
		hbaseConfigUrl := getConfigUrl(conf.String("go.config.prefix.hbase"))
		resp, _ := grequests.Get(hbaseConfigUrl, nil)
		cfg := koanf.New(".")
		cfg.Load(rawbytes.Provider([]byte(resp.String())), yaml.Parser())
		host := cfg.String("go.data.hbase.host")
		HBase = gohbase.NewClient(host)
	}
}

func hbaseClose() {
	if HBase != nil {
		HBase.Close()
	}
}
