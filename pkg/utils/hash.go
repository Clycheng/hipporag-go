package utils

// hash.go - 内容哈希工具
// 用途：为文本内容生成唯一的哈希ID，用于去重和快速查找
// 主要功能：
// - Hash: 对字符串生成 SHA256 哈希值
// - 用于检测重复文档块、实体等

import (
	"crypto/sha256"
	"encoding/hex"
)

// Hash 对字符串生成 SHA256 哈希值
// 返回十六进制字符串形式的哈希值
func Hash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// HashString 是 Hash 的别名，保持向后兼容
func HashString(s string) string {
	return Hash(s)
}
