package protocol

import (
	"encoding/json"
	"errors"
	"time"
)

func init() {
	Register("json", &JSONCodec{})
}

type JSONCodec struct{}

func (c *JSONCodec) Decode(src []byte, meta Meta) (*StandardPayload, error) {
	// 临时结构体，用来处理 原始JSON
	var rawInput struct {
		ProductKey string                 `json:"pk"` // 产品Key，用于快速索引物模型 (对应文档三元组)
		DeviceKey  string                 `json:"dk"` // 设备Key (原device_id)，唯一标识
		Method     string                 `json:"method"`
		EventType  string                 `json:"type"` // "property", "event", "service", "alarm"
		Timestamp  int64                  `json:"ts"`
		Seq        int64                  `json:"seq"`
		Data       map[string]interface{} `json:"data"`
		// 兼容扁平化数据，即，若无data字段，则 JSON整体作为data字段
		FallbackData map[string]interface{} `json:"-"`
	}

	_ = json.Unmarshal(src, &rawInput)
	// 情况 A：设备发了标准的 {"data": {...}}
	// 此时，理应解析 ProductKey 和 DeviceKey
	if rawInput.ProductKey == "" || rawInput.DeviceKey == "" {
		return nil, errors.New(ErrMissingProductKeyOrDeviceKey)
	}
	meta.ProductKey = rawInput.ProductKey
	meta.DeviceKey = rawInput.DeviceKey

	payload := NewStandardPayload(meta.ProductKey, meta.DeviceKey)
	if len(rawInput.Data) > 0 {
		// 情况 A：设备发了标准的 {"data": {...}}
		for k, v := range rawInput.Data {
			payload.SetData(k, v)
		}
	} else {
		// 情况 B：设备发了扁平数据 {"temp": 25}，直接解析到 payload 会丢数据
		// 我们重新解析一遍到 map 里，全部塞进 Values
		var flatData map[string]interface{}
		if err := json.Unmarshal(src, &flatData); err == nil {
			protocolKeys := map[string]bool{
				"pk": true, "dk": true, "method": true, "type": true,
				"ts": true, "seq": true, "data": true,
			}
			for k, v := range flatData {
				if !protocolKeys[k] {
					payload.SetData(k, v)
				}
			}
		}
	}

	payload.Method = MethodReport

	if rawInput.Timestamp > 0 {
		payload.Timestamp = rawInput.Timestamp
	} else {
		payload.Timestamp = time.Now().UnixMilli()
	}

	if rawInput.Seq > 0 {
		payload.Seq = rawInput.Seq
	}

	if rawInput.EventType != "" {
		payload.EventType = rawInput.EventType
	}

	if rawInput.Seq < 0 {
		// 当指定时间类型，却未指定时间时序时，自动指定
		// 冗余设计
		switch rawInput.EventType {
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

func (c *JSONCodec) Encode(payload *StandardPayload) ([]byte, error) {
	// 下行数据通常直接序列化
	return json.Marshal(payload)
}
