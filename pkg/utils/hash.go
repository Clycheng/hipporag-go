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

// ComputeHash 计算字符串的 MD5 哈希（带前缀）
func ComputeHash(content string, prefix string) string {
	h := sha256.New()
	h.Write([]byte(content))
	return prefix + hex.EncodeToString(h.Sum(nil))
}

// MinMaxNormalize 将分数归一化到 [0, 1] 区间
func MinMaxNormalize(scores []float64) []float64 {
	if len(scores) == 0 {
		return scores
	}

	// 找到最小值和最大值
	minScore := scores[0]
	maxScore := scores[0]
	for _, score := range scores {
		if score < minScore {
			minScore = score
		}
		if score > maxScore {
			maxScore = score
		}
	}

	// 如果所有分数相同，返回全 1
	if maxScore == minScore {
		normalized := make([]float64, len(scores))
		for i := range normalized {
			normalized[i] = 1.0
		}
		return normalized
	}

	// 归一化
	normalized := make([]float64, len(scores))
	for i, score := range scores {
		normalized[i] = (score - minScore) / (maxScore - minScore)
	}

	return normalized
}
