package mgconfig

import (
	"context"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
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
		logLevel := cfg.Int("go.data.influxdb.logLevel")
		timeout := cfg.Int("go.data.influxdb.timeout")
		batchSize := cfg.Int("go.data.influxdb.batchSize")
		if timeout == 0 {
			timeout = 1000
		}
		if batchSize == 0 {
			batchSize = 100
		}
		Influxdb = influxdb2.NewClientWithOptions(influxdbUrl, token, influxdb2.DefaultOptions().
			SetLogLevel(uint(logLevel)).
			SetHTTPRequestTimeout(uint(timeout)).
			SetBatchSize(uint(batchSize)))
	}
}

func influxdbClose() {
	if Influxdb != nil {
		Influxdb.Close()
	}
}

func influxdbCheck() {
	if Influxdb == nil {
		influxdbInit()
		return
	}
	ok, err := Influxdb.Ping(context.Background())
	if err != nil || !ok {
		influxdbClose()
		influxdbInit()
	}
}
