package hipporag

// index.go - 文档索引实现
// 用途：将文档索引到 HippoRAG 系统
// 索引流程：
// 1. 文档分块
// 2. OpenIE 提取实体和关系
// 3. 构建知识图谱（节点：实体+文档块，边：关系）
// 4. 向量化存储（文档块、实体、事实）

import (
	"context"
	"fmt"

	"github.com/example/go-scaffold/pkg/embedding"
	"github.com/example/go-scaffold/pkg/utils"
)

// Index 索引文档列表
// docs: 文档文本数组
// 返回：索引的文档块数量
func (h *HippoRAG) Index(ctx context.Context, docs []string) error {
	if len(docs) == 0 {
		return fmt.Errorf("no documents to index")
	}

	// 步骤 1: 文档分块
	fmt.Println("Step 1: Chunking documents...")
	var allChunks []string
	var chunkToDoc []int // 记录每个块属于哪个文档

	for docIdx, doc := range docs {
		chunks := utils.ChunkText(doc, h.config.ChunkSize, h.config.ChunkOverlap)
		for _, chunk := range chunks {
			allChunks = append(allChunks, chunk)
			chunkToDoc = append(chunkToDoc, docIdx)
		}
	}
	fmt.Printf("  Created %d chunks from %d documents\n", len(allChunks), len(docs))

	// 步骤 2: 向量化文档块
	fmt.Println("Step 2: Embedding chunks...")
	chunkIDs, err := h.chunkStore.Insert(ctx, allChunks)
	if err != nil {
		return fmt.Errorf("insert chunks: %w", err)
	}
	fmt.Printf("  Embedded %d chunks\n", len(chunkIDs))

	// 步骤 3: OpenIE 提取实体和关系
	fmt.Println("Step 3: Extracting entities and relations...")
	extractions, err := h.openie.ExtractBatch(ctx, allChunks)
	if err != nil {
		return fmt.Errorf("extract entities: %w", err)
	}

	// 收集所有唯一实体和事实
	entitySet := make(map[string]bool)
	var allFacts []string

	for _, extraction := range extractions {
		for _, entity := range extraction.Entities {
			entitySet[entity] = true
		}
		for _, triple := range extraction.Triples {
			// 将三元组转换为文本形式
			fact := fmt.Sprintf("%s %s %s", triple.Subject, triple.Predicate, triple.Object)
			allFacts = append(allFacts, fact)

			// 也将主语和宾语加入实体集合
			entitySet[triple.Subject] = true
			entitySet[triple.Object] = true
		}
	}

	entities := make([]string, 0, len(entitySet))
	for entity := range entitySet {
		entities = append(entities, entity)
	}
	fmt.Printf("  Extracted %d unique entities and %d facts\n", len(entities), len(allFacts))

	// 步骤 4: 向量化实体和事实
	fmt.Println("Step 4: Embedding entities and facts...")
	entityIDs, err := h.entityStore.Insert(ctx, entities)
	if err != nil {
		return fmt.Errorf("insert entities: %w", err)
	}

	var factIDs []string
	if len(allFacts) > 0 {
		factIDs, err = h.factStore.Insert(ctx, allFacts)
		if err != nil {
			return fmt.Errorf("insert facts: %w", err)
		}
	}
	fmt.Printf("  Embedded %d entities and %d facts\n", len(entityIDs), len(factIDs))

	// 步骤 5: 构建知识图谱
	fmt.Println("Step 5: Building knowledge graph...")

	// 5.1 添加文档块节点
	for _, chunkID := range chunkIDs {
		content, _ := h.chunkStore.GetContent(ctx, chunkID)
		h.graph.AddNode(chunkID, content, "chunk")
	}

	// 5.2 添加实体节点
	entityIDMap := make(map[string]string) // entity text -> ID
	for i, entity := range entities {
		entityID := entityIDs[i]
		h.graph.AddNode(entityID, entity, "entity")
		entityIDMap[entity] = entityID
	}

	// 5.3 添加边
	for chunkIdx, extraction := range extractions {
		chunkID := chunkIDs[chunkIdx]

		// 添加 passage 边：文档块 <-> 实体（双向）
		for _, entity := range extraction.Entities {
			if entityID, exists := entityIDMap[entity]; exists {
				// 正向：chunk -> entity
				h.graph.AddEdge(chunkID, entityID, 1.0, "passage")
				// 反向：entity -> chunk（让 PPR 能传播回文档块）
				h.graph.AddEdge(entityID, chunkID, 1.0, "passage_back")
			}
		}

		// 添加 fact 边：实体 <-> 实体（双向，支持双向推理）
		for _, triple := range extraction.Triples {
			subjectID, subjectExists := entityIDMap[triple.Subject]
			objectID, objectExists := entityIDMap[triple.Object]

			if subjectExists && objectExists {
				// 正向边
				h.graph.AddEdge(subjectID, objectID, 1.0, "fact")
				// 反向边（权重可以稍低）
				h.graph.AddEdge(objectID, subjectID, 0.5, "fact_back")
			}
		}
	}

	fmt.Printf("  Graph: %d nodes, %d edges\n", h.graph.NodeCount(), h.graph.EdgeCount())

	// 标记为可检索
	h.readyToRetrieve = true

	fmt.Println("Indexing completed successfully!")
	return nil
}

// IsReady 检查是否已完成索引，可以进行检索
func (h *HippoRAG) IsReady() bool {
	return h.readyToRetrieve
}

// Stats 返回索引统计信息
func (h *HippoRAG) Stats(ctx context.Context) map[string]int {
	// 注意：内存存储的 Size() 不需要 context，但为了接口统一，我们传递它
	// Weaviate 存储需要 context 来查询数据库
	return map[string]int{
		"chunks":   h.chunkStore.(*embedding.Store).Size(),
		"entities": h.entityStore.(*embedding.Store).Size(),
		"facts":    h.factStore.(*embedding.Store).Size(),
		"nodes":    h.graph.NodeCount(),
		"edges":    h.graph.EdgeCount(),
	}
}
