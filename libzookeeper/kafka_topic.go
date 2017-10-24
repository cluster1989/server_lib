package libzookeeper

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/samuel/go-zookeeper/zk"
)

type Topic struct {
	Name      string
	zookeeper *ZooKeeper
}

// 节点是否存在
func (t *Topic) Exists() (bool, error) {
	node := fmt.Sprintf("/brokers/topics/%s", t.Name)
	b, e := t.zookeeper.Exists(node)
	return b, e
}

// 创建一个topic
func (t *Topic) Create() error {
	node := fmt.Sprintf("/brokers/topics/%s", t.Name)
	return t.zookeeper.MkdirRecursive(node)
}

// 得到topic下所有partition
func (t *Topic) Partitions() (PartitionList, error) {
	node := fmt.Sprintf("%s/brokers/topics/%s", t.zookeeper.option.Chroot, t.Name)
	value, _, err := t.zookeeper.conn.Get(node)
	if err != nil {
		return nil, err
	}

	return t.parsePatitions(value)
}

// 得到topic下所有partition，并监控
func (t *Topic) WatchPartitions() (PartitionList, <-chan zk.Event, error) {
	node := fmt.Sprintf("%s/brokers/topics/%s", t.zookeeper.option.Chroot, t.Name)
	value, _, c, err := t.zookeeper.conn.GetW(node)
	if err != nil {
		return nil, nil, err
	}
	list, err := t.parsePatitions(value)
	return list, c, err
}

func (t *Topic) parsePatitions(val []byte) (PartitionList, error) {
	topicMetaModel := &KafkaTopicMetaModel{}
	if err := json.Unmarshal(val, topicMetaModel); err != nil {
		return nil, err
	}

	result := make(PartitionList, len(topicMetaModel.Partitions))
	for k, v := range topicMetaModel.Partitions {
		partitionID, err := strconv.ParseInt(k, 10, 32)
		if err != nil {
			return nil, err
		}
		replicaIDS := make([]int32, 0, len(v))

		for _, r := range v {
			replicaIDS = append(replicaIDS, int32(r))
		}
		result[partitionID] = t.NewPartition(int32(partitionID), replicaIDS)
	}
	return result, nil
}

// 生成一个新的partition
func (t *Topic) NewPartition(id int32, replicas []int32) *Partition {
	return &Partition{
		ID:       id,
		Replicas: replicas,
		topic:    t,
	}
}

//  返回topic等级的配置
func (t *Topic) Config() (map[string]string, error) {

	node := fmt.Sprintf("%s/config/topics/%s", t.zookeeper.option.Chroot, t.Name)
	value, _, err := t.zookeeper.conn.Get(node)
	if err != nil {
		return nil, err
	}
	topicConfig := &KafkaTopicConfigModel{}
	if err := json.Unmarshal(value, &topicConfig); err != nil {
		return nil, err
	}

	return topicConfig.ConfigMap, nil
}
