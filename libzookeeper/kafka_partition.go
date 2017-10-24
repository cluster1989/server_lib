package libzookeeper

import (
	"encoding/json"
	"fmt"
)

type Partition struct {
	topic    *Topic
	ID       int32
	Replicas []int32
}

func (p *Partition) Topic() *Topic {
	return p.topic
}

func (p *Partition) Key() string {
	return fmt.Sprintf("%s/%d", p.topic.Name, p.ID)
}

// 返回第一个备份的pariton id
func (p *Partition) PreferredReplica() int32 {
	if len(p.Replicas) > 0 {
		return p.Replicas[0]
	}
	return -1
}

// 返回节点的leader
func (p *Partition) Leader() (int32, error) {
	if state, err := p.State(); err != nil {
		return -1, err
	} else {
		return state.Leader, nil
	}
}

// 返回备份的isr的partition id
func (p *Partition) ISR() ([]int32, error) {
	state, err := p.State()
	if err != nil {
		return nil, err
	}
	return state.ISR, nil
}

// 是否正在备份
func (p *Partition) UnderReplicated() (bool, error) {
	state, err := p.State()
	if err != nil {
		return false, err
	}
	b := len(state.ISR) < len(p.Replicas)
	return b, nil
}

// 是否是在使用最佳备份
func (p *Partition) UsesPreferedReplica() (bool, error) {
	state, err := p.State()
	if err != nil {
		return false, err
	}
	b := len(state.ISR) > 0 && state.ISR[0] == p.Replicas[0]
	return b, nil
}

func (p *Partition) State() (*KafkaPartitionStateModel, error) {
	state := &KafkaPartitionStateModel{}

	node := fmt.Sprintf("%s/brokers/topics/%s/partitions/%d/state", p.topic.zookeeper.option.Chroot, p.topic.Name, p.ID)

	value, _, err := p.topic.zookeeper.conn.Get(node)
	if err != nil {
		return state, err
	}

	if err := json.Unmarshal(value, state); err != nil {
		return nil, err
	}
	return state, nil
}
