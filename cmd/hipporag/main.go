package main

// demo3_hipporag_qa - 命令3：HippoRAG 检索 + LLM 生成答案
// 展示：知识图谱 + PPR 检索 + LLM 生成

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/example/go-scaffold/data"
	"github.com/example/go-scaffold/pkg/embedding"
	"github.com/example/go-scaffold/pkg/hipporag"
	"github.com/example/go-scaffold/pkg/llm"
)

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║     命令3: HippoRAG 检索 + LLM 生成答案                ║")
	fmt.Println("║     方法: 实体检索 → PPR 图传播 → LLM 生成            ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")

	// 检查 API Key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("错误: 请设置 OPENAI_API_KEY 环境变量")
	}

	// 创建客户端
	embeddingClient := embedding.NewOpenAIClient(apiKey, "text-embedding-3-small")
	llmClient := llm.NewOpenAIClient(apiKey, "gpt-4o-mini")

	// 创建 HippoRAG
	config := hipporag.DefaultConfig()
	config.ChunkSize = 100 // 使用较小的块，因为文档很短
	config.ChunkOverlap = 0
	config.TopKEntities = 20 // 增加到 20，应对更多噪音
	config.TopKChunks = 15   // 增加到 15，确保能检索到所有相关文档
	config.PPRDamping = 0.5  // 降低阻尼系数，让分数传播更广

	rag := hipporag.NewHippoRAG(config, embeddingClient, llmClient)

	// 索引文档
	ctx := context.Background()
	fmt.Println("\n📚 测试文档:")
	for i, doc := range data.TestDocuments {
		fmt.Printf("  文档%d: %s\n", i+1, doc)
	}

	if err := rag.Index(ctx, data.TestDocuments); err != nil {
		log.Fatalf("索引失败: %v", err)
	}

	// 显示统计信息
	stats := rag.Stats(ctx)
	fmt.Println("\n📊 索引统计:")
	fmt.Printf("  文档块: %d\n", stats["chunks"])
	fmt.Printf("  实体: %d\n", stats["entities"])
	fmt.Printf("  事实: %d\n", stats["facts"])
	fmt.Printf("  图节点: %d\n", stats["nodes"])
	fmt.Printf("  图边: %d\n", stats["edges"])

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
		_, err := rag.QueryFull(ctx, question)
		if err != nil {
			fmt.Printf("❌ 查询失败: %v\n", err)
		}
	}
}
