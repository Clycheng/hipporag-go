package openie

// extractor.go - 开放信息抽取（OpenIE）
// 用途：从文本中提取实体和关系三元组 (主语, 谓语, 宾语)
// 主要功能：
// - Extract: 使用 LLM 从文本中提取结构化的实体关系
// - 支持批量处理

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// LLMClient LLM 客户端接口（用于调用 OpenAI 等）
type LLMClient interface {
	// Complete 生成文本补全
	Complete(ctx context.Context, prompt string) (string, error)
}

// Extractor OpenIE 提取器
type Extractor struct {
	llmClient LLMClient
}

// NewExtractor 创建 OpenIE 提取器
func NewExtractor(llmClient LLMClient) *Extractor {
	return &Extractor{
		llmClient: llmClient,
	}
}

// Triple 三元组：(主语, 谓语, 宾语)
type Triple struct {
	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`
}

// ExtractionResult 提取结果
type ExtractionResult struct {
	Entities []string `json:"entities"` // 所有实体
	Triples  []Triple `json:"triples"`  // 关系三元组
}

// Extract 从文本中提取实体和关系
func (e *Extractor) Extract(ctx context.Context, text string) (*ExtractionResult, error) {
	// 构造提示词
	prompt := buildExtractionPrompt(text)

	// 调用 LLM
	response, err := e.llmClient.Complete(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("llm complete: %w", err)
	}

	// 解析 JSON 响应
	result, err := parseExtractionResponse(response)
	if err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return result, nil
}

// buildExtractionPrompt 构造 OpenIE 提示词
func buildExtractionPrompt(text string) string {
	return fmt.Sprintf(`Extract entities and relationships from the following text.
Return the result in JSON format with two fields:
1. "entities": a list of all entities (nouns, proper nouns)
2. "triples": a list of relationship triples, each with "subject", "predicate", "object"

Text: %s

Return only valid JSON, no additional text.`, text)
}

// parseExtractionResponse 解析 LLM 返回的 JSON
func parseExtractionResponse(response string) (*ExtractionResult, error) {
	// 清理响应（移除可能的 markdown 代码块标记）
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	var result ExtractionResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("unmarshal json: %w", err)
	}

	return &result, nil
}

// ExtractBatch 批量提取（逐个处理）
func (e *Extractor) ExtractBatch(ctx context.Context, texts []string) ([]*ExtractionResult, error) {
	results := make([]*ExtractionResult, len(texts))

	for i, text := range texts {
		result, err := e.Extract(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("extract text %d: %w", i, err)
		}
		results[i] = result
	}

	return results, nil
}
