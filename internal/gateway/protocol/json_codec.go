package protocol

import (
	"encoding/json"
)

func init() {
	Register("json", &JSONCodec{})
}

type JSONCodec struct{}

func (c *JSONCodec) Decode(src []byte, meta map[string]string) (*StandardPayload, error) {
	payload := NewStandardPayload(meta["product_key"], meta["device_key"])

	// 临时结构体，用来处理 原始JSON
	var rawInput struct {
		Method    string                 `json:"method"`
		Timestamp int64                  `json:"ts"`
		Data      map[string]interface{} `json:"data"`
		// 兼容扁平化数据，即，若无data字段，则 JSON整体作为data字段
		FallbackData map[string]interface{} `json:"-"`
	}

	_ = json.Unmarshal(src, &rawInput)
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
			for k, v := range flatData {
				if k != "method" && k != "ts" && k != "data" {
					payload.SetData(k, v)
				}
			}
		}
	}

	// 一般情况下，rawInput.Method 应该为空
	if rawInput.Method != "" {
		payload.Method = rawInput.Method
	} else {
		payload.Method = MethodReport
	}

	if rawInput.Timestamp > 0 {
		payload.Timestamp = rawInput.Timestamp
	}

	return payload, nil

}

func (c *JSONCodec) Encode(payload *StandardPayload) ([]byte, error) {
	// 下行数据通常直接序列化
	return json.Marshal(payload)
}
