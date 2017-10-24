package main

import (
	"github.com/wuqifei/server_lib/libkafka"
	"github.com/wuqifei/server_lib/libzookeeper"
	"github.com/wuqifei/server_lib/logs"
	"github.com/wuqifei/server_lib/signal"
)

func main() {
	c := &libkafka.ConsumerOption{}

	c.Zookeeper = libzookeeper.NewConfig()
	c.Group = "test8"
	c.Offset = true
	c.Topics = []string{"test11231", "test11232"}
	c.Zookeeper.Addrs = []string{"127.0.0.1:2181", "127.0.0.1:2182", "127.0.0.1:2183"}
	consumer := libkafka.NewConsumer(c)
	go consumeproc(consumer)

	p := &libkafka.ProducerOption{}

	p.Zookeeper = libzookeeper.NewConfig()
	p.Zookeeper.Addrs = []string{"127.0.0.1:2181", "127.0.0.1:2182", "127.0.0.1:2183"}
	p.Sync = true

	producer := libkafka.NewProducer(p)
	for i := 0; i < 100; i++ {
		producer.EasySend("test11231", []byte("tweqtqwtwq"))
		producer.EasySend("test11232", []byte("436134636"))
	}

	signal.InitSignal()
}

func consumeproc(c *libkafka.Consumer) {
	for {
		msg, ok := <-c.ConsumerGroup.Messages()
		if !ok {
			logs.Info("consumeproc exit")
			return
		}
		logs.Info("recv:value[%s]", string(msg.Value))
		c.ConsumerGroup.CommitUpto(msg)
	}
}
