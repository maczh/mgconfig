package mgconfig

import (
	"fmt"
	"github.com/couchbase/gocb/v2"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/levigross/grequests"
	"time"
)

var CouchDB *gocb.Cluster

func couchbaseInit() {
	if CouchDB == nil {
		couchbaseConfigUrl := getConfigUrl(conf.String("go.config.prefix.couchdb"))
		resp, _ := grequests.Get(couchbaseConfigUrl, nil)
		cfg := koanf.New(".")
		cfg.Load(rawbytes.Provider([]byte(resp.String())), yaml.Parser())
		host := cfg.String("go.data.couchbase.host")
		user := cfg.String("go.data.couchbase.user")
		password := cfg.String("go.data.couchbase.password")
		couchUrl := fmt.Sprintf("couchbase://%s", host)
		var err error
		CouchDB, err = gocb.Connect(couchUrl, gocb.ClusterOptions{Username: user, Password: password})
		if err != nil {
			logger.Error("CouchBase连接错误:" + err.Error())
		} else {
			logger.Info("CouchBase连接成功")
		}
	}
}

func couchbaseCheck() {
	if CouchDB == nil {
		couchbaseInit()
		return
	}
	_, err := CouchDB.Ping(&gocb.PingOptions{Timeout: 1 * time.Second})
	if err != nil {
		couchbaseClose()
		couchbaseInit()
	}
}

func couchbaseClose() {
	if CouchDB != nil {
		CouchDB.Close(nil)
	}
}
