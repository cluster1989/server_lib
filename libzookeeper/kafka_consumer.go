package libzookeeper

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

var (
	// consumer实例正在运行
	ErrRunningInstances = errors.New("Cannot deregister a consumergroup with running instances")
	// 实例已经被注册
	ErrInstanceAlreadyRegistered = errors.New("Cannot register consumer instance because it already is registered")
	// 实例没有注册
	ErrInstanceNotRegistered = errors.New("Cannot deregister consumer instance because it not registered")
	// 不能partition已经被他人占有
	ErrPartitionClaimedByOther = errors.New("Cannot claim partition: it is already claimed by another instance")
	// 没有被这个实例占有
	ErrPartitionNotClaimed = errors.New("Cannot release partition: it is not claimed by this instance")
)

// consumer high level 将从某个Partition读取的最后一条消息的offset存于ZooKeeper中.
// 每一个High Level Consumer实例都属于一个Consumer Group，若不指定则属于默认的Group 属于consumer group的消费方式
// lowleven 需要知道查找到一个“活着”的Broker，并且找出每个Partition的Leader,topic partition等信息，来读取数据，较为灵活，但是较为复杂

// 采用high level
type ConsumerGroup struct {
	zookeeper *ZooKeeper
	Name      string
}

// consumer group 的实例
type ConsumerGroupInstance struct {
	cg *ConsumerGroup
	ID string
}

type ConsumerGroupList []*ConsumerGroup

type ConsumerGroupInstanceList []*ConsumerGroupInstance

const (
	ConsumerRegStatic    = "static"
	ConsumerRegWhiteList = "white_list"
	ConsumerRegBlackList = "black_list"
)

const (
	ConsumerRegDefaultVersion = 1
)

// 创建一个consumer对象
func (z *ZooKeeper) NewConsumerGroup(name string) *ConsumerGroup {
	return &ConsumerGroup{
		Name:      name,
		zookeeper: z,
	}
}

// 得到所有的注册的consumergroups
func (z *ZooKeeper) ConsumerGroups() (ConsumerGroupList, error) {
	root := fmt.Sprintf("%s/consumers", z.option.Chroot)
	children, _, err := z.conn.Children(root)
	if err != nil {
		return nil, err
	}
	result := make(ConsumerGroupList, 0, len(children))
	for _, child := range children {
		consumer := z.NewConsumerGroup(child)
		result = append(result, consumer)
	}
	return result, nil
}

func (cg *ConsumerGroup) Exists() (bool, error) {
	node := fmt.Sprintf("/consumers/%s", cg.Name)
	b, e := cg.zookeeper.Exists(node)
	return b, e
}

func (cg *ConsumerGroup) Create() error {
	node := fmt.Sprintf("/consumers/%s", cg.Name)
	return cg.zookeeper.MkdirRecursive(node)
}

// 删除consumer group的节点
func (cg *ConsumerGroup) Delete() error {
	if instances, err := cg.Instances(); err != nil {
		return err
	} else if len(instances) > 0 {
		return ErrRunningInstances
	}
	node := fmt.Sprintf("/consumers/%s", cg.Name)
	return cg.zookeeper.DeleteRecursive(node)
}

func (cg *ConsumerGroup) Instances() (ConsumerGroupInstanceList, error) {
	node := fmt.Sprintf("/consumers/%s/ids", cg.Name)
	root := fmt.Sprintf("%s%s", cg.zookeeper.option.Chroot, node)
	//得到consumerid
	if flag, err := cg.zookeeper.Exists(node); err != nil {
		return nil, err
	} else if flag {
		children, _, err := cg.zookeeper.conn.Children(root)
		if err != nil {
			return nil, err
		}

		result := make(ConsumerGroupInstanceList, 0)

		for _, child := range children {
			instance := cg.Instance(child)
			result = append(result, instance)
		}
		return result, nil

	}
	result := make(ConsumerGroupInstanceList, 0)
	return result, nil
}

func (cg *ConsumerGroup) WatchInstances() (ConsumerGroupInstanceList, <-chan zk.Event, error) {
	node := fmt.Sprintf("/consumers/%s/ids", cg.Name)
	root := fmt.Sprintf("%s%s", cg.zookeeper.option.Chroot, node)
	//得到consumerid
	if flag, err := cg.zookeeper.Exists(node); err != nil {
		return nil, nil, err
	} else if !flag {
		if err := cg.zookeeper.MkdirRecursive(node); err != nil {
			return nil, nil, err
		}
	}

	children, _, c, err := cg.zookeeper.conn.ChildrenW(root)
	if err != nil {
		return nil, nil, err
	}

	result := make(ConsumerGroupInstanceList, 0)

	for _, child := range children {
		fmt.Printf("[Debug]registerd instance ID:%s len[%d]\n", child, len(children))
		instance := cg.Instance(child)
		result = append(result, instance)
	}
	return result, c, nil
}

func (cg *ConsumerGroup) Instance(id string) *ConsumerGroupInstance {
	return &ConsumerGroupInstance{
		cg: cg,
		ID: id,
	}
}

func (cg *ConsumerGroup) NewInstance() *ConsumerGroupInstance {
	id, err := GenerateConsumerInstanceID()
	if err != nil {
		panic(err)
	}
	return cg.Instance(id)
}

// 返回这个partition的拥有者
func (cg *ConsumerGroup) PartitionOwner(topic string, partition int32) (*ConsumerGroupInstance, error) {
	root := fmt.Sprintf("%s/consumers/%s/owners/%s/%d", cg.zookeeper.option.Chroot, cg.Name, topic, partition)
	val, _, err := cg.zookeeper.conn.Get(root)
	if err != nil {
		if err == zk.ErrNoNode {
			return nil, nil
		}
		return nil, err
	}
	instance := cg.Instance(string(val))
	return instance, nil
}

func (cg *ConsumerGroup) WatchPartitionOwner(topic string, partition int32) (*ConsumerGroupInstance, <-chan zk.Event, error) {
	root := fmt.Sprintf("%s/consumers/%s/owners/%s/%d", cg.zookeeper.option.Chroot, cg.Name, topic, partition)
	val, _, c, err := cg.zookeeper.conn.GetW(root)
	if err != nil {
		if err == zk.ErrNoNode {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	instance := cg.Instance(string(val))
	return instance, c, nil
}

// 是否已经注册
func (cgi *ConsumerGroupInstance) IsRegistered() (bool, error) {
	node := fmt.Sprintf("/consumers/%s/ids/%s", cgi.cg.Name, cgi.ID)
	return cgi.cg.zookeeper.Exists(node)
}

// 返回consumer实例的注册的信息
func (cgi *ConsumerGroupInstance) RegisterdMsg() (*KafkaConsumerRegistryModel, error) {
	root := fmt.Sprintf("%s/consumers/%s/ids/%s", cgi.cg.zookeeper.option.Chroot, cgi.cg.Name, cgi.ID)
	val, _, err := cgi.cg.zookeeper.conn.Get(root)
	if err != nil {
		return nil, err
	}
	reg := &KafkaConsumerRegistryModel{}
	if err := json.Unmarshal(val, reg); err != nil {
		return nil, err
	}
	return reg, nil
}

// 注册一个信息到实例里面
func (cgi *ConsumerGroupInstance) RegisterWithSubscription(subscriptionJSON []byte) error {
	if exists, err := cgi.IsRegistered(); err != nil {
		return err
	} else if exists {
		return ErrInstanceAlreadyRegistered
	}

	node := fmt.Sprintf("/consumers/%s/ids/%s", cgi.cg.Name, cgi.ID)
	return cgi.cg.zookeeper.Create(node, subscriptionJSON, true)
}

// 将topics注册到consumer实例里面
func (cgi *ConsumerGroupInstance) Register(topics []string) error {
	subcription := make(map[string]int)
	for _, topic := range topics {
		subcription[topic] = 1
	}

	model := &KafkaConsumerRegistryModel{
		Pattern:      ConsumerRegStatic,
		Subscription: subcription,
		Timestamp:    time.Now().Unix(),
		Version:      ConsumerRegDefaultVersion,
	}
	b, err := json.Marshal(model)
	if err != nil {
		return nil
	}
	return cgi.RegisterWithSubscription(b)
}

// 取消这个实例在zk里面的注册
func (cgi *ConsumerGroupInstance) Deregister() error {
	node := fmt.Sprintf("/consumers/%s/ids/%s", cgi.cg.Name, cgi.ID)
	root := fmt.Sprintf("%s%s", cgi.cg.zookeeper.option.Chroot, node)
	flag, stat, err := cgi.cg.zookeeper.conn.Exists(node)
	if err != nil {
		return err
	} else if !flag {
		return ErrInstanceNotRegistered
	}
	return cgi.cg.zookeeper.conn.Delete(root, stat.Version)
}

// 占有一个Partition，如果已经被占有，则返回错误
func (cgi *ConsumerGroupInstance) ClaimPartition(topic string, partition int32) error {
	node := fmt.Sprintf("/consumers/%s/owners/%s", cgi.cg.Name, topic)
	if err := cgi.cg.zookeeper.MkdirRecursive(node); err != nil {
		return err
	}

	// 给这个partition创建一个临时节点
	node = fmt.Sprintf("%s/%d", node, partition)
	root := fmt.Sprintf("%s%s", cgi.cg.zookeeper.option.Chroot, node)
	err := cgi.cg.zookeeper.Create(node, []byte(cgi.ID), true)
	if err != nil {
		if err == zk.ErrNodeExists {
			data, _, err := cgi.cg.zookeeper.conn.Get(root)
			if err != nil {
				return err
			}
			if string(data) != cgi.ID {
				return ErrPartitionClaimedByOther
			}
			return nil
		}
		return err
	}

	return nil
}

// 释放这个partition
func (cgi *ConsumerGroupInstance) ReleasePartition(topic string, partition int32) error {
	owner, err := cgi.cg.PartitionOwner(topic, partition)
	if err != nil {
		return err
	}
	if owner == nil || owner.ID != cgi.ID {
		return ErrPartitionNotClaimed
	}
	node := fmt.Sprintf("/consumers/%s/owners/%s/%d", cgi.cg.Name, topic, partition)
	return cgi.cg.zookeeper.conn.Delete(node, 0)
}

// 这个consumer拥有的topic
func (cg *ConsumerGroup) Topics() (TopicList, error) {
	root := fmt.Sprintf("%s/consumers/%s/owners", cg.zookeeper.option.Chroot, cg.Name)
	children, _, err := cg.zookeeper.conn.Children(root)
	if err != nil {
		return nil, err
	}
	result := make(TopicList, 0, len(children))

	for _, child := range children {
		result = append(result, cg.zookeeper.NewKafkaTopic(child))
	}
	return result, nil
}

// 提交给这个group/topic/partition 一个偏移量
func (cg *ConsumerGroup) CommitOffset(topic string, partition int32, offset int64) error {
	node := fmt.Sprintf("/consumers/%s/offsets/%s/%d", cg.Name, topic, partition)
	root := fmt.Sprintf("%s%s", cg.zookeeper.option.Chroot, node)
	data := []byte(fmt.Sprintf("%d", offset))
	_, stat, err := cg.zookeeper.conn.Get(root)
	if err != nil {
		if err == zk.ErrNoNode {
			return cg.zookeeper.Create(node, data, false)
		}
		return err
	}
	_, e := cg.zookeeper.conn.Set(node, data, stat.Version)
	return e
}

func (cg *ConsumerGroup) FetchOffset(topic string, partition int32) (int64, error) {
	node := fmt.Sprintf("/consumers/%s/offsets/%s/%d", cg.Name, topic, partition)
	root := fmt.Sprintf("%s%s", cg.zookeeper.option.Chroot, node)

	val, _, err := cg.zookeeper.conn.Get(root)
	if err == zk.ErrNoNode {
		return -1, nil
	} else if err != nil {
		return -1, err
	}
	return strconv.ParseInt(string(val), 10, 64)
}

// 得到所有的偏移
func (cg *ConsumerGroup) FetchAllOffsets() (map[string]map[int32]int64, error) {
	result := make(map[string]map[int32]int64)

	offsetsRoot := fmt.Sprintf("%s/consumers/%s/offsets", cg.zookeeper.option.Chroot, cg.Name)
	topics, _, err := cg.zookeeper.conn.Children(offsetsRoot)
	if err == zk.ErrNoNode {
		return result, nil
	} else if err != nil {
		return nil, err
	}

	for _, topic := range topics {
		result[topic] = make(map[int32]int64)
		topicRoot := fmt.Sprintf("%s/consumers/%s/offsets/%s", cg.zookeeper.option.Chroot, cg.Name, topic)
		partitions, _, err := cg.zookeeper.conn.Children(topicRoot)
		if err != nil {
			return nil, err
		}

		for _, partition := range partitions {
			partitionRoot := fmt.Sprintf("%s/consumers/%s/offsets/%s/%s", cg.zookeeper.option.Chroot, cg.Name, topic, partition)
			val, _, err := cg.zookeeper.conn.Get(partitionRoot)
			if err != nil {
				return nil, err
			}

			partition, err := strconv.ParseInt(partition, 10, 32)
			if err != nil {
				return nil, err
			}

			offset, err := strconv.ParseInt(string(val), 10, 64)
			if err != nil {
				return nil, err
			}

			result[topic][int32(partition)] = offset
		}
	}

	return result, nil
}

func (cg *ConsumerGroup) ResetOffsets() error {
	offsetsRoot := fmt.Sprintf("%s/consumers/%s/offsets", cg.zookeeper.option.Chroot, cg.Name)
	topics, _, err := cg.zookeeper.conn.Children(offsetsRoot)
	if err == zk.ErrNoNode {
		return nil
	} else if err != nil {
		return err
	}

	for _, topic := range topics {
		topicRoot := fmt.Sprintf("%s/consumers/%s/offsets/%s", cg.zookeeper.option.Chroot, cg.Name, topic)
		partitions, stat, err := cg.zookeeper.conn.Children(topicRoot)
		if err != nil {
			return err
		}

		for _, partition := range partitions {
			partitionRoot := fmt.Sprintf("%s/consumers/%s/offsets/%s/%s", cg.zookeeper.option.Chroot, cg.Name, topic, partition)
			exists, stat, err := cg.zookeeper.conn.Exists(partitionRoot)
			if exists {
				if err = cg.zookeeper.conn.Delete(partitionRoot, stat.Version); err != nil {
					if err != zk.ErrNoNode {
						return err
					}
				}
			}
		}

		if err := cg.zookeeper.conn.Delete(topicRoot, stat.Version); err != nil {
			if err != zk.ErrNoNode {
				return err
			}
		}
	}

	return nil
}

func (cgl ConsumerGroupList) Find(name string) *ConsumerGroup {
	for _, cg := range cgl {
		if cg.Name == name {
			return cg
		}
	}
	return nil
}

func (cgl ConsumerGroupList) Len() int {
	return len(cgl)
}

func (cgl ConsumerGroupList) Less(i, j int) bool {
	return cgl[i].Name < cgl[j].Name
}

func (cgl ConsumerGroupList) Swap(i, j int) {
	cgl[i], cgl[j] = cgl[j], cgl[i]
}

func (cgil ConsumerGroupInstanceList) Find(id string) *ConsumerGroupInstance {
	for _, cgi := range cgil {
		if cgi.ID == id {
			return cgi
		}
	}
	return nil
}

func (cgil ConsumerGroupInstanceList) Len() int {
	return len(cgil)
}

func (cgil ConsumerGroupInstanceList) Less(i, j int) bool {
	return cgil[i].ID < cgil[j].ID
}

func (cgil ConsumerGroupInstanceList) Swap(i, j int) {
	cgil[i], cgil[j] = cgil[j], cgil[i]
}
