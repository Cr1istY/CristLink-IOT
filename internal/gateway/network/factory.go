package network

import (
	"CristLink-IoT/internal/gateway"
	"CristLink-IoT/internal/gateway/sink"
	"context"
)

// ServerFactory 是一个函数类型，根据配置创建一个 gnet 中的 EventEngine 实例
type ServerFactory func(cfg gateway.ServerConfig, producer *sink.Producer) (*GatewayServer, error)

func CreateServer(cfg gateway.ServerConfig, producer *sink.Producer) (*GatewayServer, error) {
	server := &GatewayServer{
		kafkaProducer: producer,
		ctx:           context.Background(),
		protocolType:  cfg.ProtocolType,
	}
	return server, nil
}
