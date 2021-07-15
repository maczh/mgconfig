package config

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/levigross/grequests"
	"math/rand"
	"strconv"
	"time"
)

var Consul *api.Client

//Register service into consul
func registerConsul() {
	var err error
	if Consul == nil {
		consulConfigUrl := getConfigUrl(conf.String("go.config.prefix.consul"))
		resp, _ := grequests.Get(consulConfigUrl, nil)
		cfg := koanf.New(".")
		cfg.Load(rawbytes.Provider([]byte(resp.String())), yaml.Parser())
		config := api.DefaultConfig()
		config.Address = cfg.String("go.consul.server") + ":" + cfg.String("go.consul.port")
		config.Datacenter = cfg.String("go.consul.datacenter")
		// Get a new Consul
		Consul, err = api.NewClient(config)
		if err != nil {
			logger.Error("consul连接错误:" + err.Error())
			return
		}
	}
	agent := Consul.Agent()
	agentCheck := &api.AgentServiceCheck{
		Interval:                       "5s",
		Timeout:                        "3s",
		DeregisterCriticalServiceAfter: "30s",
	}
	ips, _ := localIPv4s()
	ip := ips[0]
	if conf.Exists("go.application.ip") {
		ip = conf.String("go.application.ip")
	}
	port := conf.Int("go.application.port")
	serviceName := conf.String("go.application.name")
	agentCheck.HTTP = fmt.Sprintf("http://%s:%d/health", ip, port)
	reg := &api.AgentServiceRegistration{
		Kind:    "",
		ID:      fmt.Sprintf("%v-%v-%v", serviceName, ip, port),
		Name:    serviceName,
		Tags:    nil,
		Port:    port,
		Address: ip,
		Check:   agentCheck,
	}
	err = agent.ServiceRegister(reg)
	if err != nil {
		logger.Error("consul注册服务错误:" + err.Error())
	}
	return
}

type Service struct {
	ServiceId string
	Ip        string
	Port      int
}

func getConsulService(serviceName string) ([]Service, error) {
	if Consul == nil {
		registerConsul()
	}
	status, servicesInfo, err := Consul.Agent().AgentHealthServiceByName(serviceName)
	if err != nil {
		logger.Error("consul获取服务失败:" + err.Error())
		return nil, err
	}
	logger.Debug("consul获取服务:" + serviceName + "返回" + status + ":" + toJSON(servicesInfo))
	if len(servicesInfo) > 0 {
		result := make([]Service, 0)
		for _, v := range servicesInfo {
			if v.AggregatedStatus == api.HealthPassing {
				s := Service{
					ServiceId: v.Service.ID,
					Ip:        v.Service.Address,
					Port:      v.Service.Port,
				}
				result = append(result, s)
			}
		}
		return result, nil
	}
	return nil, nil
}

func GetConsulServiceURL(serviceName string) string {
	serviceList, err := getConsulService(serviceName)
	if err != nil {
		logger.Error("consul服务异常:" + err.Error())
		return ""
	}
	if serviceList == nil || len(serviceList) == 0 {
		logger.Error(serviceName + "服务未注册")
		return ""
	}
	service := serviceList[0]
	if len(serviceList) > 1 {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		service = serviceList[r.Intn(len(serviceList))]
	}
	url := "http://" + service.Ip + ":" + strconv.Itoa(service.Port)
	logger.Debug("Consul获取" + serviceName + "服务成功:" + url)
	return url
}

func deRegisterConsul() {
	if Consul == nil {
		return
	}
	ips, _ := localIPv4s()
	ip := ips[0]
	if conf.Exists("go.application.ip") {
		ip = conf.String("go.application.ip")
	}
	port := conf.Int("go.application.port")
	serviceName := conf.String("go.application.name")
	serviceId := fmt.Sprintf("%v-%v-%v", serviceName, ip, port)
	err := Consul.Agent().ServiceDeregister(serviceId)
	if err != nil {
		logger.Error("consul服务反注册失败:" + err.Error())
	}
	Consul = nil
}
