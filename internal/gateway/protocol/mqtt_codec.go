package protocol

import (
	"CristLink-IoT/internal/logger"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"go.uber.org/zap"
)

func init() {
	Register("mqtt", &MQTTCodec{})
}

type MQTTCodec struct {
	buf  *bytes.Buffer
	once sync.Once
}

// Decode 解析 MQTT 消息
// 注意：在 gnet 中，如果使用了 MQTT 的 Frame Codec，
// msg 参数通常就是 MQTT 的 Payload (即数据体部分)
// meta 参数中通常包含 Topic 信息
func (c *MQTTCodec) Decode(src []byte, meta Meta) (*StandardPayload, error) {
	if c.buf == nil {
		c.once.Do(func() {
			c.buf = new(bytes.Buffer)
		})
	}

	c.buf.Write(src) // 写入缓存区
	for c.buf.Len() > 0 {
		// 尝试解析第一个包
		pkt, err := packets.ReadPacket(c.buf)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				break
			}
			return nil, err
		}
		if pub, ok := pkt.(*packets.PublishPacket); ok {
			topic := pub.TopicName
			if topic == "" {
				return nil, errors.New(ErrMissingTopicInMetadata)
			}

			pk, dk, err := parseTopic(topic)
			if err != nil {
				return nil, err
			}

			payload := NewStandardPayload(pk, dk)
			var inputMsg struct {
				Method    string                 `json:"method"`
				Timestamp int64                  `json:"ts"`
				Data      map[string]interface{} `json:"data"`
			}

			if err := json.Unmarshal(pub.Payload, &inputMsg); err != nil {
				payload.SetData("raw_payload", string(pub.Payload))
			} else {
				if inputMsg.Method != "" {
					payload.Method = inputMsg.Method
				}
				if inputMsg.Timestamp > 0 {
					payload.Timestamp = inputMsg.Timestamp
				}
				for k, v := range inputMsg.Data {
					payload.SetData(k, v)
				}
			}
			payload.Method = MethodReport
			payload.SrcProtocol = "mqtt"
			logger.Logger.Debug("get mqtt", zap.String("topic", topic), zap.Any("payload", pub))
			return payload, nil
		}
	}

	return nil, nil
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
