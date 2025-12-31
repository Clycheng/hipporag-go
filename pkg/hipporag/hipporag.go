package hipporag

// hipporag.go - HippoRAG 主类
// 用途：HippoRAG 系统的核心，整合知识图谱、向量检索和 LLM
// 主要功能：
// - Index: 索引文档（分块 → OpenIE → 构建图谱 → 向量化）
// - Retrieve: 检索相关文档（向量检索 + PPR 图检索）
// - Query: 问答（检索 + LLM 生成）

import (
	"github.com/example/go-scaffold/pkg/embedding"
	"github.com/example/go-scaffold/pkg/graph"
	"github.com/example/go-scaffold/pkg/llm"
	"github.com/example/go-scaffold/pkg/openie"
)

// Config HippoRAG 配置
type Config struct {
	// 文本分块参数
	ChunkSize    int // 每块字符数，默认 512
	ChunkOverlap int // 块重叠字符数，默认 50

	// PPR 参数
	PPRDamping   float64 // 阻尼系数，默认 0.5
	PPRMaxIter   int     // 最大迭代次数，默认 100
	PPRTolerance float64 // 收敛阈值，默认 1e-6

	// 检索参数
	TopKEntities int // 检索的实体数量，默认 10
	TopKChunks   int // 最终返回的文档块数量，默认 5
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		ChunkSize:    512,
		ChunkOverlap: 50,
		PPRDamping:   0.5,
		PPRMaxIter:   100,
		PPRTolerance: 1e-6,
		TopKEntities: 10,
		TopKChunks:   5,
	}
}

// HippoRAG 主类
type HippoRAG struct {
	// 配置
	config *Config

	// 客户端
	llmClient       *llm.OpenAIClient
	embeddingClient embedding.Client

	// 存储（使用接口，支持内存或 Weaviate）
	chunkStore  embedding.VectorStore // 文档块向量存储
	entityStore embedding.VectorStore // 实体向量存储
	factStore   embedding.VectorStore // 事实向量存储

	// 知识图谱
	graph *graph.Graph

	// OpenIE 抽取器
	openie *openie.Extractor

	// 缓存
	queryEmbeddings map[string][]float64

	// 状态
	readyToRetrieve bool
}

// NewHippoRAG 创建 HippoRAG 实例（使用内存存储）
func NewHippoRAG(
	config *Config,
	embeddingClient embedding.Client,
	llmClient *llm.OpenAIClient,
) *HippoRAG {
	if config == nil {
		config = DefaultConfig()
	}

	return &HippoRAG{
		config:          config,
		embeddingClient: embeddingClient,
		llmClient:       llmClient,
		chunkStore:      embedding.NewStore(embeddingClient),
		entityStore:     embedding.NewStore(embeddingClient),
		factStore:       embedding.NewStore(embeddingClient),
		graph:           graph.NewGraph(),
		openie:          openie.NewExtractor(llmClient),
		queryEmbeddings: make(map[string][]float64),
		readyToRetrieve: false,
	}
}

// NewHippoRAGWithStores 创建 HippoRAG 实例（使用自定义存储，如 Weaviate）
func NewHippoRAGWithStores(
	config *Config,
	embeddingClient embedding.Client,
	llmClient *llm.OpenAIClient,
	chunkStore embedding.VectorStore,
	entityStore embedding.VectorStore,
	factStore embedding.VectorStore,
) *HippoRAG {
	if config == nil {
		config = DefaultConfig()
	}

	return &HippoRAG{
		config:          config,
		embeddingClient: embeddingClient,
		llmClient:       llmClient,
		chunkStore:      chunkStore,
		entityStore:     entityStore,
		factStore:       factStore,
		graph:           graph.NewGraph(),
		openie:          openie.NewExtractor(llmClient),
		queryEmbeddings: make(map[string][]float64),
		readyToRetrieve: false,
	}
}

// QuerySolution 查询解决方案（检索结果）
type QuerySolution struct {
	Query      string    // 查询文本
	ChunkIDs   []string  // 相关文档块 ID
	ChunkTexts []string  // 相关文档块内容
	Scores     []float64 // 相关性分数
}
