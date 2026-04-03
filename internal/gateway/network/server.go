package network

import (
	"CristLink-IoT/internal/gateway/protocol"
	"CristLink-IoT/internal/gateway/sink"
	"CristLink-IoT/internal/logger"
	"context"
	"encoding/json"

	"github.com/matoous/go-nanoid/v2"
	"github.com/panjf2000/gnet/v2"
	"go.uber.org/zap"
)

// GatewayServer 网关服务结构
// 只负责处理逻辑，不负责创建自己
type GatewayServer struct {
	gnet.BuiltinEventEngine // 内嵌 gnet 事件引擎
	kafkaProducer           *sink.Producer
	ctx                     context.Context
	protocolType            string
}

// OnOpen 当新连接建立时触发
func (gs *GatewayServer) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	// 1. 获取真实的客户端 IP 和 端口
	remoteAddr := c.RemoteAddr().String()
	localAddr := c.LocalAddr().String()
	meta := protocol.Meta{
		SourceIp:  remoteAddr,
		LocalAddr: localAddr,
	}
	meta.MsgID, _ = gonanoid.New(16)
	c.SetContext(meta)
	logger.Logger.Info("New connection", zap.String("remoteAddr", remoteAddr), zap.String("type", gs.protocolType))
	// 可以在这里做简单的握手鉴权，MVP 暂略
	return nil, gnet.None
}

// OnClose 连接断开时触发
func (gs *GatewayServer) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	logger.Logger.Info("Connection closed", zap.String("remoteAddr", c.RemoteAddr().String()), zap.String("type", gs.protocolType))
	return gnet.None
}

func (gs *GatewayServer) OnTraffic(c gnet.Conn) (action gnet.Action) {
	// 准备元数据 (辅助信息)
	meta, ok := c.Context().(protocol.Meta)
	if !ok {
		return gnet.Close
	}

	// Peek(-1) 表示读取所有可用数据，但不消费（不移除）
	buf, _ := c.Peek(-1)

	// 必须拷贝一份，因为 Peek 返回的内存会被 gnet 复用，异步操作（如 Kafka）会有并发问题
	data := make([]byte, len(buf))
	copy(data, buf)

	// 消费数据
	// 这一步非常重要！必须告诉 gnet 你已经处理完了，否则缓冲区会爆满
	_, _ = c.Discard(len(buf))

	// 解析协议
	// 获取对应的解析器
	codec, err := protocol.GetCodec(gs.protocolType)
	if err != nil {
		logger.Error("获取解析器失败", "error", err)
		return gnet.Close
	}

	// 执行解码
	payload, err := codec.Decode(data, meta)
	if err != nil {
		logger.Error("协议解码失败", "error", err)
		return gnet.Close
	}

	// 此时 payload 已经是标准的 StandardPayload 了
	// 接下来推送到 Kafka
	go func() {
		key := []byte(payload.DeviceKey)
		valueBytes, err := json.Marshal(payload)
		if err != nil {
			logger.Error("Kafka 消息序列化失败", "error", err)
			return
		}
		if err := gs.kafkaProducer.Send(gs.ctx, key, valueBytes); err != nil {
			logger.Error("Failed to send to kafka", "error", err)
		}
	}()

	return gnet.None
}
