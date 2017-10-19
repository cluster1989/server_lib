package libzookeeper

type KafkaBrokerModel struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type KakfaControllerModel struct {
	BrokerID int32 `json:"brokerid"`
}

type KafkaTopicMetaModel struct {
	Partitions map[string][]int32 `json:"partitions"`
}

type KafkaTopicConfigModel struct {
	ConfigMap map[string]string `json:"config"`
}

type KafkaPartitionStateModel struct {
	Leader int32   `json:"leader"`
	ISR    []int32 `json:"isr"`
}

type KafkaConsumerRegistryModel struct {
	Pattern      string         `json:"pattern"`
	Subscription map[string]int `json:"subscription"`
	Timestamp    int64          `json:"timestamp"`
	Version      int            `json:"version"`
}
