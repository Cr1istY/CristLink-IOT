package main

import (
	"CristLink-IoT/internal/gateway"
	"CristLink-IoT/internal/gateway/network"
	"CristLink-IoT/internal/gateway/sink"
	"CristLink-IoT/internal/logger"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/panjf2000/gnet/v2"
)

func main() {
	// 开启 zap
	if err := logger.InitLogger(); err != nil {
		panic(err)
	}

	defer func() {
		if err := logger.Logger.Sync(); err != nil {
			log.Printf("Failed to flush logger: %v", err)
		}
	}()
	logger.Info("Logger started")

	// 加载 Kafka 配置
	kafkaCfg := gateway.GetKafkaConfig()
	// 初始化 Kafka 生产者
	producer := sink.NewProducer(kafkaCfg.KafkaAddr, kafkaCfg.KafkaTopic)

	logger.Info("Kafka Producer init", "addr", kafkaCfg.KafkaAddr)

	defer func(producer *sink.Producer) {
		err := producer.Close()
		if err != nil {
			logger.Error(err.Error())
		}
	}(producer)

	tcpForwarderConfig := gateway.GetTCPForwarderConfig()
	tcpForwarder := sink.NewTCPForwarder(tcpForwarderConfig)
	if tcpForwarder == nil {
		logger.Error("TCP Forwarder init failed")
		return
	}

	tcpForwarder.Start()

	configs := []gateway.ServerConfig{
		{
			Port:         9000,
			ProtocolType: "json", // JSON 协议服务
			Name:         "JSON_Gateway",
		},
		{
			Port:         9001,
			ProtocolType: "mqtt", // MQTT 协议服务
			Name:         "Mqtt_Gateway",
		},
		{
			Port:         9020,
			ProtocolType: "mqtt_consumer", // MQTT 协议服务
			Name:         "Mqtt_consumer_Gateway",
		},
	}

	// 工厂循环启动 (核心逻辑)
	for _, cfg := range configs {
		// 使用工厂创建实例
		server, err := network.CreateServer(cfg, producer)
		if err != nil {
			logger.Error("Failed to create server ", "server_name", cfg.Name, "error", err)
			continue
		}

		// 启动服务
		addr := fmt.Sprintf("tcp://0.0.0.0:%d", cfg.Port)
		logger.Info("Starting gnet server", "server_name", cfg.Name, "server_addr", addr)
		// 注意：gnet.Run 是阻塞的，如果要在主线程同时运行多个，需要开 Goroutine
		go func(s *network.GatewayServer, a string) {
			// gnet.Run 接收实现了 gnet.Events 接口的对象
			// 因为 *GatewayServer 实现了 OnTraffic/OnOpen 等方法，所以它天然满足接口要求
			if err := gnet.Run(s, a, gnet.WithMulticore(true)); err != nil {
				logger.Error("Server error", "error", err)
			}
		}(server, addr)
	}

	// 阻塞主线程
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	receivedSignal := <-sigChan
	fmt.Printf("\n\nReceived signal: %v. Shutting down gracefully...\n", receivedSignal)
	tcpForwarder.Stop()

	fmt.Println("Program exited.")
}
