package main

// demo1_traditional_rag_qa - 命令1：传统 RAG 检索 + LLM 生成答案
// 展示：简单的向量相似度检索 + LLM 生成

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
	fmt.Println("║     命令1: 传统 RAG 检索 + LLM 生成答案                ║")
	fmt.Println("║     方法: 向量相似度检索 → LLM 生成                    ║")
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

	// 交互式问答
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💬 进入交互模式（输入 'quit' 或 'exit' 退出）")
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

		// 执行查询
		_, err := traditionalRAG.Query(ctx, question)
		if err != nil {
			fmt.Printf("❌ 查询失败: %v\n", err)
		}
	}
}
