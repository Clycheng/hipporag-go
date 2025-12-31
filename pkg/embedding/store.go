package embedding

// store.go - 向量存储
// 用途：存储文本及其向量表示，支持相似度搜索和持久化
// 主要功能：
// - Insert: 插入文本并自动生成向量（自动去重）
// - Search: 基于向量相似度搜索最相关的文本
// - Get/GetContent: 根据 ID 获取向量或原始内容
// - Save/Load: 持久化存储到文件

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/example/go-scaffold/pkg/utils"
)

// Store 向量存储（内存）
type Store struct {
	client     Client
	embeddings map[string][]float64 // ID -> 向量
	contents   map[string]string    // ID -> 内容
	hashToID   map[string]string    // 内容哈希 -> ID
	mu         sync.RWMutex
}

// NewStore 创建向量存储
func NewStore(client Client) *Store {
	return &Store{
		client:     client,
		embeddings: make(map[string][]float64),
		contents:   make(map[string]string),
		hashToID:   make(map[string]string),
	}
}

// Insert 插入文本并生成向量
func (s *Store) Insert(ctx context.Context, texts []string) ([]string, error) {
	if len(texts) == 0 {
		return []string{}, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 去重：检查哪些文本已存在
	var newTexts []string
	var newIndices []int
	textToID := make(map[int]string)

	for i, text := range texts {
		hash := utils.Hash(text)
		if existingID, exists := s.hashToID[hash]; exists {
			textToID[i] = existingID
		} else {
			newTexts = append(newTexts, text)
			newIndices = append(newIndices, i)
		}
	}

	// 为新文本生成向量
	var embeddings [][]float64
	var err error
	if len(newTexts) > 0 {
		embeddings, err = s.client.Embed(ctx, newTexts)
		if err != nil {
			return nil, fmt.Errorf("embed texts: %w", err)
		}
	}

	// 存储新文本和向量
	ids := make([]string, len(texts))
	for i, idx := range newIndices {
		text := texts[idx]
		hash := utils.Hash(text)
		id := hash[:16] // 使用哈希前16位作为ID

		s.embeddings[id] = embeddings[i]
		s.contents[id] = text
		s.hashToID[hash] = id
		textToID[idx] = id
	}

	// 构建返回的ID列表
	for i := range texts {
		ids[i] = textToID[i]
	}

	return ids, nil
}

// Get 获取向量
func (s *Store) Get(ctx context.Context, id string) ([]float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	embedding, exists := s.embeddings[id]
	if !exists {
		return nil, fmt.Errorf("embedding not found: %s", id)
	}
	return embedding, nil
}

// GetContent 获取内容
func (s *Store) GetContent(ctx context.Context, id string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	content, exists := s.contents[id]
	if !exists {
		return "", fmt.Errorf("content not found: %s", id)
	}
	return content, nil
}

type searchResult struct {
	id    string
	score float64
}

// Search 搜索最相似的向量
func (s *Store) Search(ctx context.Context, query []float64, topK int) ([]string, []float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.embeddings) == 0 {
		return []string{}, []float64{}, nil
	}

	// 计算所有向量的相似度
	var results []searchResult
	for id, embedding := range s.embeddings {
		score := utils.CosineSimilarity(query, embedding)
		results = append(results, searchResult{id: id, score: score})
	}

	// 按相似度排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	// 取 topK
	if topK > len(results) {
		topK = len(results)
	}

	ids := make([]string, topK)
	scores := make([]float64, topK)
	for i := 0; i < topK; i++ {
		ids[i] = results[i].id
		scores[i] = results[i].score
	}

	return ids, scores, nil
}

type storeData struct {
	Embeddings map[string][]float64 `json:"embeddings"`
	Contents   map[string]string    `json:"contents"`
	HashToID   map[string]string    `json:"hash_to_id"`
}

// Save 保存到文件
func (s *Store) Save(path string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data := storeData{
		Embeddings: s.embeddings,
		Contents:   s.contents,
		HashToID:   s.hashToID,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal data: %w", err)
	}

	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// Load 从文件加载
func (s *Store) Load(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	jsonData, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	var data storeData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("unmarshal data: %w", err)
	}

	s.embeddings = data.Embeddings
	s.contents = data.Contents
	s.hashToID = data.HashToID

	return nil
}

// Size 返回存储的向量数量
func (s *Store) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.embeddings)
}
