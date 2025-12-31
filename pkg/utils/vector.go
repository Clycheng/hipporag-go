package utils

// vector.go - 向量计算工具
// 用途：向量相似度计算、向量归一化等操作
// 主要功能：
// - CosineSimilarity: 计算两个向量的余弦相似度
// - Normalize: 向量归一化（L2范数）
// - DotProduct: 向量点积

import (
	"math"
)

// CosineSimilarity 计算两个向量的余弦相似度
// 返回值范围 [-1, 1]，值越大表示越相似
func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	dotProduct := 0.0
	normA := 0.0
	normB := 0.0

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// DotProduct 计算两个向量的点积
func DotProduct(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	result := 0.0
	for i := 0; i < len(a); i++ {
		result += a[i] * b[i]
	}

	return result
}

// Normalize 对向量进行 L2 归一化
// 返回归一化后的新向量
func Normalize(v []float64) []float64 {
	norm := 0.0
	for _, val := range v {
		norm += val * val
	}
	norm = math.Sqrt(norm)

	if norm == 0 {
		return v
	}

	result := make([]float64, len(v))
	for i, val := range v {
		result[i] = val / norm
	}

	return result
}
