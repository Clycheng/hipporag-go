package rag

// traditional.go - 传统 RAG 实现
// 用途：简单的向量相似度检索 + LLM 生成
// 特点：只基于向量相似度，不考虑实体关系

import (
	"context"
	"fmt"
	"strings"

	"github.com/example/go-scaffold/pkg/embedding"
	"github.com/example/go-scaffold/pkg/llm"
)

// TraditionalRAG 传统 RAG 系统
type TraditionalRAG struct {
	embeddingClient embedding.Client
	llmClient       *llm.OpenAIClient
	store           embedding.VectorStore
	topK            int
}

// NewTraditionalRAG 创建传统 RAG 实例
func NewTraditionalRAG(
	embeddingClient embedding.Client,
	llmClient *llm.OpenAIClient,
	topK int,
) *TraditionalRAG {
	return &TraditionalRAG{
		embeddingClient: embeddingClient,
		llmClient:       llmClient,
		store:           embedding.NewStore(embeddingClient),
		topK:            topK,
	}
}

// Index 索引文档
func (r *TraditionalRAG) Index(ctx context.Context, docs []string) error {
	fmt.Println("\n=== 传统 RAG 索引 ===")
	fmt.Printf("索引 %d 个文档...\n", len(docs))

	ids, err := r.store.Insert(ctx, docs)
	if err != nil {
		return fmt.Errorf("insert documents: %w", err)
	}

	fmt.Printf("✓ 成功索引 %d 个文档\n", len(ids))
	return nil
}

// Retrieve 检索相关文档（仅检索，不生成答案）
func (r *TraditionalRAG) Retrieve(ctx context.Context, query string) ([]string, []float64, error) {
	fmt.Println("\n=== 传统 RAG 检索过程 ===")
	fmt.Printf("问题: %s\n\n", query)

	// 1. 向量化查询
	fmt.Println("步骤 1: 向量化查询...")
	queryVec, err := r.embeddingClient.EmbedSingle(ctx, query)
	if err != nil {
		return nil, nil, fmt.Errorf("embed query: %w", err)
	}
	fmt.Printf("✓ 查询向量维度: %d\n", len(queryVec))

	// 2. 向量相似度搜索
	fmt.Printf("\n步骤 2: 向量相似度搜索 (Top-%d)...\n", r.topK)
	ids, scores, err := r.store.Search(ctx, queryVec, r.topK)
	if err != nil {
		return nil, nil, fmt.Errorf("search: %w", err)
	}

	// 3. 获取文档内容
	fmt.Println("\n检索结果:")
	fmt.Println("---")
	docs := make([]string, len(ids))
	for i, id := range ids {
		content, _ := r.store.GetContent(ctx, id)
		docs[i] = content
		fmt.Printf("%d. [相似度: %.4f] %s\n", i+1, scores[i], content)
	}
	fmt.Println("---")

	return docs, scores, nil
}

// Query 检索并生成答案
func (r *TraditionalRAG) Query(ctx context.Context, query string) (string, error) {
	// 检索相关文档
	docs, _, err := r.Retrieve(ctx, query)
	if err != nil {
		return "", err
	}

	// 构造提示词
	fmt.Println("\n步骤 3: 使用 LLM 生成答案...")
	context := strings.Join(docs, "\n")
	prompt := fmt.Sprintf(`基于以下文档回答问题。如果文档中没有足够信息，请说明。

文档:
%s

问题: %s

答案:`, context, query)

	// 生成答案
	answer, err := r.llmClient.Complete(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("generate answer: %w", err)
	}

	fmt.Println("\n=== 生成的答案 ===")
	fmt.Println(answer)
	fmt.Println("==================")

	return answer, nil
}
