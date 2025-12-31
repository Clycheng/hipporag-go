package hipporag

// retrieve_full.go - HippoRAG 完整检索实现
// 用途：实现 DEMO.md 中描述的完整检索流程
// 包含：事实检索、LLM 重排序、DPR、权重合并

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/example/go-scaffold/pkg/utils"
)

// RetrieveFull 完整版检索（包含所有步骤）
func (h *HippoRAG) RetrieveFull(ctx context.Context, queries []string, topK int) ([]QuerySolution, error) {
	if !h.readyToRetrieve {
		return nil, fmt.Errorf("index not ready, please call Index first")
	}

	solutions := make([]QuerySolution, len(queries))

	for i, query := range queries {
		fmt.Printf("\n=== HippoRAG 完整检索过程 (查询 %d/%d) ===\n", i+1, len(queries))
		fmt.Printf("问题: %s\n\n", query)

		// ========== 步骤 1: 准备检索对象 ==========
		// （已在 Index 阶段完成）

		// ========== 步骤 2: 查询向量化 ==========
		fmt.Println("步骤 2: 查询向量化...")

		// 2.1 生成 query_to_fact 向量（用于事实检索）
		queryVecForFact, err := h.embeddingClient.EmbedSingle(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("embed query for fact: %w", err)
		}

		// 2.2 生成 query_to_passage 向量（用于段落检索）
		queryVecForPassage, err := h.embeddingClient.EmbedSingle(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("embed query for passage: %w", err)
		}

		fmt.Printf("✓ 查询向量维度: %d\n", len(queryVecForFact))

		// ========== 步骤 3: 事实检索 ==========
		fmt.Printf("\n步骤 3: 事实检索 (Top-%d)...\n", h.config.TopKEntities)

		factIDs, factScores, err := h.factStore.Search(ctx, queryVecForFact, h.config.TopKEntities)
		if err != nil {
			return nil, fmt.Errorf("search facts: %w", err)
		}

		fmt.Println("找到的相关事实:")
		for j, id := range factIDs {
			content, _ := h.factStore.GetContent(ctx, id)
			fmt.Printf("  %d. [分数: %.4f] %s\n", j+1, factScores[j], content)
		}

		// ========== 步骤 4: 事实重排序（Recognition Memory）==========
		fmt.Println("\n步骤 4: LLM 重排序事实...")

		// 构建重排序 prompt
		var factsText strings.Builder
		for j, id := range factIDs {
			content, _ := h.factStore.GetContent(ctx, id)
			factsText.WriteString(fmt.Sprintf("%d. %s\n", j+1, content))
		}

		rerankerPrompt := fmt.Sprintf(`给定查询："%s"

请对以下事实按相关性排序（最相关的排在前面）：
%s

只返回排序后的序号，用逗号分隔。例如：3,1,4,2,5

排序结果：`, query, factsText.String())

		rerankedIndices := factIDs // 默认不重排序

		// 调用 LLM 重排序（可选，如果 LLM 调用失败则跳过）
		if response, err := h.llmClient.Complete(ctx, rerankerPrompt); err == nil {
			// 解析 LLM 返回的排序
			parts := strings.Split(strings.TrimSpace(response), ",")
			if len(parts) > 0 {
				newIndices := make([]string, 0, len(parts))
				for _, part := range parts {
					var idx int
					if _, err := fmt.Sscanf(strings.TrimSpace(part), "%d", &idx); err == nil {
						if idx > 0 && idx <= len(factIDs) {
							newIndices = append(newIndices, factIDs[idx-1])
						}
					}
				}
				if len(newIndices) > 0 {
					rerankedIndices = newIndices
					fmt.Println("✓ LLM 重排序完成")
				}
			}
		} else {
			fmt.Println("⚠️  LLM 重排序失败，使用原始排序")
		}

		// ========== 步骤 5: 密集段落检索（DPR）==========
		fmt.Printf("\n步骤 5: 密集段落检索 (Top-%d)...\n", topK)

		chunkIDs, chunkScores, err := h.chunkStore.Search(ctx, queryVecForPassage, topK)
		if err != nil {
			return nil, fmt.Errorf("search chunks: %w", err)
		}

		fmt.Println("DPR 检索到的段落:")
		for j, id := range chunkIDs {
			content, _ := h.chunkStore.GetContent(ctx, id)
			fmt.Printf("  %d. [分数: %.4f] %s\n", j+1, chunkScores[j], content)
		}

		// ========== 步骤 6: 图搜索与 PPR ==========
		fmt.Println("\n步骤 6: 图搜索与 PPR...")

		// 6.1 从重排序后的事实中提取实体
		entityWeights := make(map[string]float64)

		for rank, factID := range rerankedIndices {
			if rank >= len(factScores) {
				break
			}

			// 获取事实内容
			factContent, _ := h.factStore.GetContent(ctx, factID)

			// 解析三元组（简化版：假设格式为 "[subject, predicate, object]"）
			// 实际应该从索引时保存的映射中获取
			entities := extractEntitiesFromFact(factContent)

			// 为每个实体分配权重
			factScore := factScores[rank]
			for _, entity := range entities {
				entityID := utils.ComputeHash(entity, "entity-")
				if _, exists := h.graph.GetNode(entityID); exists {
					entityWeights[entityID] += factScore / float64(len(entities))
				}
			}
		}

		fmt.Printf("✓ 从事实中提取了 %d 个实体\n", len(entityWeights))

		// 6.2 合并段落权重（DPR 结果）
		passageNodeWeight := 0.05 // 段落权重系数

		// 归一化 DPR 分数
		normalizedChunkScores := utils.MinMaxNormalize(chunkScores)

		for j, chunkID := range chunkIDs {
			if _, exists := h.graph.GetNode(chunkID); exists {
				entityWeights[chunkID] = normalizedChunkScores[j] * passageNodeWeight
			}
		}

		fmt.Printf("✓ 合并了 %d 个段落权重\n", len(chunkIDs))

		// 6.3 运行 PPR 算法
		fmt.Printf("\n执行 PPR 算法...\n")
		fmt.Printf("  参数: damping=%.2f, maxIter=%d, tolerance=%.0e\n",
			h.config.PPRDamping, h.config.PPRMaxIter, h.config.PPRTolerance)

		pprScores := h.graph.PPR(
			entityWeights,
			h.config.PPRDamping,
			h.config.PPRMaxIter,
			h.config.PPRTolerance,
		)

		fmt.Printf("✓ PPR 完成，计算了 %d 个节点的分数\n", len(pprScores))

		// ========== 步骤 7: 返回 Top-K 文档 ==========
		fmt.Println("\n步骤 7: 筛选和排序文档块...")

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

		resultChunkIDs := make([]string, topK)
		resultChunkTexts := make([]string, topK)
		resultScores := make([]float64, topK)

		fmt.Println("\n最终检索结果:")
		fmt.Println("---")
		for j := 0; j < topK; j++ {
			resultChunkIDs[j] = chunks[j].id
			resultScores[j] = chunks[j].score
			content, _ := h.chunkStore.GetContent(ctx, chunks[j].id)
			resultChunkTexts[j] = content
			fmt.Printf("%d. [PPR分数: %.6f] %s\n", j+1, resultScores[j], content)
		}
		fmt.Println("---")

		solutions[i] = QuerySolution{
			Query:      query,
			ChunkIDs:   resultChunkIDs,
			ChunkTexts: resultChunkTexts,
			Scores:     resultScores,
		}
	}

	return solutions, nil
}

// extractEntitiesFromFact 从事实字符串中提取实体
// 简化版实现，实际应该从索引时保存的映射中获取
func extractEntitiesFromFact(factContent string) []string {
	// 移除方括号和引号
	factContent = strings.Trim(factContent, "[]")
	factContent = strings.ReplaceAll(factContent, "'", "")
	factContent = strings.ReplaceAll(factContent, "\"", "")

	// 按逗号分割
	parts := strings.Split(factContent, ",")

	entities := make([]string, 0)
	for i, part := range parts {
		part = strings.TrimSpace(part)
		// 跳过谓词（中间的部分）
		if i == 1 {
			continue
		}
		if part != "" {
			entities = append(entities, part)
		}
	}

	return entities
}
