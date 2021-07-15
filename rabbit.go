package config

import (
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/levigross/grequests"
	"github.com/socifi/jazz"
)

var rabbit *jazz.Connection
var exchange string

func rabbitMQInit() {
	if rabbit == nil {
		rabbitConfigUrl := getConfigUrl(conf.String("go.config.prefix.rabbitmq"))
		resp, _ := grequests.Get(rabbitConfigUrl, nil)
		cfg := koanf.New(".")
		cfg.Load(rawbytes.Provider([]byte(resp.String())), yaml.Parser())
		dsn := cfg.String("go.rabbitmq.uri")
		client, err := jazz.Connect(dsn)
		if err != nil {
			logger.Error("RabbitMQ连接错误:" + err.Error())
		} else {
			rabbit = client
			exchange = cfg.String("go.rabbitmq.exchange")
		}
	}
}

func rabbitMQClose() {
	if rabbit != nil {
		rabbit.Close()
		rabbit = nil
	}
}

func RabbitSendMessage(queueName string, msg string) {
	err := rabbit.SendMessage(exchange, queueName, msg)
	if err != nil {
		logger.Error("RabbitMQ发送消息错误:" + err.Error())
	}
}

func RabbitMessageListener(queueName string, listener func(msg []byte)) {
	//侦听之前先创建队列
	RabbitCreateNewQueue(queueName)
	//启动侦听消息处理线程
	go rabbit.ProcessQueue(queueName, listener)
}

func RabbitCreateNewQueue(queueName string) {
	queues := make(map[string]jazz.QueueSpec)
	binding := &jazz.Binding{
		Exchange: exchange,
		Key:      queueName,
	}
	queueSpec := &jazz.QueueSpec{
		Durable:  true,
		Bindings: []jazz.Binding{*binding},
	}
	queues[queueName] = *queueSpec
	setting := &jazz.Settings{
		Queues: queues,
	}
	err := rabbit.CreateScheme(*setting)
	if err != nil {
		logger.Error("RabbitMQ创建队列失败:" + err.Error())
	}
}
