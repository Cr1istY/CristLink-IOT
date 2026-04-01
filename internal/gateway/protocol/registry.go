package protocol

import "fmt"

// 全局注册表
var codecs = make(map[string]Codec)

// Register 注册协议解析器
func Register(name string, codec Codec) {
	if _, exists := codecs[name]; exists {
		panic("codec already registered: " + name)
	}
	codecs[name] = codec
}

// GetCodec 获取协议解析器
func GetCodec(name string) (Codec, error) {
	codec, ok := codecs[name]
	if !ok {
		return nil, fmt.Errorf("codec not found: %s", name)
	}
	return codec, nil
}
