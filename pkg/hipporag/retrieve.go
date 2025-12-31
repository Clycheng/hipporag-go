package hipporag

// retrieve.go - HippoRAG 检索实现
// 用途：基于知识图谱的检索
// 检索流程：
// 1. 向量检索找到相关实体
// 2. 使用 PPR 在图上传播
// 3. 返回高分数的文档块

import (
	"context"
	"fmt"
	"sort"
)

// Retrieve 检索相关文档块（不生成答案）
// queries: 查询列表
// topK: 返回的文档块数量
func (h *HippoRAG) Retrieve(ctx context.Context, queries []string, topK int) ([]QuerySolution, error) {
	if !h.readyToRetrieve {
		return nil, fmt.Errorf("index not ready, please call Index first")
	}
	
	solutions := make([]QuerySolution, len(queries))
	
	for i, query := range queries {
		fmt.Printf("\n=== HippoRAG 检索过程 (查询 %d/%d) ===\n", i+1, len(queries))
		fmt.Printf("问题: %s\n\n", query)
		
		// 步骤 1: 向量化查询
		fmt.Println("步骤 1: 向量化查询...")
		queryVec, err := h.embeddingClient.EmbedSingle(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("embed query: %w", err)
		}
		fmt.Printf("✓ 查询向量维度: %d\n", len(queryVec))
		
		// 步骤 2: 在实体存储中搜索相关实体
		fmt.Printf("\n步骤 2: 搜索相关实体 (Top-%d)...\n", h.config.TopKEntities)
		entityIDs, entityScores, err := h.entityStore.Search(ctx, queryVec, h.config.TopKEntities)
		if err != nil {
			return nil, fmt.Errorf("search entities: %w", err)
		}
		
		fmt.Println("找到的相关实体:")
		for j, id := range entityIDs {
			content, _ := h.entityStore.GetContent(ctx, id)
			fmt.Printf("  %d. [分数: %.4f] %s\n", j+1, entityScores[j], content)
		}
		
		// 步骤 3: 使用 PPR 在图上传播
		fmt.Printf("\n步骤 3: 在知识图谱上执行 PPR 算法...\n")
		fmt.Printf("  参数: damping=%.2f, maxIter=%d, tolerance=%.0e\n", 
			h.config.PPRDamping, h.config.PPRMaxIter, h.config.PPRTolerance)
		
		// 构建种子节点权重
		seedWeights := make(map[string]float64)
		for j, id := range entityIDs {
			seedWeights[id] = entityScores[j]
		}
		
		// 执行 PPR
		pprScores := h.graph.PPR(
			seedWeights,
			h.config.PPRDamping,
			h.config.PPRMaxIter,
			h.config.PPRTolerance,
		)
		
		fmt.Printf("✓ PPR 完成，计算了 %d 个节点的分数\n", len(pprScores))
		
		// 步骤 4: 筛选文档块节点并排序
		fmt.Println("\n步骤 4: 筛选和排序文档块...")
		type chunkScore struct {
			id    string
			score float64
		}
		
		var chunks []chunkScore
		for nodeID, score := range pprScores {
			node, exists := h.graph.GetNode(nodeID)
			if exists && node.Type == "chunk" {
				chunks = append(chunks, chunkScore{id: nodeID, score: score})
			}
		}
		
		// 按分数降序排序
		sort.Slice(chunks, func(i, j int) bool {
			return chunks[i].score > chunks[j].score
		})
		
		// 取 topK
		if topK > len(chunks) {
			topK = len(chunks)
		}
		
		chunkIDs := make([]string, topK)
		chunkTexts := make([]string, topK)
		scores := make([]float64, topK)
		
		fmt.Println("\n检索到的文档块:")
		fmt.Println("---")
		for j := 0; j < topK; j++ {
			chunkIDs[j] = chunks[j].id
			scores[j] = chunks[j].score
			content, _ := h.chunkStore.GetContent(ctx, chunks[j].id)
			chunkTexts[j] = content
			fmt.Printf("%d. [PPR分数: %.6f] %s\n", j+1, scores[j], content)
		}
		fmt.Println("---")
		
		solutions[i] = QuerySolution{
			Query:      query,
			ChunkIDs:   chunkIDs,
			ChunkTexts: chunkTexts,
			Scores:     scores,
		}
	}
	
	return solutions, nil
}