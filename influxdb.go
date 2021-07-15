package config

import (
	influxdb2 "github.com/influxdata/influxdb-client-go"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/levigross/grequests"
)

var Influxdb influxdb2.Client
var InfluxdbBucket, InfluxdbOrg string

func influxdbInit() {
	if Influxdb == nil {
		influxdbConfigUrl := getConfigUrl(conf.String("go.config.prefix.influxdb"))
		resp, _ := grequests.Get(influxdbConfigUrl, nil)
		cfg := koanf.New(".")
		cfg.Load(rawbytes.Provider([]byte(resp.String())), yaml.Parser())
		influxdbUrl := cfg.String("go.data.influxdb.url")
		token := cfg.String("go.data.influxdb.token")
		InfluxdbBucket = cfg.String("go.data.influxdb.bucket")
		InfluxdbOrg = cfg.String("go.data.influxdb.org")
		Influxdb = influxdb2.NewClient(influxdbUrl, token)
	}
}

func influxdbClose() {
	if Influxdb != nil {
		Influxdb.Close()
	}
}
