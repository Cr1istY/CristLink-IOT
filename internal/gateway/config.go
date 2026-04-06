package gateway

import "github.com/google/uuid"

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

type TCPForwarderConfig struct {
	Addr          string `json:"addr"`      // MQTT Broker 地址
	ClientID      string `json:"client_id"` // 客户端ID - 唯一
	UserName      string `json:"username"`  // 用户名
	Password      string `json:"password"`  // 密码
	AutoReconnect bool   `json:"auto_reconnect"`
	CleanSession  bool   `json:"clean_session"`

	BufferSize int    `json:"buffer_size"` // 缓冲区大小
	TCPTarget  string `json:"tcp_target"`  // TCP 目标地址
	Topic      string `json:"topic"`       // 目标 Topic
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

// GetTCPForwarderConfig 返回 MQTT 配置
func GetTCPForwarderConfig() *TCPForwarderConfig {
	unique := "forwarder-" + uuid.New().String()
	return &TCPForwarderConfig{
		Addr:          "localhost:1883",
		ClientID:      unique,
		UserName:      "",
		Password:      "",
		AutoReconnect: true,
		CleanSession:  true,
		BufferSize:    1024,
		TCPTarget:     "localhost:9001",
		Topic:         "/sensor/#",
	}
}
