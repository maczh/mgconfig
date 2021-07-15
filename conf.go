package config

import (
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/sadlil/gologger"
	"strings"
	"time"
)

var conf *koanf.Koanf
var logger = gologger.GetLogger()

const config_file = "./application.yml"
const AUTO_CHECK_MINUTES = 30 //自动检查连接间隔时间，单位为分钟

func InitConfig(cf string) {
	if cf == "" {
		cf = config_file
	}
	logger.Debug("读取配置文件:" + cf)
	conf = koanf.New(".")
	f := file.Provider(cf)
	err := conf.Load(f, yaml.Parser())
	if err != nil {
		logger.Error("读取配置文件错误:" + err.Error())
	}

	configs := conf.String("go.config.used")

	if strings.Contains(configs, "mysql") {
		logger.Info("正在连接MySQL")
		mysqlInit()
	}
	if strings.Contains(configs, "pgsql") {
		logger.Info("正在连接PostgreSQL")
		pgsqlInit()
	}
	if strings.Contains(configs, "mssql") {
		logger.Info("正在连接MSSQL")
		mssqlInit()
	}
	if strings.Contains(configs, "mongodb") {
		logger.Info("正在连接MongoDB")
		mgoInit()
	}
	if strings.Contains(configs, "redis") {
		logger.Info("正在连接Redis")
		redisInit()
	}
	if strings.Contains(configs, "ssdb") {
		logger.Info("正在连接SSDB")
		ssdbInit()
	}
	if strings.Contains(configs, "nacos") {
		logger.Info("正在注册到Nacos")
		registerNacos()
	}
	if strings.Contains(configs, "consul") {
		logger.Info("正在注册到Consul")
		registerConsul()
	}
	if strings.Contains(configs, "rabbitmq") {
		logger.Info("正在连接RabbitMQ")
		rabbitMQInit()
	}
	if strings.Contains(configs, "elasticsearch") {
		logger.Info("正在连接Elasticsearch")
		elasticInit()
	}
	if strings.Contains(configs, "opensearch") {
		logger.Info("正在连接OpenSearch")
		openSearchInit()
	}
	if strings.Contains(configs, "hbase") {
		logger.Info("正在连接HBase")
		hbaseInit()
	}
	if strings.Contains(configs, "hive") {
		logger.Info("正在连接Hive")
		hiveInit()
	}
	if strings.Contains(configs, "couchdb") {
		logger.Info("正在连接Couchbase")
		couchbaseInit()
	}
	if strings.Contains(configs, "influxdb") {
		logger.Info("正在连接Influxdb")
		influxdbInit()
	}

	//设置定时任务自动检查
	ticker := time.NewTicker(time.Minute * AUTO_CHECK_MINUTES)
	go func() {
		for _ = range ticker.C {
			checkAll()
		}
	}()

}

func GetConfigString(name string) string {
	if conf.Exists(name) {
		return conf.String(name)
	} else {
		return ""
	}
}

func GetConfigInt(name string) int {
	if conf.Exists(name) {
		return conf.Int(name)
	} else {
		return 0
	}
}

func SafeExit() {

	configs := conf.String("go.config.used")

	if strings.Contains(configs, "mysql") {
		logger.Info("正在关闭MySQL连接")
		mysqlClose()
	}
	if strings.Contains(configs, "pgsql") {
		logger.Info("正在关闭PostgreSQL连接")
		pgsqlClose()
	}
	if strings.Contains(configs, "mssql") {
		logger.Info("正在关闭MSSQL连接")
		mssqlClose()
	}
	if strings.Contains(configs, "mongodb") {
		logger.Info("正在关闭MongoDB连接")
		mgoClose()
	}
	if strings.Contains(configs, "redis") {
		logger.Info("正在关闭Redis连接")
		redisClose()
	}
	if strings.Contains(configs, "ssdb") {
		logger.Info("正在关闭SSDB连接")
		ssdbClose()
	}
	if strings.Contains(configs, "nacos") {
		logger.Info("正在注销Nacos")
		deRegisterNacos()
	}
	if strings.Contains(configs, "consul") {
		logger.Info("正在注销Consul")
		deRegisterConsul()
	}
	if strings.Contains(configs, "rabbitmq") {
		logger.Info("正在关闭RabbitMQ连接")
		rabbitMQClose()
	}
	if strings.Contains(configs, "elasticsearch") {
		logger.Info("正在关闭Elasticsearch连接")
		elasticClose()
	}
	if strings.Contains(configs, "opensearch") {
		logger.Info("正在关闭OpenSearch连接")
		openSearchClose()
	}
	if strings.Contains(configs, "hbase") {
		logger.Info("正在关闭HBase连接")
		hbaseClose()
	}
	if strings.Contains(configs, "hive") {
		logger.Info("正在关闭Hive连接")
		hiveClose()
	}
	if strings.Contains(configs, "couchdb") {
		logger.Info("正在关闭Couchbase连接")
		couchbaseClose()
	}
	if strings.Contains(configs, "influxdb") {
		logger.Info("正在关闭Influxdb连接")
		influxdbClose()
	}

}

func checkAll() {

	configs := conf.String("go.config.used")

	if strings.Contains(configs, "mysql") {
		logger.Debug("正在检查MySQL")
		mysqlCheck()
	}
	if strings.Contains(configs, "pgsql") {
		logger.Debug("正在检查PostgreSQL")
		pgsqlCheck()
	}
	if strings.Contains(configs, "mssql") {
		logger.Debug("正在检查MSSQL")
		mssqlCheck()
	}
	if strings.Contains(configs, "mongodb") {
		logger.Debug("正在检查MongoDB")
		MgoCheck()
	}
	if strings.Contains(configs, "elasticsearch") {
		logger.Debug("正在检查Elasticsearch")
		elasticCheck()
	}
	if strings.Contains(configs, "opensearch") {
		logger.Debug("正在检查OpenSearch")
		openSearchCheck()
	}
	if strings.Contains(configs, "redis") {
		logger.Debug("正在检查Redis")
		RedisCheck()
	}
	if strings.Contains(configs, "ssdb") {
		logger.Debug("正在检查SSDB")
		ssdbCheck()
	}
	if strings.Contains(configs, "couchdb") {
		logger.Debug("正在检查Couchbase")
		couchbaseCheck()
	}
}

func getConfigUrl(prefix string) string {
	serverType := conf.String("go.config.server_type")
	configUrl := conf.String("go.config.server")
	switch serverType {
	case "nacos":
		configUrl = configUrl + "nacos/v1/cs/configs?group=DEFAULT_GROUP&dataId=" + prefix + conf.String("go.config.mid") + conf.String("go.config.env") + conf.String("go.config.type")
	case "consul":
		configUrl = configUrl + "v1/kv/" + prefix + conf.String("go.config.mid") + conf.String("go.config.env") + conf.String("go.config.type") + "?dc=dc1&raw=true"
	case "springconfig":
		configUrl = configUrl + prefix + conf.String("go.config.mid") + conf.String("go.config.env") + conf.String("go.config.type")
	default:
		configUrl = configUrl + prefix + conf.String("go.config.mid") + conf.String("go.config.env") + conf.String("go.config.type")
	}
	return configUrl
}
