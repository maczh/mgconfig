package mgconfig

import (
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
)

var Nacos naming_client.INamingClient
var cluster, group string
var json = jsoniter.ConfigCompatibleWithStandardLibrary
var lan bool
var lanNetwork string

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
		var err error
		path, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		path += "/cache"
		_, err = os.Stat(path)
		if err != nil && os.IsNotExist(err) {
			os.Mkdir(path, 0777)
			path += "/naming"
			os.Mkdir(path, 0777)
		}
		lan = cfg.Bool("go.nacos.lan")
		lanNetwork = cfg.String("go.nacos.lanNet")
		serverConfigs := []constant.ServerConfig{}
		ipstr := cfg.String("go.nacos.server")
		portstr := cfg.String("go.nacos.port")
		group = cfg.String("go.nacos.group")
		if group == "" {
			group = "DEFAULT_GROUP"
		}
		ips := strings.Split(ipstr, ",")
		ports := strings.Split(portstr, ",")
		for i, ip := range ips {
			port, _ := strconv.Atoi(ports[i])
			serverConfig := constant.ServerConfig{
				IpAddr:      ip,
				Port:        uint64(port),
				ContextPath: "/nacos",
			}
			serverConfigs = append(serverConfigs, serverConfig)
		}
		logger.Debug("Nacos服务器配置: " + toJSON(serverConfigs))
		clientConfig := constant.ClientConfig{}
		clientConfig.LogLevel = "error"
		if conf.Exists("go.nacos.clientConfig.logLevel") {
			clientConfig.LogLevel = conf.String("go.nacos.clientConfig.logLevel")
		}
		clientConfig.UpdateCacheWhenEmpty = true
		if conf.Exists("go.nacos.clientConfig.updateCacheWhenEmpty") {
			clientConfig.UpdateCacheWhenEmpty = conf.Bool("go.nacos.client.updateCacheWhenEmpty")
		}
		logger.Debug("Nacos客户端配置: " + toJSON(clientConfig))
		Nacos, err = clients.CreateNamingClient(map[string]interface{}{
			"serverConfigs": serverConfigs,
			"clientConfig":  clientConfig,
		})
		if err != nil {
			logger.Error("Nacos服务连接失败:" + err.Error())
			return
		}
		localip, _ := localIPv4s(lan, lanNetwork)
		ip := localip[0]
		if conf.Exists("go.application.ip") {
			ip = conf.String("go.application.ip")
		}
		cluster = cfg.String("go.nacos.clusterName")
		port := uint64(conf.Int("go.application.port"))
		metadata := make(map[string]string)
		if port == 0 || conf.String("go.application.port_ssl") != "" {
			port = uint64(conf.Int64("go.application.port_ssl"))
			metadata["ssl"] = "true"
		}
		success, regerr := Nacos.RegisterInstance(vo.RegisterInstanceParam{
			Ip:          ip,
			Port:        port,
			ServiceName: conf.String("go.application.name"),
			Weight:      1,
			ClusterName: cluster,
			Enable:      true,
			Healthy:     true,
			Ephemeral:   true,
			Metadata:    metadata,
			GroupName:   group,
		})
		if !success {
			logger.Error("Nacos注册服务失败:" + regerr.Error())
			return
		}

		err = Nacos.Subscribe(&vo.SubscribeParam{
			ServiceName: conf.String("go.application.name"),
			Clusters:    []string{cluster},
			GroupName:   group,
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
		GroupName:   group,
	})
	if err != nil {
		instance, err = Nacos.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
			ServiceName: servicename,
			Clusters:    []string{cluster},
			GroupName:   "DEFAULT_GROUP",
		})
		if err != nil {
			logger.Error("获取Nacos服务" + servicename + "失败:" + err.Error())
			return ""
		}
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
		GroupName:   group,
		SubscribeCallback: func(services []model.SubscribeService, err error) {
			logger.Debug("callback return services:" + toJSON(services))
		},
	})
	if err != nil {
		logger.Error("Nacos服务订阅失败:" + err.Error())
	}
	ips, _ := localIPv4s(lan, lanNetwork)
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

func localIPv4s(lan bool, lanNetwork string) ([]string, error) {
	var ips, ipLans, ipWans []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips, err
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			if ipnet.IP.IsPrivate() {
				ipLans = append(ipLans, ipnet.IP.String())
				if lan && strings.HasPrefix(ipnet.IP.String(), lanNetwork) {
					ips = append(ips, ipnet.IP.String())
				}
			}
			if !ipnet.IP.IsPrivate() {
				ipWans = append(ipWans, ipnet.IP.String())
				if !lan {
					ips = append(ips, ipnet.IP.String())
				}
			}
		}
	}
	if len(ips) == 0 {
		if lan {
			ips = append(ips, ipWans...)
		} else {
			ips = append(ips, ipLans...)
		}
	}
	return ips, nil
}
