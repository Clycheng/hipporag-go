# HippoRAG-Go 项目结构

## 目录结构

```
hipporag-go/
├── cmd/                           # 可执行程序
│   ├── traditional_rag/           # 传统 RAG 演示
│   │   └── main.go
│   └── hipporag/                  # HippoRAG 演示
│       └── main.go
│
├── pkg/                           # 核心库
│   ├── embedding/                 # 向量化和存储
│   │   ├── interface.go           # 接口定义
│   │   ├── client.go              # 客户端封装
│   │   ├── openai.go              # OpenAI 实现
│   │   ├── store.go               # 向量存储
│   │   └── weaviate.go            # Weaviate 集成
│   │
│   ├── graph/                     # 知识图谱
│   │   ├── graph.go               # 图结构
│   │   └── ppr.go                 # PPR 算法
│   │
│   ├── hipporag/                  # HippoRAG 核心
│   │   ├── hipporag.go            # 主类
│   │   ├── index.go               # 索引实现
│   │   ├── retrieve.go            # 简单检索
│   │   ├── retrieve_full.go       # 完整检索（事实检索+LLM重排序+DPR+PPR）
│   │   └── qa.go                  # 问答实现
│   │
│   ├── llm/                       # LLM 客户端
│   │   └── openai.go              # OpenAI 实现
│   │
│   ├── openie/                    # 信息抽取
│   │   └── extractor.go           # OpenIE 提取器
│   │
│   ├── rag/                       # 传统 RAG
│   │   └── traditional.go         # 传统 RAG 实现
│   │
│   └── utils/                     # 工具函数
│       ├── hash.go                # 哈希和归一化
│       ├── text.go                # 文本处理
│       └── vector.go              # 向量计算
│
├── data/                          # 测试数据
│   └── documents.go               # 测试文档集（8个核心 + 52个干扰）
│
├── bin/                           # 编译输出（make build）
│
├── .env                           # 环境变量（不提交）
├── .env.example                   # 环境变量模板
├── .gitignore                     # Git 忽略文件
├── docker-compose.yml             # Docker 配置
├── go.mod                         # Go 模块定义
├── go.sum                         # Go 依赖锁定
├── LICENSE                        # 许可证
├── Makefile                       # 命令定义
├── README.md                      # 项目说明
├── DEMO.md                        # 演示指南
└── PROJECT_STRUCTURE.md           # 本文件
```

## 核心组件说明

### 1. 传统 RAG (`pkg/rag/`)

**功能**：简单的向量检索 + LLM 生成

**流程**：
```
文档 → 向量化 → 向量存储
查询 → 向量化 → 相似度搜索 → Top-K 文档 → LLM 生成答案
```

**文件**：
- `traditional.go`: 传统 RAG 实现

### 2. HippoRAG (`pkg/hipporag/`)

**功能**：知识图谱 + 向量检索 + PPR 传播

**流程**：
```
索引阶段：
文档 → 分块 → OpenIE 提取 → 知识图谱 → 向量存储

检索阶段（完整版）：
查询 → 事实检索 → LLM 重排序 → DPR → PPR 图传播 → Top-K 文档 → LLM 生成答案
```

**文件**：
- `hipporag.go`: 主类，配置和初始化
- `index.go`: 索引实现（分块、OpenIE、图构建）
- `retrieve.go`: 简单检索（实体检索 + PPR）
- `retrieve_full.go`: 完整检索（事实检索 + LLM重排序 + DPR + PPR）
- `qa.go`: 问答实现（Query 和 QueryFull）

### 3. 知识图谱 (`pkg/graph/`)

**功能**：图结构和 PPR 算法

**节点类型**：
- `chunk`: 文档块节点
- `entity`: 实体节点

**边类型**：
- `passage`: 文档块 ↔ 实体（双向）
- `passage_back`: 实体 → 文档块（PPR 传播）
- `fact`: 实体 ↔ 实体（双向）
- `fact_back`: 实体 ← 实体（PPR 传播）
- `synonymy`: 同义实体

**文件**：
- `graph.go`: 图结构定义和操作
- `ppr.go`: Personalized PageRank 算法

### 4. 向量化 (`pkg/embedding/`)

**功能**：文本向量化和向量存储

**实现**：
- OpenAI text-embedding-3-small
- 内存向量存储（余弦相似度搜索）
- Weaviate 集成（可选）

**文件**：
- `interface.go`: 接口定义
- `client.go`: 客户端封装
- `openai.go`: OpenAI 实现
- `store.go`: 向量存储
- `weaviate.go`: Weaviate 集成

### 5. 信息抽取 (`pkg/openie/`)

**功能**：从文档中提取实体和关系三元组

**实现**：使用 LLM 进行 OpenIE 提取

**输出**：
- 实体列表：["爱因斯坦", "1879年", "德国"]
- 三元组：[("爱因斯坦", "出生于", "1879年")]

**文件**：
- `extractor.go`: OpenIE 提取器

### 6. LLM 客户端 (`pkg/llm/`)

**功能**：LLM 调用封装

**实现**：OpenAI gpt-4o-mini

**文件**：
- `openai.go`: OpenAI 实现

### 7. 工具函数 (`pkg/utils/`)

**功能**：通用工具函数

**文件**：
- `hash.go`: 哈希计算、MinMax 归一化
- `text.go`: 文本分块、清理
- `vector.go`: 向量计算（余弦相似度、归一化）

## 演示程序

### 1. 传统 RAG (`cmd/traditional_rag/`)

**运行**：`make rag`

**功能**：
- 索引 60 个测试文档
- 交互式问答
- 展示向量检索效果

### 2. HippoRAG (`cmd/hipporag/`)

**运行**：`make hippo`

**功能**：
- 索引 60 个测试文档
- 构建知识图谱
- 使用完整版检索（事实检索 + LLM重排序 + DPR + PPR）
- 交互式问答
- 展示多跳推理能力

## 测试数据 (`data/`)

### TestDocuments（60个文档）

**核心文档（8个）**：
- 爱因斯坦相关（5个）
- 小明相关（3个）

**干扰文档（52个）**：
- 其他科学家（10个）
- 历史事件（10个）
- 地理信息（10个）
- 教育机构（10个）
- 时间相关（10个）
- 其他名人（2个）

### 测试问题

1. **爱因斯坦出生于哪个世纪？**
   - 需要推理：1879年 → 19世纪
   - 需要文档：文档2 + 文档4

2. **爱因斯坦的同学毕业的时候年龄是多少？**
   - 需要多跳推理：爱因斯坦 → 同学小明 → 1996年 → 2018年 → 22岁
   - 需要文档：文档6 + 文档7 + 文档8

## 配置文件

### .env

环境变量配置：

```bash
# OpenAI API 配置
OPENAI_API_KEY=your-api-key-here
OPENAI_BASE_URL=https://api.agicto.cn/v1

# 代理配置（可选）
HTTPS_PROXY=http://127.0.0.1:7890
```

### Makefile

可用命令：

```bash
make help    # 显示帮助信息
make rag     # 运行传统 RAG 演示
make hippo   # 运行 HippoRAG 演示
make build   # 编译演示程序
make clean   # 清理编译文件
make test    # 运行测试
make deps    # 下载依赖
```

## 技术栈

- **语言**: Go 1.21+
- **Embedding**: OpenAI text-embedding-3-small (1536维)
- **LLM**: OpenAI gpt-4o-mini
- **图算法**: Personalized PageRank
- **向量检索**: 余弦相似度

## 核心算法

### 1. Personalized PageRank (PPR)

**作用**：在知识图谱上传播分数，找到间接相关的文档

**参数**：
- `damping`: 阻尼系数（默认 0.5）
- `maxIter`: 最大迭代次数（默认 100）
- `tolerance`: 收敛阈值（默认 1e-6）

### 2. OpenIE 信息抽取

**作用**：从文档中提取结构化知识

**输入**：文档文本

**输出**：
- 实体列表
- 关系三元组

### 3. Recognition Memory（LLM 重排序）

**作用**：用 LLM 对候选事实进行重排序

**过程**：
1. 向量检索找到候选事实
2. LLM 理解语义后重新排序
3. 使用重排序后的结果

### 4. 密集段落检索（DPR）

**作用**：直接在文档块向量库中搜索相关段落

**过程**：
1. 查询向量化
2. 在 chunk_embeddings 中搜索
3. 返回 Top-K 相似段落

## 参数说明

### HippoRAG 配置

| 参数 | 默认值 | 说明 |
|------|--------|------|
| ChunkSize | 100 | 文档块大小 |
| ChunkOverlap | 0 | 块重叠大小 |
| TopKEntities | 20 | 检索的实体数量（PPR 种子节点） |
| TopKChunks | 15 | 返回的文档块数量（给 LLM） |
| PPRDamping | 0.5 | PPR 阻尼系数 |
| PPRMaxIter | 100 | PPR 最大迭代次数 |
| PPRTolerance | 1e-6 | PPR 收敛阈值 |

### 传统 RAG 配置

| 参数 | 默认值 | 说明 |
|------|--------|------|
| TopK | 3 | 返回的文档数量 |

## 性能对比

| 特性 | 传统 RAG | HippoRAG |
|------|---------|----------|
| **索引速度** | 快 | 慢（需要 OpenIE） |
| **检索速度** | 快 | 中等（需要 PPR） |
| **简单查询** | 好 | 好 |
| **复杂查询** | 差 | 好 |
| **多跳推理** | 不支持 | 支持 |
| **抗噪音能力** | 弱 | 强 |

## 开发指南

### 添加新的 Embedding 实现

1. 在 `pkg/embedding/` 中创建新文件
2. 实现 `EmbeddingClient` 接口
3. 在 `client.go` 中注册

### 添加新的 LLM 实现

1. 在 `pkg/llm/` 中创建新文件
2. 实现 `LLMClient` 接口
3. 在演示程序中使用

### 修改 PPR 算法

1. 编辑 `pkg/graph/ppr.go`
2. 调整参数或算法逻辑
3. 运行测试验证

### 添加新的测试文档

1. 编辑 `data/documents.go`
2. 添加到 `TestDocuments` 数组
3. 运行演示验证

## 故障排查

### 问题：API 调用失败

**解决方案**：
- 检查 `.env` 文件中的 `OPENAI_API_KEY`
- 测试代理：`curl -x http://127.0.0.1:7890 https://api.agicto.cn/v1/models`
- 确认 `OPENAI_BASE_URL` 正确

### 问题：检索结果为空

**解决方案**：
- 检查图构建时是否添加了双向边
- 调整 PPR 参数（damping, maxIter）
- 查看 OpenIE 提取的实体数量

### 问题：检索到的文档不相关

**解决方案**：
- 增加 `TopKEntities`（如 10 → 20）
- 增加 `TopKChunks`（如 5 → 10）
- 降低 `PPRDamping`（如 0.5 → 0.3）

## 参考资料

- [HippoRAG 论文](https://arxiv.org/abs/2405.14831)
- [Personalized PageRank](https://en.wikipedia.org/wiki/PageRank#Personalized_PageRank)
- [OpenIE](https://nlp.stanford.edu/software/openie.html)

## License

MIT
