package mgconfig

import (
	"context"
	"errors"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/levigross/grequests"
	"strings"
)

type kafka struct {
	confUrl string
	conf    *koanf.Koanf
	client  sarama.Client
	topics  []string
	servers []string
	config  *sarama.Config
}

var Kafka = &kafka{}

func (k *kafka) getConfig() *sarama.Config {
	ack := k.conf.String("go.data.kafka.ack")
	autoCommit := k.conf.Bool("go.data.kafka.auto_commit")
	partitioner := k.conf.String("go.data.kafka.partitioner")
	ver := k.conf.String("go.data.kafka.version")
	acks := map[string]sarama.RequiredAcks{
		"no":    sarama.NoResponse,
		"local": sarama.WaitForLocal,
		"all":   sarama.WaitForAll,
	}
	version, _ := sarama.ParseKafkaVersion(ver)
	config := sarama.NewConfig()
	config.Version = version
	config.Producer.RequiredAcks = acks[ack]
	config.Consumer.Offsets.AutoCommit.Enable = autoCommit
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	switch partitioner {
	case "hash":
		config.Producer.Partitioner = sarama.NewHashPartitioner
	case "random":
		config.Producer.Partitioner = sarama.NewRandomPartitioner
	case "round-robin":
		config.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	}
	return config
}

func (k *kafka) Init() {
	k.confUrl = getConfigUrl(conf.String("go.config.prefix.kafka"))
	if k.confUrl == "" {
		logger.Error("Kafka配置Url为空")
		return
	}
	if k.conf == nil {
		logger.Debug("正在获取kafka配置: " + k.confUrl)
		resp, err := grequests.Get(k.confUrl, nil)
		if err != nil {
			logger.Error("kafka配置下载失败! " + err.Error())
			return
		}
		k.conf = koanf.New(".")
		err = k.conf.Load(rawbytes.Provider([]byte(resp.String())), yaml.Parser())
		if err != nil {
			logger.Error("Kafka配置文件解析错误:" + err.Error())
			k.conf = nil
			return
		}
	}
	k.servers = strings.Split(k.conf.String("go.data.kafka.servers"), ",")
	k.config = k.getConfig()
	client, err := sarama.NewClient(k.servers, k.getConfig())
	if err != nil {
		logger.Error("Kafka建立连接失败: " + err.Error())
		return
	}
	k.client = client
	k.topics, err = client.Topics()
	if err != nil {
		logger.Error("Kafka获取topic清单失败: " + err.Error())
		k.topics = make([]string, 0)
	}
	if strings.Contains(client.Brokers()[0].Addr(), "127.0.0.1") {
		logger.Error("Kafka服务器配置错误，请修改服务端侦听地址")
	}
	logger.Info("Kafka建立连接成功")
}

func (k *kafka) Close() {
	err := k.client.Close()
	if err != nil {
		logger.Error("Kafka关闭连接失败: " + err.Error())
		return
	}
	return
}

func (k *kafka) Check() error {
	if k.client.Closed() {
		logger.Error("kafka client has closed")
		k.Init()
		if k.client.Closed() {
			return fmt.Errorf("kafka client closed")
		}
	}
	return nil
}

func (k *kafka) GetProducer() (sarama.AsyncProducer, error) {
	producer, err := sarama.NewAsyncProducerFromClient(k.client)
	return producer, err
}

func (k *kafka) GetConsumer() (sarama.Consumer, error) {
	consumer, err := sarama.NewConsumer(k.servers, k.getConfig())
	return consumer, err
}

func (k *kafka) GetAdminClient() (sarama.ClusterAdmin, error) {
	admin, err := sarama.NewClusterAdminFromClient(k.client)
	return admin, err
}

func (k *kafka) GetConsumerGroup(id string) (sarama.ConsumerGroup, error) {
	consumerGroup, err := sarama.NewConsumerGroupFromClient(id, k.client)
	return consumerGroup, err
}

func (k *kafka) CreateTopic(topic string) error {
	admin, err := k.GetAdminClient()
	if err != nil {
		logger.Error("Kafka连接失败:" + err.Error())
		return err
	}
	err = admin.CreateTopic(topic, &sarama.TopicDetail{NumPartitions: 1, ReplicationFactor: 1}, false)
	if err != nil {
		logger.Error("Kafka创建topic: " + topic + "失败: " + err.Error())
	}
	return err
}

func (k *kafka) Send(topic, data string) error {
	if !stringArrayContains(k.topics, topic) {
		err := k.CreateTopic(topic)
		if err != nil {
			logger.Error("Kafka创建topic失败:" + err.Error())
			return err
		}
		k.topics = append(k.topics, topic)
	}
	producer, err := k.GetProducer()
	if err != nil {
		logger.Error("Kafka连接失败:" + err.Error())
		return err
	}
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(data),
	}
	producer.Input() <- msg
	logger.Debug(fmt.Sprintf("Kafka发送消息到%s成功!内容:%s", topic, data))
	return nil
}

func (k *kafka) SendMsgs(topic string, data []string) error {
	if !stringArrayContains(k.topics, topic) {
		err := k.CreateTopic(topic)
		if err != nil {
			logger.Error("Kafka创建topic失败:" + err.Error())
			return err
		}
		k.topics = append(k.topics, topic)
	}
	producer, err := k.GetProducer()
	if err != nil {
		logger.Error("Kafka连接失败:" + err.Error())
		return err
	}
	if data == nil || len(data) == 0 {
		return errors.New("No data to send")
	}
	for _, d := range data {
		msg := &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.StringEncoder(d),
		}
		producer.Input() <- msg
	}
	return nil
}

func (k *kafka) MessageListener(groupId, topic string, listener func(msg string) error) error {
	if !stringArrayContains(k.topics, topic) {
		err := k.CreateTopic(topic)
		if err != nil {
			logger.Error("Kafka创建topic失败:" + err.Error())
			return err
		}
		k.topics = append(k.topics, topic)
	}
	handler := MsgHandler{
		Handle: listener,
	}
	consumerGroup, err := k.GetConsumerGroup(groupId)
	if err != nil {
		logger.Error("Kafka获取consumerGroup失败:" + err.Error())
		return err
	}

	go func() {
		if err := consumerGroup.Consume(context.Background(), []string{topic}, handler); err != nil {
			logger.Error("Kafka创建消费者错误: " + err.Error())
		}
	}()
	return nil
}

type MsgHandler struct {
	Handle func(msg string) error
}

func (MsgHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (MsgHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h MsgHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		//logger.Debug(fmt.Sprintf("Message topic:%q partition:%d offset:%d, msg: %s\n", msg.Topic, msg.Partition, msg.Offset, string(msg.Value)))
		err := h.Handle(string(msg.Value))
		if err != nil {
			logger.Error("Kafka消息消费处理错误: " + err.Error())
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}

func stringArrayContains(src []string, dst string) bool {
	if src == nil || len(src) == 0 {
		return false
	}
	for _, str := range src {
		if str == dst {
			return true
		}
	}
	return false
}
