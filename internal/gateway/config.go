package gateway

// KafkaConfig 定义网关配置
type KafkaConfig struct {
	Port       int    `json:"port"`        // 监听端口
	KafkaTopic string `json:"kafka_topic"` // 目标 Kafka Topic
	KafkaAddr  string `json:"kafka_addr"`  // Kafka 地址
}

// ServerConfig 定义了一个网关服务实例的配置
type ServerConfig struct {
	// Port 监听端口
	Port int
	// ProtocolType 协议类型标识，用于选择解析器
	ProtocolType string
	// Name 服务名称 (可选)
	Name string
}

type MqttBrokerConfig struct {
	Addr     string `json:"addr"`      // MQTT Broker 地址
	ClientID string `json:"client_id"` // 客户端ID - 唯一
	UserName string `json:"username"`  // 用户名
	Password string `json:"password"`  // 密码
}

// TODO: 从 env 文件获取

// GetKafkaConfig 返回 Kafka 默认配置
func GetKafkaConfig() *KafkaConfig {
	return &KafkaConfig{
		Port:       9000,
		KafkaTopic: "iot_data_raw",
		KafkaAddr:  "localhost:9092", // 生产环境请读取环境变量
	}
}

// GetMqttBrokerConfig 返回 MQTT 默认配置
func GetMqttBrokerConfig() *MqttBrokerConfig {
	return &MqttBrokerConfig{
		Addr:     "tcp://localhost:1883",
		ClientID: "gateway",
		UserName: "",
		Password: "",
	}
}
