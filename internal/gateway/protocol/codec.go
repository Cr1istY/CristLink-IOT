package protocol

import (
	"bytes"
	"fmt"
)

// MVP 阶段我们不实现复杂的 MQTT 协议包解析，而是实现一个简单的 行协议 或者 JSON 透传。
// 假设设备发送格式为：DeviceID|JSONPayload (例如: dev_001|{"temp": 25})。

// Packet 解析后的数据包
type Packet struct {
	DeviceID string
	Payload  []byte // 原始业务数据
}

// Decode 解析字节流
// 假设协议格式：DeviceID + "|" + JSON数据
func Decode(data []byte) (*Packet, error) {
	// 简单查找分隔符
	idx := bytes.IndexByte(data, '|')
	if idx == -1 {
		return nil, fmt.Errorf("invalid format: missing separator")
	}

	deviceID := string(data[:idx])
	payload := data[idx+1:]

	// 这里可以添加 JSON 校验逻辑，MVP 阶段直接透传
	return &Packet{
		DeviceID: deviceID,
		Payload:  payload,
	}, nil
}
