package main

import (
	"fmt"

	"github.com/wuqifei/server_lib/libkafka"
	"github.com/wuqifei/server_lib/libzookeeper"
	"github.com/wuqifei/server_lib/logs"
	"github.com/wuqifei/server_lib/signal"
)

func main() {

	c := &libkafka.ConsumerOption{}

	c.Zookeeper = libzookeeper.NewConfig()
	c.Group = "test-123"
	c.Offset = true
	c.Topics = []string{"test-1-7", "test-1-8"}
	c.Zookeeper.Addrs = []string{"127.0.0.1:2181", "127.0.0.1:2182", "127.0.0.1:2183"}
	consumer := libkafka.NewConsumer(c)
	go consumeproc1(consumer)
	go consumeproc2(consumer)

	p := &libkafka.ProducerOption{}

	p.Zookeeper = libzookeeper.NewConfig()
	p.Zookeeper.Addrs = []string{"127.0.0.1:2181", "127.0.0.1:2182", "127.0.0.1:2183"}
	p.Sync = false

	producer := libkafka.NewProducer(p)
	for i := 45; i < 80; i++ {
		producer.EasySend("test-1-7", []byte(fmt.Sprintf("tasdads-%d", i)))
		producer.EasySend("test-1-8", []byte(fmt.Sprintf("090010231-%d", i)))
	}

	signal.InitSignal()
}

func consumeproc1(c *libkafka.Consumer) {
	for {
		msg, ok := <-c.ConsumerGroup.Messages()
		if !ok {
			logs.Info("consumeproc1 exit")
			return
		}
		logs.Info("consumeproc1 recv:value[%s]", string(msg.Value))
		c.ConsumerGroup.CommitUpto(msg)
	}
}

func consumeproc2(c *libkafka.Consumer) {
	for {
		msg, ok := <-c.ConsumerGroup.Messages()
		if !ok {
			logs.Info("consumeproc2 exit")
			return
		}
		logs.Info("consumeproc2 recv:value[%s]", string(msg.Value))
		c.ConsumerGroup.CommitUpto(msg)
	}
}
