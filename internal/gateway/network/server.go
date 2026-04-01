package network

import (
	"context"
	"encoding/json"
	"log"

	"CristLink-IoT/internal/gateway/protocol"
	"CristLink-IoT/internal/gateway/sink"

	"github.com/panjf2000/gnet/v2"
)

// GatewayServer 网关服务结构
type GatewayServer struct {
	gnet.BuiltinEventEngine // 内嵌 gnet 事件引擎
	kafkaProducer           *sink.Producer
	ctx                     context.Context
}

// NewGatewayServer 创建服务实例
func NewGatewayServer(producer *sink.Producer) *GatewayServer {
	return &GatewayServer{
		kafkaProducer: producer,
		ctx:           context.Background(),
	}
}

// OnOpen 当新连接建立时触发
func (gs *GatewayServer) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	log.Printf("New connection: %s", c.RemoteAddr().String())
	// 可以在这里做简单的握手鉴权，MVP 暂略
	return nil, gnet.None
}

// OnClose 连接断开时触发
func (gs *GatewayServer) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	log.Printf("Connection closed: %s", c.RemoteAddr().String())
	return gnet.None
}

// OnTraffic 处理接收到的数据 (核心逻辑)
func (gs *GatewayServer) OnTraffic(c gnet.Conn) (action gnet.Action) {
	// 1. 读取数据 (gnet 使用 RingBuffer，这里直接取全部)
	buf, _ := c.Peek(-1)
	data := make([]byte, len(buf))
	copy(data, buf)
	_, _ = c.Discard(len(buf)) // 消费掉缓冲区数据

	// 2. 协议解析
	packet, err := protocol.Decode(data)
	if err != nil {
		log.Printf("Decode error: %v", err)
		return gnet.None
	}

	// 3. 构造投递消息 (添加设备ID信息)
	type Message struct {
		DeviceID string `json:"device_id"`
		Data     string `json:"data"`
	}
	msg := Message{
		DeviceID: packet.DeviceID,
		Data:     string(packet.Payload),
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("JSON Marshal error: %v", err)
		return // 或者处理错误
	}

	// 4. 异步投递到 Kafka (注意：生产环境应使用 Channel 缓冲，避免阻塞 IO 协程)
	// 这里为了 MVP 简单直接调用，实际建议使用 goroutine 或 channel
	go func() {
		if err := gs.kafkaProducer.Send(gs.ctx, []byte(packet.DeviceID), msgBytes); err != nil {
			log.Printf("Failed to send to kafka: %v", err)
		}
	}()

	return gnet.None
}

func (gs *GatewayServer) OnMessage(c gnet.Conn, msg []byte) (out []byte, action gnet.Action) {
	// 1. 获取设备对应的协议类型
	// 这一步通常通过 Session 或 设备配置中心 获取
	// 假设我们通过 Topic 或端口区分，这里硬编码演示
	protocolType := "json" // 或者 "modbus"

	// 2. 获取对应的解析器
	codec, err := protocol.GetCodec(protocolType)
	if err != nil {
		log.Printf("获取解析器失败: %v", err)
		return nil, gnet.Close
	}

	// 3. 准备元数据 (辅助信息)
	meta := map[string]string{
		"product_key": "pk_123",
		"device_id":   "dev_001",
		"source_ip":   c.RemoteAddr().String(),
	}

	// 4. 执行解码
	payload, err := codec.Decode(msg, meta)
	if err != nil {
		log.Printf("协议解码失败: %v", err)
		return nil, gnet.Close
	}

	// 5. 此时 payload 已经是标准的 StandardPayload 了
	// 接下来推送到 Kafka
	go func() {
		key := []byte(payload.DeviceKey)
		valueBytes, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Kafka 消息序列化失败: %v", err)
			return
		}
		if err := gs.kafkaProducer.Send(gs.ctx, key, valueBytes); err != nil {
			log.Printf("Failed to send to kafka: %v", err)
		}
	}()

	return nil, gnet.None
}
