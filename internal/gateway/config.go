package gateway

// Config 定义网关配置
type Config struct {
	Port       int    `json:"port"`        // 监听端口
	KafkaTopic string `json:"kafka_topic"` // 目标 Kafka Topic
	KafkaAddr  string `json:"kafka_addr"`  // Kafka 地址
}

// GetConfig 返回默认配置 (MVP 阶段直接写死或简单读取 env)
func GetConfig() *Config {
	return &Config{
		Port:       9000,
		KafkaTopic: "iot_data_raw",
		KafkaAddr:  "localhost:9092", // 生产环境请读取环境变量
	}
}
