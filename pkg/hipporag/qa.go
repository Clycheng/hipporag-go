package hipporag

// qa.go - HippoRAG 问答实现
// 用途：基于检索结果生成答案
// 流程：检索 → LLM 生成

import (
	"context"
	"fmt"
	"strings"
)

// Query 问答：检索 + 生成答案
// query: 用户问题
// 返回：生成的答案
func (h *HippoRAG) Query(ctx context.Context, query string) (string, error) {
	// 检索相关文档
	solutions, err := h.Retrieve(ctx, []string{query}, h.config.TopKChunks)
	if err != nil {
		return "", err
	}

	if len(solutions) == 0 {
		return "", fmt.Errorf("no solutions found")
	}

	solution := solutions[0]

	// 构造提示词
	fmt.Println("\n步骤 5: 使用 LLM 生成答案...")
	context := strings.Join(solution.ChunkTexts, "\n")
	prompt := fmt.Sprintf(`基于以下文档回答问题。请仔细阅读所有文档，找出相关信息并进行推理。

文档:
%s

问题: %s

请一步步思考，然后给出简洁的答案。

答案:`, context, query)

	// 生成答案
	answer, err := h.llmClient.Complete(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("generate answer: %w", err)
	}

	fmt.Println("\n=== 生成的答案 ===")
	fmt.Println(answer)
	fmt.Println("==================")

	return answer, nil
}

// QueryFull 完整版问答：使用完整检索流程 + 生成答案
// query: 用户问题
// 返回：生成的答案
func (h *HippoRAG) QueryFull(ctx context.Context, query string) (string, error) {
	// 使用完整检索
	solutions, err := h.RetrieveFull(ctx, []string{query}, h.config.TopKChunks)
	if err != nil {
		return "", err
	}

	if len(solutions) == 0 {
		return "", fmt.Errorf("no solutions found")
	}

	solution := solutions[0]

	// 构造提示词
	fmt.Println("\n步骤 8: 使用 LLM 生成答案...")
	context := strings.Join(solution.ChunkTexts, "\n")
	prompt := fmt.Sprintf(`基于以下文档回答问题。请仔细阅读所有文档，找出相关信息并进行推理。

文档:
%s

问题: %s

请一步步思考，然后给出简洁的答案。

答案:`, context, query)

	// 生成答案
	answer, err := h.llmClient.Complete(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("generate answer: %w", err)
	}

	fmt.Println("\n=== 生成的答案 ===")
	fmt.Println(answer)
	fmt.Println("==================")

	return answer, nil
}
