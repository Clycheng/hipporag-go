# 快速开始指南

## 不使用 Docker 的快速测试

如果你暂时不想安装 Docker/Weaviate，可以使用内存存储快速测试：

### 1. 设置环境变量

```bash
export OPENAI_API_KEY="your-openai-api-key-here"
```

### 2. 运行内存存储示例

```bash
go run examples/index_example.go
```

这个示例会：
- 使用内存存储（无需 Docker）
- 索引 3 个关于物理学家的文档
- 显示索引统计信息

### 3. 预期输出

```
Starting indexing...
==================
Step 1: Chunking documents...
  Created 6 chunks from 3 documents
Step 2: Embedding chunks...
  Embedded 6 chunks
Step 3: Extracting entities and relations...
  Extracted 15 unique entities and 8 facts
Step 4: Embedding entities and facts...
  Embedded 15 entities and 8 facts
Step 5: Building knowledge graph...
  Graph: 29 nodes, 42 edges
Indexing completed successfully!

Indexing Statistics:
===================
  chunks: 6
  entities: 15
  facts: 8
  nodes: 29
  edges: 42

✓ Indexing completed successfully!
  System is ready for retrieval and query.
```

## 使用 Weaviate（生产环境推荐）

### 1. 安装 Docker Desktop

**macOS:**
```bash
# 使用 Homebrew
brew install --cask docker

# 或者从官网下载
# https://www.docker.com/products/docker-desktop/
```

启动 Docker Desktop 应用。

### 2. 启动 Weaviate

```bash
# 新版 Docker（推荐）
docker compose up -d

# 或者旧版
docker-compose up -d
```

### 3. 验证 Weaviate 运行

```bash
curl http://localhost:8080/v1/meta
```

应该返回 Weaviate 的元数据信息。

### 4. 运行 Weaviate 示例

```bash
export OPENAI_API_KEY="your-openai-api-key-here"
go run examples/weaviate_example.go
```

### 5. 停止 Weaviate

```bash
docker compose down
```

## 常见问题

### Q: 我没有 OpenAI API Key 怎么办？

A: 你需要：
1. 访问 https://platform.openai.com/
2. 注册账号
3. 在 API Keys 页面创建新的 API Key
4. 设置环境变量：`export OPENAI_API_KEY="sk-..."`

### Q: 内存存储和 Weaviate 有什么区别？

A: 
- **内存存储**：数据在内存中，重启丢失，适合开发测试
- **Weaviate**：数据持久化，性能更好，适合生产环境

### Q: 如何查看 Weaviate 中的数据？

A: 访问 http://localhost:8080/v1/schema 查看 Schema

或使用 GraphQL 查询：
```bash
curl -X POST http://localhost:8080/v1/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "{
      Aggregate {
        DocumentChunk {
          meta {
            count
          }
        }
      }
    }"
  }'
```

### Q: 如何清空 Weaviate 数据？

A: 
```bash
docker compose down -v  # 删除数据卷
docker compose up -d    # 重新启动
```

## 下一步

索引完成后，你可以：
1. 实现检索功能（Retrieve）
2. 实现问答功能（Query）
3. 添加更多文档
4. 调整配置参数优化性能
