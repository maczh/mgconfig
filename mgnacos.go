package mgconfig

import (
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/levigross/grequests"
	"github.com/maczh/nacos"
	"net"
	"strings"
)

var cluster, group string
var lan bool
var lanNetwork string

func registerToNacos() {
	nacosConfigUrl := getConfigUrl(conf.String("go.config.prefix.nacos"))
	resp, _ := grequests.Get(nacosConfigUrl, nil)
	cfg := koanf.New(".")
	cfg.Load(rawbytes.Provider([]byte(resp.String())), yaml.Parser())
	lan = cfg.Bool("go.nacos.lan")
	lanNetwork = cfg.String("go.nacos.lanNet")
	serverIp := cfg.String("go.nacos.server")
	serverPort := cfg.Int("go.nacos.port")
	group = cfg.String("go.nacos.group")
	localip, _ := localIPv4s(lan, lanNetwork)
	ip := localip[0]
	if conf.Exists("go.application.ip") {
		ip = conf.String("go.application.ip")
	}
	port := conf.Int("go.application.port")
	metadata := make(map[string]string)
	if port == 0 || conf.String("go.application.port_ssl") != "" {
		port = conf.Int("go.application.port_ssl")
		metadata["ssl"] = "true"
	}
	if conf.Exists("go.application.debug") && conf.Bool("go.application.debug") {
		metadata["debug"] = "true"
	}
	serviceName := conf.String("go.application.name")
	nacos.Init(serverIp,serverPort,serviceName,group,ip,port,metadata)
}

func deRegisterFromNacos() {
	err := nacos.GetNamingInstance().Unregister()
	if err != nil {
		logger.Error("nacos unregister failed: "+ err.Error())
	}
}

func GetNacosServiceURL(serviceName string) string {
	return nacos.GetNamingInstance().GetServiceUrl(serviceName)
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
