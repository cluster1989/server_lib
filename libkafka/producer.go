package libkafka

import (
	"context"
	"fmt"
	"time"

	"github.com/wuqifei/server_lib/libkafka/zk_help"
	"github.com/wuqifei/server_lib/logs"

	"github.com/Shopify/sarama"
	"github.com/wuqifei/server_lib/libzookeeper"
)

var (
	ZookeeperConfigurationErr   = fmt.Errorf("zookeeper configuration error")
	KafkaProducerNotInitialized = fmt.Errorf("kafka producer not initialized")
)

type Producer struct {
	asyncProducer sarama.AsyncProducer
	syncProducer  sarama.SyncProducer
	c             *ProducerOption
	brokers       []string
}

type ProducerOption struct {
	Sync      bool
	Zookeeper *libzookeeper.Option
}

// 新建生产者
func NewProducer(c *ProducerOption) *Producer {
	p := &Producer{}
	p.c = c

	//先初始化zk,如果无法初始化直接panic
	if c.Zookeeper == nil || len(c.Zookeeper.Addrs) == 0 {
		panic(ZookeeperConfigurationErr)
	}

	// 初始化zk
	err := zk_help.Init(c.Zookeeper)
	if err != nil {
		panic(err)
	}

	brokers, err := zk_help.FetchBrokers()
	p.brokers = brokers

	if len(brokers) > 0 && err == nil {
		if !c.Sync {
			if err := p.asyncDial(); err != nil {
				go p.reAsyncDial()
			}
		} else {
			if err := p.syncDial(); err != nil {
				go p.reSyncDial()
			}
		}
	}

	return p
}

func (p *Producer) syncDial() (err error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll // Wait for all in-sync replicas to ack the message
	config.Producer.Retry.Max = 10                   // Retry up to 10 times to produce the message
	p.syncProducer, err = sarama.NewSyncProducer(p.brokers, config)
	return
}

func (p *Producer) reSyncDial() {
	var err error
	for {
		if err = p.syncDial(); err == nil {
			logs.Info("kafka retry new sync producer ok")
			return
		}

		logs.Info("dial kafka producer error: ", err)

		time.Sleep(time.Second)
	}
}

func (p *Producer) asyncDial() (err error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal       // Only wait for the leader to ack
	config.Producer.Compression = sarama.CompressionSnappy   // Compress messages
	config.Producer.Flush.Frequency = 500 * time.Millisecond // Flush batches every 500ms
	if p.asyncProducer, err = sarama.NewAsyncProducer(p.brokers, config); err == nil {
		go p.errproc()
		go p.successproc()
	}
	return
}

func (p *Producer) reAsyncDial() {
	var err error
	for {
		if err = p.asyncDial(); err == nil {
			logs.Info("kafka retry new async producer ok")
			return
		}
		logs.Error("dial kafka producer error: ", err)
		time.Sleep(time.Second)
	}
}

// errproc errors when aync producer publish messages.
func (p *Producer) errproc() {
	var errChan <-chan *sarama.ProducerError
	if !p.c.Sync {
		errChan = p.asyncProducer.Errors()
	} else {
		return
	}
	for {
		e, ok := <-errChan
		if !ok {
			return
		}

		logs.Error("kafka producer send message(%v) failed error(%v)", e.Msg, e.Err)

	}
}

func (p *Producer) successproc() {
	var msgChan <-chan *sarama.ProducerMessage
	if !p.c.Sync {
		msgChan = p.asyncProducer.Successes()
	} else {
		return
	}
	for {
		msg, ok := <-msgChan
		if !ok {
			return
		}
		if _, ok := msg.Metadata.(context.Context); ok {
		}
	}
}

// 简单的队列发送
func (p *Producer) EasySend(topic string, val []byte) error {

	msg := &sarama.ProducerMessage{}
	msg.Metadata = context.Background()
	msg.Value = sarama.ByteEncoder(val)
	msg.Topic = topic

	if p.c.Sync {
		if p.syncProducer == nil {
			return KafkaProducerNotInitialized
		}

		if p.syncProducer != nil {
			patition, offset, err := p.syncProducer.SendMessage(msg)
			if err != nil {
				logs.Error("producer:easy send msg error! partition:[%d] offset[%d] err[%v](topic[%s]key[%s]val[%v])", patition, offset, err, topic, val)
			} else {
				logs.Info("producer:easy send msg partition:[%d] offset[%d](topic[%s]key[%s]val[%v])", patition, offset, topic, val)
			}

			return err
		}
	}

	if p.asyncProducer == nil {
		return KafkaProducerNotInitialized
	}

	p.asyncProducer.Input() <- msg

	return nil

}

// multisend to kafka
func (p *Producer) MultiSend(c context.Context, msg *sarama.ProducerMessage) (err error) {
	if !p.c.Sync {
		if p.asyncProducer == nil {
			err = KafkaProducerNotInitialized
		} else {
			msg.Metadata = c
			p.asyncProducer.Input() <- msg
		}
	} else {
		if p.syncProducer == nil {
			err = KafkaProducerNotInitialized
		} else {
			if _, _, err = p.syncProducer.SendMessage(msg); err != nil {
				logs.Error(err)
			}
		}
	}
	return
}

func (p *Producer) Close() (err error) {
	if !p.c.Sync {
		if p.asyncProducer != nil {
			return p.asyncProducer.Close()
		}
	}
	if p.syncProducer != nil {
		return p.syncProducer.Close()
	}
	return
}
