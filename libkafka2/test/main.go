package main

import (
	"fmt"
	"time"

	"github.com/Shopify/sarama"

	"github.com/wuqifei/server_lib/libkafka2"
	"github.com/wuqifei/server_lib/libzookeeper"
	"github.com/wuqifei/server_lib/signal"
)

func main() {
	// go producer()
	consumer()

	signal.InitSignal()
}

func producer() {
	zkOPtion := &libzookeeper.Option{}
	zkOPtion.Addrs = []string{"127.0.0.1:2181"}
	zkOPtion.Timeout = 30 * time.Second
	option := &libkafka2.ProducerOption{}
	option.Zookeeper = zkOPtion
	producer := libkafka2.NewProducer(option)
	var i = 111111
	for {
		i++
		producer.EasySend("golangtest2", []byte(fmt.Sprintf("new:[%d]", i)))
		time.Sleep(1 * time.Second)
	}
}

func consumer() {
	zkOPtion := &libzookeeper.Option{}
	zkOPtion.Addrs = []string{"127.0.0.1:2181"}
	zkOPtion.Timeout = 30 * time.Second

	option := &libkafka2.ConsumerOption{}
	option.Zookeeper = zkOPtion

	option.Group = "1251"
	option.LatestOffset = false //olddest1
	option.Topics = []string{"golangtest2"}
	option.NotHighLevel = false
	msgchan := make(chan *sarama.ConsumerMessage, 0)

	c := libkafka2.NewConsumer(option)
	c.MessageChan = msgchan
	c.MarkPartitionOffset("golangtest2", 0, 12, "")

	for val := range c.MessageChan {
		c.MarkOffset(val, "")
	}
}
