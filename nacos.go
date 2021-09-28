package mgconfig

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/levigross/grequests"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"net"
	"strconv"
	"strings"
)

var Nacos naming_client.INamingClient
var cluster string
var json = jsoniter.ConfigCompatibleWithStandardLibrary

func toJSON(o interface{}) string {
	j, err := json.Marshal(o)
	if err != nil {
		return "{}"
	} else {
		js := string(j)
		js = strings.Replace(js, "\\u003c", "<", -1)
		js = strings.Replace(js, "\\u003e", ">", -1)
		js = strings.Replace(js, "\\u0026", "&", -1)
		return js
	}
}

func registerNacos() {
	if Nacos == nil {
		nacosConfigUrl := getConfigUrl(conf.String("go.config.prefix.nacos"))
		resp, _ := grequests.Get(nacosConfigUrl, nil)
		cfg := koanf.New(".")
		cfg.Load(rawbytes.Provider([]byte(resp.String())), yaml.Parser())
		server := constant.ServerConfig{
			IpAddr:      cfg.String("go.nacos.server"),
			Port:        uint64(cfg.Int64("go.nacos.port")),
			ContextPath: "/nacos",
		}
		logger.Debug("Nacos服务器配置: " + toJSON(server))
		var err error
		Nacos, err = clients.CreateNamingClient(map[string]interface{}{
			"serverConfigs": []constant.ServerConfig{server},
		})
		if err != nil {
			logger.Error("Nacos服务连接失败:" + err.Error())
			return
		}
		ips, _ := localIPv4s()
		ip := ips[0]
		if conf.Exists("go.application.ip") {
			ip = conf.String("go.application.ip")
		}
		metadata := make(map[string]string)
		if conf.Exists("go.application.cert") {
			metadata["ssl"] = "true"
		}
		cluster = cfg.String("go.nacos.clusterName")
		success, regerr := Nacos.RegisterInstance(vo.RegisterInstanceParam{
			Ip:          ip,
			Port:        uint64(conf.Int("go.application.port")),
			ServiceName: conf.String("go.application.name"),
			Weight:      1,
			ClusterName: cluster,
			Enable:      true,
			Healthy:     true,
			Ephemeral:   true,
			Metadata:    metadata,
		})
		if !success {
			logger.Error("Nacos注册服务失败:" + regerr.Error())
			return
		}
		err = Nacos.Subscribe(&vo.SubscribeParam{
			ServiceName: conf.String("go.application.name"),
			Clusters:    []string{cluster},
			GroupName:   "DEFAULT_GROUP",
			SubscribeCallback: func(services []model.SubscribeService, err error) {
				logger.Debug("callback return services:" + toJSON(services))
			},
		})
		if err != nil {
			logger.Error("Nacos服务订阅失败:" + err.Error())
		}
	}

}

func GetNacosServiceURL(servicename string) string {
	instance, err := Nacos.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		ServiceName: servicename,
		Clusters:    []string{cluster},
	})
	if err != nil {
		logger.Error("获取Nacos服务" + servicename + "失败:" + err.Error())
		return ""
	}
	url := "http://" + instance.Ip + ":" + strconv.Itoa(int(instance.Port))
	if instance.Metadata != nil && instance.Metadata["ssl"] == "true" {
		url = "https://" + instance.Ip + ":" + strconv.Itoa(int(instance.Port))
	}
	logger.Debug("Nacos获取" + servicename + "服务成功:" + url)
	return url
}

func deRegisterNacos() {
	err := Nacos.Unsubscribe(&vo.SubscribeParam{
		ServiceName: conf.String("go.application.name"),
		Clusters:    []string{cluster},
		GroupName:   "DEFAULT_GROUP",
		SubscribeCallback: func(services []model.SubscribeService, err error) {
			logger.Debug("callback return services:" + toJSON(services))
		},
	})
	if err != nil {
		logger.Error("Nacos服务订阅失败:" + err.Error())
	}
	ips, _ := localIPv4s()
	ip := ips[0]
	if conf.Exists("go.application.ip") {
		ip = conf.String("go.application.ip")
	}
	success, regerr := Nacos.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          ip,
		Port:        uint64(conf.Int("go.application.port")),
		ServiceName: conf.String("go.application.name"),
		Cluster:     cluster,
		Ephemeral:   true,
	})
	if !success {
		logger.Error("Nacos取消注册服务失败:" + regerr.Error())
		return
	}

}

func localIPv4s() ([]string, error) {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips, err
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			ips = append(ips, ipnet.IP.String())
		}
	}

	return ips, nil
}
