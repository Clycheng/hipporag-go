package embedding

// weaviate.go - Weaviate 向量存储实现
// 用途：使用 Weaviate 向量数据库存储和检索向量
// 主要功能：
// - 自动创建 Schema（集合）
// - 插入文本和向量
// - 向量相似度搜索
// - 支持批量操作

import (
	"context"
	"fmt"

	"github.com/example/go-scaffold/pkg/utils"
	"github.com/go-openapi/strfmt"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
)

// WeaviateStore Weaviate 向量存储
type WeaviateStore struct {
	client    *weaviate.Client
	className string // Weaviate 中的类名（集合名）
	embClient Client // Embedding 客户端
}

// WeaviateConfig Weaviate 配置
type WeaviateConfig struct {
	Host      string // Weaviate 服务地址，例如 "localhost:8080"
	Scheme    string // "http" 或 "https"
	ClassName string // 集合名称，例如 "Document"
}

// NewWeaviateStore 创建 Weaviate 存储
func NewWeaviateStore(config WeaviateConfig, embClient Client) (*WeaviateStore, error) {
	// 创建 Weaviate 客户端
	cfg := weaviate.Config{
		Host:   config.Host,
		Scheme: config.Scheme,
	}

	client, err := weaviate.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("create weaviate client: %w", err)
	}

	store := &WeaviateStore{
		client:    client,
		className: config.ClassName,
		embClient: embClient,
	}

	// 创建 Schema（如果不存在）
	if err := store.createSchema(context.Background()); err != nil {
		return nil, fmt.Errorf("create schema: %w", err)
	}

	return store, nil
}

// createSchema 创建 Weaviate Schema
func (s *WeaviateStore) createSchema(ctx context.Context) error {
	// 检查类是否已存在
	exists, err := s.client.Schema().ClassExistenceChecker().
		WithClassName(s.className).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("check class existence: %w", err)
	}

	if exists {
		return nil // 已存在，无需创建
	}

	// 创建类定义
	classObj := &models.Class{
		Class:       s.className,
		Description: "Document chunks with embeddings",
		Properties: []*models.Property{
			{
				Name:        "content",
				DataType:    []string{"text"},
				Description: "The text content",
			},
			{
				Name:        "hash",
				DataType:    []string{"text"},
				Description: "Content hash for deduplication",
			},
		},
		Vectorizer: "none", // 我们自己提供向量
	}

	err = s.client.Schema().ClassCreator().
		WithClass(classObj).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("create class: %w", err)
	}

	return nil
}

// Insert 插入文本并生成向量
func (s *WeaviateStore) Insert(ctx context.Context, texts []string) ([]string, error) {
	if len(texts) == 0 {
		return []string{}, nil
	}

	// 生成向量
	embeddings, err := s.embClient.Embed(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("embed texts: %w", err)
	}

	// 批量插入
	batcher := s.client.Batch().ObjectsBatcher()
	ids := make([]string, len(texts))

	for i, text := range texts {
		hash := utils.Hash(text)
		id := hash[:16] // 使用哈希前16位作为 ID
		ids[i] = id

		// 转换 float64 到 float32
		vector32 := make([]float32, len(embeddings[i]))
		for j, v := range embeddings[i] {
			vector32[j] = float32(v)
		}

		// 创建对象
		obj := &models.Object{
			Class: s.className,
			ID:    strfmt.UUID(id),
			Properties: map[string]interface{}{
				"content": text,
				"hash":    hash,
			},
			Vector: vector32,
		}

		batcher = batcher.WithObject(obj)
	}

	// 执行批量插入
	_, err = batcher.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("batch insert: %w", err)
	}

	return ids, nil
}

// Search 向量相似度搜索
func (s *WeaviateStore) Search(ctx context.Context, queryVec []float64, topK int) ([]string, []float64, error) {
	// 转换 float64 到 float32
	queryVec32 := make([]float32, len(queryVec))
	for i, v := range queryVec {
		queryVec32[i] = float32(v)
	}

	// 构建 GraphQL 查询
	nearVector := s.client.GraphQL().NearVectorArgBuilder().
		WithVector(queryVec32)

	result, err := s.client.GraphQL().Get().
		WithClassName(s.className).
		WithFields(graphql.Field{Name: "content"}, graphql.Field{Name: "_additional { id distance }"}).
		WithNearVector(nearVector).
		WithLimit(topK).
		Do(ctx)

	if err != nil {
		return nil, nil, fmt.Errorf("search: %w", err)
	}

	// 解析结果
	data, ok := result.Data["Get"].(map[string]interface{})
	if !ok {
		return []string{}, []float64{}, nil
	}

	items, ok := data[s.className].([]interface{})
	if !ok {
		return []string{}, []float64{}, nil
	}

	ids := make([]string, 0, len(items))
	scores := make([]float64, 0, len(items))

	for _, item := range items {
		obj := item.(map[string]interface{})
		additional := obj["_additional"].(map[string]interface{})

		id := additional["id"].(string)
		distance := additional["distance"].(float64)

		// 将距离转换为相似度分数（距离越小，相似度越高）
		similarity := 1.0 - distance

		ids = append(ids, id)
		scores = append(scores, similarity)
	}

	return ids, scores, nil
}

// Get 根据 ID 获取向量
func (s *WeaviateStore) Get(ctx context.Context, id string) ([]float64, error) {
	result, err := s.client.Data().ObjectsGetter().
		WithClassName(s.className).
		WithID(id).
		WithVector().
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("object not found: %s", id)
	}

	// 转换 float32 到 float64
	vector32 := result[0].Vector
	vector64 := make([]float64, len(vector32))
	for i, v := range vector32 {
		vector64[i] = float64(v)
	}

	return vector64, nil
}

// GetContent 根据 ID 获取内容
func (s *WeaviateStore) GetContent(ctx context.Context, id string) (string, error) {
	result, err := s.client.Data().ObjectsGetter().
		WithClassName(s.className).
		WithID(id).
		Do(ctx)

	if err != nil {
		return "", fmt.Errorf("get object: %w", err)
	}

	if len(result) == 0 {
		return "", fmt.Errorf("object not found: %s", id)
	}

	content, ok := result[0].Properties.(map[string]interface{})["content"].(string)
	if !ok {
		return "", fmt.Errorf("content field not found")
	}

	return content, nil
}

// Delete 删除对象
func (s *WeaviateStore) Delete(ctx context.Context, id string) error {
	err := s.client.Data().Deleter().
		WithClassName(s.className).
		WithID(id).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("delete object: %w", err)
	}

	return nil
}

// DeleteAll 删除所有对象（清空集合）
func (s *WeaviateStore) DeleteAll(ctx context.Context) error {
	err := s.client.Schema().ClassDeleter().
		WithClassName(s.className).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("delete class: %w", err)
	}

	// 重新创建 Schema
	return s.createSchema(ctx)
}

// Size 返回存储的对象数量
func (s *WeaviateStore) Size(ctx context.Context) (int, error) {
	result, err := s.client.GraphQL().Aggregate().
		WithClassName(s.className).
		WithFields(graphql.Field{Name: "meta { count }"}).
		Do(ctx)

	if err != nil {
		return 0, fmt.Errorf("aggregate query: %w", err)
	}

	data, ok := result.Data["Aggregate"].(map[string]interface{})
	if !ok {
		return 0, nil
	}

	items, ok := data[s.className].([]interface{})
	if !ok || len(items) == 0 {
		return 0, nil
	}

	meta := items[0].(map[string]interface{})["meta"].(map[string]interface{})
	count := int(meta["count"].(float64))

	return count, nil
}
