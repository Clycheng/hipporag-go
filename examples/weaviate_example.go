package main

// weaviate_example.go - 使用 Weaviate 的 HippoRAG 索引示例
// 演示如何使用 Weaviate 向量数据库替代内存存储

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/example/go-scaffold/pkg/embedding"
	"github.com/example/go-scaffold/pkg/hipporag"
	"github.com/example/go-scaffold/pkg/llm"
)

func main() {
	// 从环境变量获取 API Key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// 创建客户端
	embeddingClient := embedding.NewOpenAIClient(apiKey, "text-embedding-3-small")
	llmClient := llm.NewOpenAIClient(apiKey, "gpt-4o-mini")

	// 创建 Weaviate 存储
	weaviateConfig := embedding.WeaviateConfig{
		Host:      "localhost:8080", // 确保 Weaviate 服务正在运行
		Scheme:    "http",
		ClassName: "DocumentChunk",
	}

	chunkStore, err := embedding.NewWeaviateStore(weaviateConfig, embeddingClient)
	if err != nil {
		log.Fatalf("Failed to create chunk store: %v", err)
	}

	// 为实体和事实创建不同的集合
	entityConfig := weaviateConfig
	entityConfig.ClassName = "Entity"
	entityStore, err := embedding.NewWeaviateStore(entityConfig, embeddingClient)
	if err != nil {
		log.Fatalf("Failed to create entity store: %v", err)
	}

	factConfig := weaviateConfig
	factConfig.ClassName = "Fact"
	factStore, err := embedding.NewWeaviateStore(factConfig, embeddingClient)
	if err != nil {
		log.Fatalf("Failed to create fact store: %v", err)
	}

	// 创建 HippoRAG 实例（使用 Weaviate 存储）
	config := hipporag.DefaultConfig()
	config.ChunkSize = 256
	config.ChunkOverlap = 30

	rag := hipporag.NewHippoRAGWithStores(
		config,
		embeddingClient,
		llmClient,
		chunkStore,
		entityStore,
		factStore,
	)

	// 准备测试文档
	docs := []string{
		`Albert Einstein was a German-born theoretical physicist. He developed the theory of relativity, 
		one of the two pillars of modern physics. Einstein's work is also known for its influence on 
		the philosophy of science.`,

		`The theory of relativity usually encompasses two interrelated theories by Albert Einstein: 
		special relativity and general relativity. Special relativity applies to all physical phenomena 
		in the absence of gravity. General relativity explains the law of gravitation and its relation 
		to other forces of nature.`,

		`Isaac Newton was an English mathematician, physicist, astronomer, and author. He is widely 
		recognised as one of the greatest mathematicians and physicists of all time. Newton made 
		seminal contributions to optics, and shares credit with Leibniz for developing infinitesimal calculus.`,
	}

	// 索引文档
	fmt.Println("Starting indexing with Weaviate...")
	fmt.Println("====================================")

	ctx := context.Background()
	if err := rag.Index(ctx, docs); err != nil {
		log.Fatalf("Index error: %v", err)
	}

	// 显示统计信息
	fmt.Println("\nIndexing Statistics:")
	fmt.Println("===================")
	stats := rag.Stats(ctx)
	for key, value := range stats {
		fmt.Printf("  %s: %d\n", key, value)
	}

	fmt.Println("\n✓ Indexing completed successfully with Weaviate!")
	fmt.Println("  Data is persisted in Weaviate database.")
	fmt.Println("  You can query it even after restarting the application.")
}
