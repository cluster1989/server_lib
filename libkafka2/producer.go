package libkafka2

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/Shopify/sarama"
	"github.com/wuqifei/server_lib/libzookeeper"
)

var (
	ZookeeperConfigurationErr   = fmt.Errorf("zookeeper configuration error")
	KafkaProducerNotInitialized = fmt.Errorf("kafka producer not initialized")
)

type Producer struct {
	asyncProducer sarama.AsyncProducer
	option        *ProducerOption
	brokers       []string
	zk            *libzookeeper.ZooKeeper

	watcher chan []string

	MessageChan chan *sarama.ProducerMessage
	ErrorChan   chan error
}

type ProducerOption struct {
	Zookeeper *libzookeeper.Option
}

// 新建生产者
func NewProducer(option *ProducerOption) *Producer {
	p := &Producer{}
	p.option = option

	// 初始化zk
	zk, err := NewZK(option.Zookeeper)
	if err != nil {
		panic(err)
	}
	p.zk = zk
	p.watcher = make(chan []string, 0)

	p.MessageChan = make(chan *sarama.ProducerMessage, 0)
	p.ErrorChan = make(chan error, 0)
	p.brokers, err = fetchBrokers(zk)
	if err != nil {
		panic(err)
	}
	go watchBrokers(zk, p.watcher)

	if err := p.asyncDial(); err != nil {
		go p.reAsyncDial()
	}

	return p
}

func (p *Producer) asyncDial() (err error) {
	config := sarama.NewConfig()
	// 等待leader返回即可
	config.Producer.RequiredAcks = sarama.WaitForLocal       // Only wait for the leader to ack
	config.Producer.Compression = sarama.CompressionSnappy   // Compress messages
	config.Producer.Flush.Frequency = 500 * time.Millisecond // Flush batches every 500ms
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	if p.asyncProducer, err = sarama.NewAsyncProducer(p.brokers, config); err == nil {
		fmt.Printf("[Info]kafka producer succeed on \n")
		go p.runProducer()
	}
	return
}

func (p *Producer) reAsyncDial() {
	var err error
	for {
		if err = p.asyncDial(); err == nil {
			fmt.Printf("kafka retry new async producer ok\n")
			return
		}
		fmt.Printf("dial kafka producer error:[%v] \n", err)
		time.Sleep(time.Second)
	}
}

func (p *Producer) runProducer() {

	for {
		select {
		case producerErr := <-p.asyncProducer.Errors():
			{
				if producerErr != nil {
					p.ErrorChan <- producerErr
				}
			}
		case msg := <-p.asyncProducer.Successes():
			{

				if msg != nil {
					p.MessageChan <- msg
				}
			}

		case brokers := <-p.watcher:
			{
				fmt.Printf("[Emergency]kafka brokers changed \n")
				// 因为broker变化
				p.brokers = brokers
				// 先关闭消费
				p.Close()
				// 重新连接
				p.reNewKafka()

				// 放入空闲列表
				runtime.Goexit()
				return
			}
		}
	}

}

func (p *Producer) reNewKafka() {
	var err error
	p.brokers, err = fetchBrokers(p.zk)
	if err != nil {
		panic(err)
	}
	go watchBrokers(p.zk, p.watcher)

	if err := p.asyncDial(); err != nil {
		go p.reAsyncDial()
	}
}

func (p *Producer) EasySend(topic string, val []byte) error {
	msg := &sarama.ProducerMessage{}
	msg.Metadata = context.Background()
	msg.Value = sarama.ByteEncoder(val)
	msg.Topic = topic

	if p.asyncProducer == nil {
		return KafkaProducerNotInitialized
	}
	p.asyncProducer.Input() <- msg
	return nil
}

func (p *Producer) MultiSend(c context.Context, msg *sarama.ProducerMessage) (err error) {
	if p.asyncProducer == nil {
		err = KafkaProducerNotInitialized
	} else {
		msg.Metadata = c
		p.asyncProducer.Input() <- msg
	}
	return
}

func (p *Producer) Close() (err error) {
	return p.asyncProducer.Close()
}
