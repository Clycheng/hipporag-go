package embedding

// client.go - Embedding 客户端接口定义
// 用途：定义统一的 embedding 接口，支持不同的 embedding 服务（OpenAI、本地模型等）
// 主要功能：
// - Client 接口：定义 Embed 方法，将文本转换为向量

import "context"

// Client Embedding 客户端接口
type Client interface {
	// Embed 将文本转换为向量
	Embed(ctx context.Context, texts []string) ([][]float64, error)

	// EmbedSingle 将单个文本转换为向量
	EmbedSingle(ctx context.Context, text string) ([]float64, error)
}
