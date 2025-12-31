package embedding

// interface.go - 向量存储接口定义
// 用途：定义统一的向量存储接口，支持不同的实现（内存、Weaviate 等）
// 主要功能：
// - VectorStore 接口：插入、搜索、获取等操作

import "context"

// VectorStore 向量存储接口
// 可以有多种实现：
// - Store: 内存存储（适合小规模、快速原型）
// - WeaviateStore: Weaviate 数据库（适合生产环境）
type VectorStore interface {
	// Insert 插入文本并生成向量
	// 返回：每个文本对应的 ID
	Insert(ctx context.Context, texts []string) ([]string, error)

	// Search 向量相似度搜索
	// 返回：ID 列表和对应的相似度分数
	Search(ctx context.Context, queryVec []float64, topK int) ([]string, []float64, error)

	// Get 根据 ID 获取向量
	Get(ctx context.Context, id string) ([]float64, error)

	// GetContent 根据 ID 获取原始内容
	GetContent(ctx context.Context, id string) (string, error)
}
