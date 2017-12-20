package libkafka2

import (
	"runtime"
	"time"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"github.com/wuqifei/server_lib/libzookeeper"
	"github.com/wuqifei/server_lib/logs"
)

type ConsumerOption struct {
	Zookeeper    *libzookeeper.Option
	Group        string
	LatestOffset bool //new offset or old offset
	Topics       []string
	NotHighLevel bool //is not high level model
}

type Consumer struct {
	option  *ConsumerOption
	brokers []string
	zk      *libzookeeper.ZooKeeper
	config  *cluster.Config
	*cluster.Consumer
	watcher     chan []string
	MessageChan chan *sarama.ConsumerMessage
	ErrorChan   chan error
}

func NewConsumer(option *ConsumerOption) *Consumer {
	c := &Consumer{}
	// 初始化zk
	zk, err := NewZK(option.Zookeeper)
	if err != nil {
		panic(err)
	}
	c.zk = zk
	c.option = option
	c.watcher = make(chan []string, 0)
	c.MessageChan = make(chan *sarama.ConsumerMessage, 0)
	c.ErrorChan = make(chan error, 0)

	brokers, err := fetchBrokers(c.zk)
	if err != nil || len(brokers) == 0 {
		panic(err)
	}
	c.brokers = brokers
	config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	// 心跳时间
	config.Group.Heartbeat.Interval = 3 * time.Second
	config.Group.Session.Timeout = 30 * time.Second
	config.Group.Offsets.Retry.Max = 3
	// config.Group

	if c.option.NotHighLevel {
		config.Group.Mode = cluster.ConsumerModePartitions
	} else {
		config.Group.Mode = cluster.ConsumerModeMultiplex
	}

	if c.option.LatestOffset {
		// 从log head
		config.Consumer.Offsets.Initial = sarama.OffsetNewest
	} else {
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	c.brokers, err = fetchBrokers(c.zk)
	if err != nil {
		panic(err)
	}

	go watchBrokers(c.zk, c.watcher)
	c.config = config
	if err := c.dial(); err != nil {
		go c.redial()
	} else {
	}
	go c.runConsumer()
	return c
}

func (c *Consumer) reNewKafka() {
	var err error
	c.brokers, err = fetchBrokers(c.zk)
	if err != nil {
		panic(err)
	}

	go watchBrokers(c.zk, c.watcher)
	if err := c.dial(); err != nil {
		logs.Error("redial zk: ", err)
		go c.redial()
	} else {
		logs.Info("already connect kafka")
	}
	go c.runConsumer()
}

func (c *Consumer) dial() (err error) {

	consumer, err := cluster.NewConsumer(c.brokers, c.option.Group, c.option.Topics, c.config)
	if err != nil {
		return err
	}
	c.Consumer = consumer
	return nil
}

func (c *Consumer) redial() {
	var err error
	for {
		if err = c.dial(); err == nil {
			logs.Info("kafka retry new consumer ok")
			return
		}
		logs.Error("dial kafka consumer error: ", err)
		time.Sleep(time.Second)
	}
}

func (c *Consumer) runConsumer() {
	for {
		select {
		// 消费error
		case err := <-c.Consumer.Errors():

			if err != nil {
				c.ErrorChan <- err
			}
		case ntf := <-c.Consumer.Notifications():
			{

				logs.Info("kafka consumer reblance [%v]", ntf)

			}
		case msg, ok := <-c.Consumer.Messages():
			{
				if ok {
					c.MessageChan <- msg
				}
			}
		case part, ok := <-c.Consumer.Partitions():
			{
				if ok {
					// 消费partition
					go func(pc cluster.PartitionConsumer, conumer *Consumer) {
						for msg := range pc.Messages() {
							conumer.MessageChan <- msg
						}
					}(part, c)
				}
			}
		case brokers := <-c.watcher:
			{
				// 因为broker变化
				c.brokers = brokers
				// 先关闭消费
				c.Close()
				// 重新连接
				c.reNewKafka()

				// 放入空闲列表
				runtime.Goexit()
				return
			}
		}
	}
}

// 关闭消费者
func (c *Consumer) Close() error {
	c.zk.Close()
	return c.Consumer.Close()
}
