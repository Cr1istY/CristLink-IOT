package protocol

import "time"

// StandardPayload 是所有协议解析后的统一格式
// 符合 crist-link 项目“云边协同”与“统一治理”架构
type StandardPayload struct {
	// 1. 设备标识 (Identity)
	ProductKey string `json:"pk"` // 产品Key，用于快速索引物模型 (对应文档三元组)
	DeviceKey  string `json:"dk"` // 设备Key (原device_id)，唯一标识

	// 2. 消息元数据 (Metadata)
	MsgID     string `json:"msgid"` // 消息唯一ID，用于防重放攻击与链路追踪
	Timestamp int64  `json:"ts"`    // 时间戳 (毫秒)，建议统一为UTC
	Seq       int64  `json:"seq"`   // 序列号，用于消息排序 越大优先级越高

	// 3. 通信指令 (Method)
	// 类似 HTTP Method，区分云端下发指令与设备上报
	Method string `json:"method,omitempty"` // "REPORT", "SET", "GET", "INVOKE"

	// 4. 事件与数据 (Payload)
	EventType string                 `json:"type"` // "property", "event", "service", "alarm"
	Values    map[string]interface{} `json:"data"` // 具体业务数据

	// 5. 状态与扩展 (Status & Ext)
	// 仅在云端回复设备时填充 (如 SET 指令的 ACK)
	Code    int    `json:"code,omitempty"` // 状态码 200 成功, 404 参数错误等
	Message string `json:"msg,omitempty"`  // 错误信息

	// 6. 边缘相关 (Edge)
	// 如果是边缘网关上报，此处标记原始协议
	SrcProtocol string `json:"proto,omitempty"` // "coap", "modbus", "opcua"
	GatewayID   string `json:"gid,omitempty"`   // 边缘网关ID (如果是直连设备为空)
}

type Meta struct {
	SourceIp   string `json:"source_ip"`  // 源IP
	LocalAddr  string `json:"local_addr"` // 本地端口
	ProductKey string `json:"pk"`         // 产品Key，用于快速索引物模型 (对应文档三元组)
	DeviceKey  string `json:"dk"`         // 设备Key (原device_id)，唯一标识
	MsgID      string `json:"msgid"`      // 消息唯一ID，用于防重放攻击与链路追踪
	Timestamp  int64  `json:"ts"`         // 时间戳 (毫秒)，建议统一为UTC
	Seq        int64  `json:"seq"`        // 序列号，用于消息排序

	Topic string `json:"topic"` // MQTT Topic
}

// --- 辅助常量与方法 ---

const (
	// EventType 定义
	EventTypeProperty = "property" // 属性上报
	EventTypeEvent    = "event"    // 事件上报
	EventTypeService  = "service"  // 服务调用结果

	// Method 定义
	MethodReport = "REPORT" // 设备主动上报
	MethodSet    = "SET"    // 云端下发设置
	MethodGet    = "GET"    // 云端查询
)

// NewStandardPayload 快速构建一个标准Payload
func NewStandardPayload(productKey, deviceKey string) *StandardPayload {
	return &StandardPayload{
		ProductKey: productKey,
		DeviceKey:  deviceKey,
		Timestamp:  time.Now().UnixNano() / 1e6, // 默认当前时间
		Values:     make(map[string]interface{}),
	}
}

// SetData 添加业务数据
func (p *StandardPayload) SetData(key string, value interface{}) {
	if p.Values == nil {
		p.Values = make(map[string]interface{})
	}
	p.Values[key] = value
}
