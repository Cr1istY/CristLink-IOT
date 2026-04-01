package protocol

// 使用工厂模式
// 未来的所有协议解析器，都应该实现这个接口

// Codec 定义了协议编解码器的标准接口
type Codec interface {
	// Decode 将原始字节流解析为标准载荷
	// src: 原始数据 (如 Modbus Hex, MQTT Payload)
	// meta: 元数据 (如 SourceIP, Topic, DeviceID)，辅助解析
	Decode(src []byte, meta map[string]string) (*StandardPayload, error)

	// Encode 将标准载荷封装为设备可识别的字节流
	// payload: 标准业务数据
	Encode(payload *StandardPayload) ([]byte, error)
}
