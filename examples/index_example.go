package main

// index_example.go - HippoRAG 索引示例
// 演示如何使用 HippoRAG 索引文档

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

	// 创建 HippoRAG 实例
	config := hipporag.DefaultConfig()
	config.ChunkSize = 256 // 使用较小的块大小用于演示
	config.ChunkOverlap = 30

	rag := hipporag.NewHippoRAG(config, embeddingClient, llmClient)

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
	fmt.Println("Starting indexing...")
	fmt.Println("==================")

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

	fmt.Println("\n✓ Indexing completed successfully!")
	fmt.Println("  System is ready for retrieval and query.")
}
