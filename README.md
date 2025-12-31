# HippoRAG - Go Implementation

HippoRAG (Hippocampus-Inspired Retrieval-Augmented Generation) 是一个受海马体启发的知识检索系统，结合了知识图谱和向量检索的优势。

## 项目结构

```
├── pkg/
│   ├── hipporag/       # HippoRAG 核心实现
│   │   ├── hipporag.go # 主类和配置
│   │   ├── index.go    # 文档索引
│   │   ├── retrieve.go # 检索实现（待实现）
│   │   └── qa.go       # 问答实现（待实现）
│   ├── embedding/      # 向量化相关
│   │   ├── client.go   # Embedding 接口
│   │   ├── openai.go   # OpenAI Embedding 实现
│   │   └── store.go    # 向量存储
│   ├── graph/          # 知识图谱
│   │   ├── graph.go    # 图结构
│   │   └── ppr.go      # Personalized PageRank
│   ├── openie/         # 开放信息抽取
│   │   └── extractor.go # 实体关系提取
│   ├── llm/            # LLM 客户端
│   │   └── openai.go   # OpenAI LLM 实现
│   └── utils/          # 工具函数
│       ├── hash.go     # 哈希工具
│       ├── text.go     # 文本处理
│       └── vector.go   # 向量计算
└── examples/
    └── index_example.go # 索引示例
```

## 已实现功能

### ✅ 索引部分 (Index)

索引流程将文档转换为可检索的知识图谱：

1. **文档分块** - 将长文档切分成固定大小的块（支持重叠）
2. **OpenIE 提取** - 使用 LLM 提取实体和关系三元组
3. **构建图谱** - 创建包含实体和文档块的知识图谱
4. **向量化存储** - 将文档块、实体、事实转换为向量并存储

#### 核心组件

- **pkg/hipporag/index.go** - 索引主流程
- **pkg/embedding/** - 向量化和存储
- **pkg/openie/** - 实体关系提取
- **pkg/graph/** - 知识图谱构建
- **pkg/utils/** - 文本处理、向量计算等工具

## 快速开始

**👉 查看 [QUICKSTART.md](QUICKSTART.md) 获取详细的快速开始指南**

**👉 查看 [DEMO.md](DEMO.md) 了解如何运行对比演示**

### 环境要求

- Go 1.21+
- OpenAI API Key
- (可选) Docker - 用于运行 Weaviate

### 环境配置

1. 复制环境变量模板：
```bash
cp .env.example .env
```

2. 编辑 `.env` 文件，填入你的 OpenAI API Key：
```bash
OPENAI_API_KEY=your-actual-api-key-here
```

⚠️ **重要**: `.env` 文件包含敏感信息，已被 `.gitignore` 忽略，不会被提交到 Git。

### 最简单的开始方式（无需 Docker）

```bash
# 1. 配置环境变量（见上方）

# 2. 下载依赖
go mod download

# 3. 运行示例（使用内存存储）
go run examples/index_example.go
```

### 使用 Weaviate（推荐用于生产）

### 使用 Weaviate（推荐用于生产）

首先安装 Docker Desktop：
- **Homebrew**: `brew install --cask docker`
- **官网下载**: https://www.docker.com/products/docker-desktop/

然后：

1. 启动 Weaviate：
```bash
# 新版 Docker
docker compose up -d

# 或旧版
docker-compose up -d
```

2. 运行示例：
```bash
go run examples/weaviate_example.go
```

## 演示对比

我们提供了2个命令来对比传统 RAG 和 HippoRAG 的效果：

```bash
# 传统 RAG（向量检索 + LLM）
make traditional

# HippoRAG（知识图谱 + PPR + LLM）
make hippo
```

**测试问题**: "爱因斯坦出生于哪个世纪？"

这个问题需要结合两个文档才能回答：
- 文档2: "爱因斯坦于1879年3月14日出生于德国乌尔姆"
- 文档4: "19世纪是指1801年到1900年这段时期"

**预期结果**:
- **传统 RAG**: 可能只检索到包含"爱因斯坦"的文档，遗漏"19世纪"定义
- **HippoRAG**: 通过知识图谱发现"1879"和"19世纪"的关联，检索到两个文档

详细说明请查看 [DEMO.md](DEMO.md)

## 使用示例

### 索引文档（内存存储）

```go
package main

import (
    "context"
    "log"
    
    "github.com/example/go-scaffold/pkg/embedding"
    "github.com/example/go-scaffold/pkg/hipporag"
    "github.com/example/go-scaffold/pkg/llm"
)

func main() {
    // 创建客户端
    embeddingClient := embedding.NewOpenAIClient(apiKey, "text-embedding-3-small")
    llmClient := llm.NewOpenAIClient(apiKey, "gpt-4o-mini")
    
    // 创建 HippoRAG 实例（使用内存存储）
    config := hipporag.DefaultConfig()
    rag := hipporag.NewHippoRAG(config, embeddingClient, llmClient)
    
    // 索引文档
    docs := []string{
        "Your document text here...",
        "Another document...",
    }
    
    ctx := context.Background()
    if err := rag.Index(ctx, docs); err != nil {
        log.Fatal(err)
    }
    
    // 查看统计信息
    stats := rag.Stats(ctx)
    // stats["chunks"], stats["entities"], stats["facts"], etc.
}
```

### 索引文档（Weaviate 存储）

```go
package main

import (
    "context"
    "log"
    
    "github.com/example/go-scaffold/pkg/embedding"
    "github.com/example/go-scaffold/pkg/hipporag"
    "github.com/example/go-scaffold/pkg/llm"
)

func main() {
    // 创建客户端
    embeddingClient := embedding.NewOpenAIClient(apiKey, "text-embedding-3-small")
    llmClient := llm.NewOpenAIClient(apiKey, "gpt-4o-mini")
    
    // 创建 Weaviate 存储
    weaviateConfig := embedding.WeaviateConfig{
        Host:      "localhost:8080",
        Scheme:    "http",
        ClassName: "DocumentChunk",
    }
    
    chunkStore, _ := embedding.NewWeaviateStore(weaviateConfig, embeddingClient)
    
    // 为实体和事实创建不同的集合
    entityConfig := weaviateConfig
    entityConfig.ClassName = "Entity"
    entityStore, _ := embedding.NewWeaviateStore(entityConfig, embeddingClient)
    
    factConfig := weaviateConfig
    factConfig.ClassName = "Fact"
    factStore, _ := embedding.NewWeaviateStore(factConfig, embeddingClient)
    
    // 创建 HippoRAG 实例（使用 Weaviate）
    config := hipporag.DefaultConfig()
    rag := hipporag.NewHippoRAGWithStores(
        config,
        embeddingClient,
        llmClient,
        chunkStore,
        entityStore,
        factStore,
    )
    
    // 索引文档
    ctx := context.Background()
    if err := rag.Index(ctx, docs); err != nil {
        log.Fatal(err)
    }
}
```

### 配置参数

```go
config := &hipporag.Config{
    ChunkSize:    512,   // 每块字符数
    ChunkOverlap: 50,    // 块重叠字符数
    PPRDamping:   0.5,   // PPR 阻尼系数
    PPRMaxIter:   100,   // PPR 最大迭代次数
    PPRTolerance: 1e-6,  // PPR 收敛阈值
    TopKEntities: 10,    // 检索的实体数量
    TopKChunks:   5,     // 返回的文档块数量
}
```

## 架构说明

### HippoRAG 工作原理

HippoRAG 模拟人脑海马体的记忆机制：

1. **编码阶段（Index）**
   - 将文档分解为语义单元（chunks）
   - 提取实体和关系，构建知识图谱
   - 向量化存储，支持相似度检索

2. **检索阶段（Retrieve）** - 待实现
   - 向量检索找到相关实体
   - 使用 PPR 在图上传播，找到关联信息
   - 返回最相关的文档块

3. **生成阶段（Query）** - 待实现
   - 基于检索结果生成答案
   - 使用 LLM 进行推理和总结

### 知识图谱结构

- **节点类型**
  - `entity`: 实体节点（人名、地名、概念等）
  - `chunk`: 文档块节点

- **边类型**
  - `fact`: 实体间的关系边（来自三元组）
  - `passage`: 文档块到实体的边
  - `synonymy`: 同义实体边（待实现）

## 待实现功能

- [ ] 检索功能 (Retrieve)
- [ ] 问答功能 (Query)
- [ ] 同义实体识别
- [ ] 持久化存储
- [ ] 批量索引优化
- [ ] 增量索引

## 技术栈

- **Embedding**: OpenAI text-embedding-3-small
- **LLM**: OpenAI gpt-4o-mini
- **图算法**: Personalized PageRank
- **向量检索**: 余弦相似度

## License

MIT


## 向量存储选择

### 内存存储 (Store)

**优点：**
- 零配置，开箱即用
- 快速原型开发
- 适合小规模数据（< 10,000 条）

**缺点：**
- 数据不持久化（重启丢失）
- 线性搜索，大规模数据慢
- 内存占用大

**使用场景：**
- 开发和测试
- 小规模应用
- 学习和理解原理

### Weaviate 存储 (WeaviateStore)

**优点：**
- 数据持久化
- 高效的 ANN 搜索（HNSW 算法）
- 支持大规模数据（百万级+）
- 支持混合搜索（向量 + 关键词）
- 分布式部署

**缺点：**
- 需要额外的服务部署
- 配置相对复杂

**使用场景：**
- 生产环境
- 大规模数据
- 需要持久化存储

### 性能对比

| 特性 | 内存存储 | Weaviate |
|------|---------|----------|
| 搜索速度（1万条） | ~10ms | ~1ms |
| 搜索速度（100万条） | ~1s | ~5ms |
| 内存占用 | 高 | 低 |
| 持久化 | ❌ | ✅ |
| 分布式 | ❌ | ✅ |

