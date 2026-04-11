package protocol

import (
	"CristLink-IoT/internal/logger"
	"encoding/json"
	"time"

	"go.uber.org/zap"
)

// 用来分析由其他 MQTT Broker 转发的消息
// 参考格式 mosquitto_pub -h localhost -p 1883 -t "sensor/sys/ws/sdt/up" -m '{"id":"dev_001","temp":25.6,"humidity":60,"status":"active"}'
// 其中 pk 为 ws，dk 为 sdt

func init() {
	Register("mqtt_consumer", &MQTTConsumerCodec{})
	// 启动 MQTT Broker
}

type MQTTConsumerCodec struct{}

func (c *MQTTConsumerCodec) Decode(src []byte, meta Meta) (*StandardPayload, error) {
	var messageData struct {
		Topic string `json:"topic"`
		Data  []byte `json:"data"`
	}
	err := json.Unmarshal(src, &messageData)
	if err != nil {
		logger.Logger.Error("JSON Unmarshal failed", zap.Error(err))
		return nil, err
	}
	pk, dk, err := parseTopic(messageData.Topic)
	if err != nil || pk == "" || dk == "" {
		logger.Logger.Error("get topic failed", zap.Error(err))
		return nil, err
	}

	// 进行赋值操作
	payload := NewStandardPayload(pk, dk)

	var rawData map[string]interface{}
	err = json.Unmarshal(messageData.Data, &rawData)
	if err != nil {
		logger.Logger.Error("JSON Unmarshal failed", zap.Error(err))
		return nil, err
	}

	for k, v := range rawData {
		payload.SetData(k, v)
	}

	payload.Method = MethodReport

	if payload.Timestamp < 0 {
		payload.Timestamp = time.Now().UnixMilli()
	}

	if payload.EventType == "" {
		payload.EventType = EventTypeProperty
	}

	if payload.Seq < 0 {
		// 当指定时间类型，却未指定时间时序时，自动指定
		// 冗余设计
		switch payload.EventType {
		case EventTypeProperty:
			payload.Seq = 0
		case EventTypeEvent:
			payload.Seq = 100
		case EventTypeService:
			payload.Seq = 50
		default:
			payload.Seq = 0
		}
	}

	payload.MsgID = meta.MsgID
	return payload, nil

}

func (c *MQTTConsumerCodec) Encode(payload *StandardPayload) ([]byte, error) {
	// TODO: 实现编码逻辑
	return json.Marshal(payload)
}
