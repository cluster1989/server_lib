package libzookeeper

import (
	"encoding/json"
	"fmt"
	"path"
	"strconv"

	"github.com/samuel/go-zookeeper/zk"
)

// 返回所有的kafka的所有broker
func (z *ZooKeeper) KafkaBrokers() (map[int32]string, error) {
	root := fmt.Sprintf("%s/brokers/ids", z.option.Chroot)
	children, _, err := z.conn.Children(root)
	if err != nil {
		return nil, err
	}

	result := make(map[int32]string)
	for _, child := range children {
		brokerID, err := strconv.ParseInt(child, 10, 32)
		if err != nil {
			return nil, err
		}
		val, _, err := z.conn.Get(path.Join(root, child))
		if err != nil {
			return nil, err
		}

		model := &KafkaBrokerModel{}
		if err := json.Unmarshal(val, model); err != nil {
			return nil, err
		}

		result[int32(brokerID)] = fmt.Sprintf("%s:%d", model.Host, model.Port)
	}
	return result, nil
}

// 返回borker的数组
func (z *ZooKeeper) KafkaBrokerList() ([]string, error) {
	brokers, err := z.KafkaBrokers()
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(brokers))
	for _, broker := range brokers {
		result = append(result, broker)
	}
	return result, nil
}

// 如果kafka的控制器已经选举出来，会直接返回控制器的brokerid
func (z *ZooKeeper) KafkaController() (int32, error) {
	node := fmt.Sprintf("%s/controller", z.option.Chroot)
	data, _, err := z.conn.Get(node)
	if err != nil {
		return -1, err
	}
	controller := &KakfaControllerModel{}
	if err := json.Unmarshal(data, &controller); err != nil {
		return -1, err
	}
	return controller.BrokerID, nil
}

// 新建一个topic对象
func (z *ZooKeeper) NewKafkaTopic(topic string) *Topic {
	return &Topic{Name: topic, zookeeper: z}
}

// 得到所有kafka注册的topic的名称
func (z *ZooKeeper) KafkaTopics() (TopicList, error) {
	root := fmt.Sprintf("%s/brokers/topics", z.option.Chroot)

	children, _, err := z.conn.Children(root)
	if err != nil {
		return nil, err
	}

	result := make(TopicList, 0, len(children))
	for _, name := range children {
		result = append(result, z.NewKafkaTopic(name))
	}
	return result, nil
}

// 返回kafka的topic，并且监听改变
func (z *ZooKeeper) WatchKafkaTopics() (TopicList, <-chan zk.Event, error) {
	root := fmt.Sprintf("%s/brokers/topics", z.option.Chroot)
	children, _, c, err := z.conn.ChildrenW(root)
	if err != nil {
		return nil, nil, err
	}
	result := make(TopicList, 0, len(children))
	for _, name := range children {
		result = append(result, z.NewKafkaTopic(name))
	}
	return result, c, nil
}
