package main

import (
	"CristLink-IoT/internal/gateway"
	"CristLink-IoT/internal/gateway/network"
	"CristLink-IoT/internal/gateway/sink"
	"fmt"
	"log"

	"github.com/panjf2000/gnet/v2"
)

func main() {
	// 1. 加载配置
	cfg := gateway.GetConfig()

	// 2. 初始化 Kafka 生产者
	producer := sink.NewProducer(cfg.KafkaAddr, cfg.KafkaTopic)
	defer func(producer *sink.Producer) {
		err := producer.Close()
		if err != nil {

		}
	}(producer)

	configs := []gateway.ServerConfig{
		{
			Port:         9000,
			ProtocolType: "json", // JSON 协议服务
			Name:         "JSON_Gateway",
		},
		{
			Port:         9001,
			ProtocolType: "mqtt", // Modbus 协议服务
			Name:         "Mqtt_Gateway",
		},
	}

	// 3. 工厂循环启动 (核心逻辑)
	for _, cfg := range configs {
		// 使用工厂创建实例
		server, err := network.CreateServer(cfg, producer)
		if err != nil {
			log.Printf("Failed to create server for %s: %v", cfg.Name, err)
			continue
		}

		// 启动服务
		addr := fmt.Sprintf("tcp://:%d", cfg.Port)
		log.Printf("Starting %s on %s", cfg.Name, addr)
		// 注意：gnet.Run 是阻塞的，如果要在主线程同时运行多个，需要开 Goroutine
		go func(s *network.GatewayServer, a string) {
			// gnet.Run 接收实现了 gnet.Events 接口的对象
			// 因为 *GatewayServer 实现了 OnTraffic/OnOpen 等方法，所以它天然满足接口要求
			if err := gnet.Run(s, a); err != nil {
				log.Printf("Server error: %v", err)
			}
		}(server, addr)
	}

	// 阻塞主线程
	select {}

}
