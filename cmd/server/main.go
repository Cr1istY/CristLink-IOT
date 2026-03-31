package main

import (
	"CristLink-IoT/internal/gateway"
	"CristLink-IoT/internal/gateway/network"
	"CristLink-IoT/internal/gateway/sink"
	"strconv"

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

	// 3. 初始化网络服务
	server := network.NewGatewayServer(producer)

	// 4. 启动服务
	log.Printf("🚀 Gateway MVP starting on port %d...", cfg.Port)
	err := gnet.Run(server,
		"tcp://0.0.0.0:"+strconv.Itoa(cfg.Port), // 监听地址
		gnet.WithMulticore(true),                // 开启多核模式
		gnet.WithReusePort(true),                // 端口复用
		gnet.WithTicker(true),
	)

	if err != nil {
		log.Fatalf("Gateway crashed: %v", err)
	}
}
