package protocol

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

func init() {
	Register("mqtt", &MQTTCodec{})
}

type MQTTCodec struct {
}

// Decode 解析 MQTT 消息
// 注意：在 gnet 中，如果使用了 MQTT 的 Frame Codec，
// msg 参数通常就是 MQTT 的 Payload (即数据体部分)
// meta 参数中通常包含 Topic 信息
func (c *MQTTCodec) Decode(src []byte, meta map[string]string) (*StandardPayload, error) {
	// 1. 从元数据中提取 Topic
	topic, ok := meta["topic"]
	if !ok {
		return nil, errors.New(ErrMissingTopicInMetadata)
	}
	// 2. 解析 Topic，
	// Topic 通常包含路由信息（如 /sys/{pk}/{dk}/up）

	pk, dk, err := parseTopic(topic)
	if err != nil {
		return nil, err
	}

	// 3. 构造Payload
	payload := NewStandardPayload(pk, dk)

	// 4. 解析尝试
	var inputMsg struct {
		Method    string                 `json:"method"`
		Timestamp int64                  `json:"ts"`
		Data      map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(src, &inputMsg); err != nil {
		// 如果解析失败，可能是纯文本或二进制，我们可以把它作为整体放入 data 字段，或者报错
		// 这里演示：如果解析失败，将整个 payload 当作 string 放入 data["raw"]
		payload.SetData("raw_payload", string(src))
	} else {
		// 如果解析成功，则填充 payload
		if inputMsg.Method != "" {
			payload.Method = inputMsg.Method
		} else {
			payload.Method = MethodReport
		}
		if inputMsg.Timestamp > 0 {
			payload.Timestamp = inputMsg.Timestamp
		}
		for k, v := range inputMsg.Data {
			payload.SetData(k, v)
		}
	}
	// 5. 填充 MQTT 特有的元数据
	payload.SrcProtocol = "mqtt"
	// 如果 Topic 里还有额外层级，比如 /sys/pk/dk/attribute/up，可以把 "attribute" 也记录下来
	return payload, nil
}

// Encode 用于云端下发指令给设备
// 将 StandardPayload 转换回 MQTT 协议能识别的格式
func (c *MQTTCodec) Encode(payload *StandardPayload) ([]byte, error) {
	response := map[string]interface{}{
		"code": payload.Code,
		"msg":  payload.Message,
		"data": payload.Values,
		"ts":   time.Now().UnixNano() / 1e6,
	}
	return json.Marshal(response)
}

func parseTopic(topic string) (string, string, error) {
	topic = strings.TrimSpace(topic)
	parts := strings.Split(topic, "/")

	// "/sys/{pk}/{dk}/up"
	// parts[0] = "" (因为是 / 开头)
	// parts[1] = "sys"
	// parts[2] = {pk}
	// parts[3] = {dk}
	// parts[4] = "up"

	if len(parts) < 5 {
		return "", "", errors.New(ErrTopicTooShort)
	}
	if parts[1] != "sys" {
		return "", "", errors.New(ErrInvalidTopicPrefix)
	}

	productKey := parts[2]
	deviceKey := parts[3]

	return productKey, deviceKey, nil
}
