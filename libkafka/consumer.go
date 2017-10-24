package libkafka

import (
	"time"

	"github.com/Shopify/sarama"
	"github.com/wuqifei/server_lib/libkafka/consumergroup"
	"github.com/wuqifei/server_lib/libkafka/zk_help"
	"github.com/wuqifei/server_lib/libzookeeper"
	"github.com/wuqifei/server_lib/logs"
)

type ConsumerOption struct {
	Zookeeper *libzookeeper.Option
	Group     string
	Offset    bool //new offset or old offset
	Topics    []string
}

type Consumer struct {
	ConsumerGroup *consumergroup.ConsumerGroup
	c             *ConsumerOption
	brokers       []string
}

func NewConsumer(option *ConsumerOption) *Consumer {
	c := &Consumer{}
	c.c = option
	// 初始化zk
	err := zk_help.Init(option.Zookeeper)
	if err != nil {
		panic(err)
	}

	if err := c.dial(); err != nil {
		logs.Error("redial zk: ", err)
		go c.redial()
	} else {
		logs.Info("already connect kalka")
	}
	return c
}

func (c *Consumer) dial() (err error) {
	cfg := consumergroup.NewConfig()
	if c.c.Offset {
		cfg.Offsets.Initial = sarama.OffsetNewest
	} else {
		cfg.Offsets.Initial = sarama.OffsetOldest
	}
	c.ConsumerGroup, err = consumergroup.JoinConsumerGroup(c.c.Group, c.c.Topics, cfg)
	return
}

func (c *Consumer) redial() {
	var err error
	for {
		if err = c.dial(); err == nil {
			logs.Info("kafka retry new consumer ok")
			return
		} else {
			logs.Error("dial kafka consumer error: ", err)
		}
		time.Sleep(time.Second)
	}
}

func (c *Consumer) Close() error {
	if c.ConsumerGroup != nil {
		return c.ConsumerGroup.Close()
	}
	return nil
}
