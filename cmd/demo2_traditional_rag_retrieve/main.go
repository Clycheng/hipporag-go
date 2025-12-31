package main

// demo2_traditional_rag_retrieve - 命令2：传统 RAG 仅检索（不生成答案）
// 展示：向量相似度检索过程

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/example/go-scaffold/data"
	"github.com/example/go-scaffold/pkg/embedding"
	"github.com/example/go-scaffold/pkg/llm"
	"github.com/example/go-scaffold/pkg/rag"
)

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║     命令2: 传统 RAG 仅检索（不生成答案）               ║")
	fmt.Println("║     方法: 向量相似度检索                               ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")

	// 检查 API Key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("错误: 请设置 OPENAI_API_KEY 环境变量")
	}

	// 创建客户端
	embeddingClient := embedding.NewOpenAIClient(apiKey, "text-embedding-3-small")
	llmClient := llm.NewOpenAIClient(apiKey, "gpt-4o-mini")

	// 创建传统 RAG
	traditionalRAG := rag.NewTraditionalRAG(embeddingClient, llmClient, 3)

	// 索引文档
	ctx := context.Background()
	fmt.Println("\n📚 测试文档:")
	for i, doc := range data.TestDocuments {
		fmt.Printf("  文档%d: %s\n", i+1, doc)
	}

	if err := traditionalRAG.Index(ctx, data.TestDocuments); err != nil {
		log.Fatalf("索引失败: %v", err)
	}

	// 交互式检索
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🔍 进入检索模式（输入 'quit' 或 'exit' 退出）")
	fmt.Println("   注意：此模式仅展示检索过程，不生成答案")
	fmt.Println(strings.Repeat("=", 60))

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\n❓ 请输入问题: ")
		if !scanner.Scan() {
			break
		}

		question := strings.TrimSpace(scanner.Text())
		if question == "" {
			continue
		}

		if question == "quit" || question == "exit" {
			fmt.Println("\n👋 再见！")
			break
		}

		// 执行检索（不生成答案）
		_, _, err := traditionalRAG.Retrieve(ctx, question)
		if err != nil {
			fmt.Printf("❌ 检索失败: %v\n", err)
		}
	}
}
