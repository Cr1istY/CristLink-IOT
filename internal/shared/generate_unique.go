package shared

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateUniqueID 基于字符串和时间生成唯一的 SHA256 哈希值
func GenerateUniqueID(str1, str2 string) string {
	// 1. 拼接字符串：建议加上分隔符，防止 "a"+"bc" 和 "ab"+"c" 混淆
	// 加上纳秒级时间戳，保证即使同一毫秒请求也不重复
	raw := fmt.Sprintf("%s|%s|%d", str1, str2, time.Now().UnixNano())

	// 2. 计算 SHA256
	sum := sha256.Sum256([]byte(raw))

	// 3. 转换为十六进制字符串
	return fmt.Sprintf("%x", sum)
}

func GenerateShortMsgID(pk, dk string) string {
	raw := fmt.Sprintf("%s|%s|%d", pk, dk, time.Now().UnixNano())
	sum := sha256.Sum256([]byte(raw))

	// 只取前 12 个字节（24 个十六进制字符）
	// 24 字符对于去重来说已经足够安全（碰撞概率约为 2^-96）
	return hex.EncodeToString(sum[:12])
}
