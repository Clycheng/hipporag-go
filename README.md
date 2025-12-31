# HippoRAG - Go Implementation

HippoRAG (Hippocampus-Inspired Retrieval-Augmented Generation) 是一个受海马体启发的知识检索系统，结合了知识图谱和向量检索的优势。

## 快速开始

### 1. 环境配置

复制环境变量模板并配置：

```bash
cp .env.example .env
# 编辑 .env 文件，填入你的 OpenAI API Key
```

### 2. 安装依赖

```bash
make deps
```

### 3. 运行演示

```bash
# 传统 RAG（向量检索 + LLM）
make rag

# HippoRAG（知识图谱 + PPR + LLM）
make hippo
```

## 项目结构

```
├── cmd/
│   ├── traditional_rag/               # 传统 RAG 演示
│   └── hipporag/                      # HippoRAG 演示
├── pkg/
│   ├── hipporag/                      # HippoRAG 核心实现
│   ├── rag/                           # 传统 RAG 实现
│   ├── embedding/                     # 向量化和存储
│   ├── graph/                         # 知识图谱和 PPR
│   ├── openie/                        # 实体关系提取
│   ├── llm/                           # LLM 客户端
│   └── utils/                         # 工具函数
├── data/                              # 测试数据
├── Makefile                           # 命令定义
├── README.md                          # 本文件
└── DEMO.md                            # 详细演示指南
```

## 核心功能

### 传统 RAG

**流程**：
```
文档 → 向量化 → 向量存储
查询 → 向量化 → 相似度搜索 → Top-K 文档 → LLM 生成答案
```

**特点**：
- 简单直接
- 速度快
- 适合简单查询

### HippoRAG

**流程**：
```
文档 → 分块 → OpenIE 提取 → 知识图谱 → 向量存储
查询 → 事实检索 → LLM 重排序 → DPR → PPR 图传播 → Top-K 文档 → LLM 生成答案
```

**特点**：
- 知识图谱 + 向量检索
- 支持多跳推理
- 适合复杂查询

## 测试数据

项目包含两个测试文档集：

1. **TestDocuments**（8个文档）- 基础测试
2. **TestDocumentsWithNoise**（60个文档）- 噪音环境测试

测试问题：
- "爱因斯坦出生于哪个世纪？"（需要推理：1879 → 19世纪）
- "爱因斯坦的同学毕业的时候年龄是多少？"（需要多跳推理）

## 配置参数

### HippoRAG 配置

```go
config := hipporag.DefaultConfig()
config.ChunkSize = 100       // 文档块大小
config.ChunkOverlap = 0      // 块重叠大小
config.TopKEntities = 20     // 检索的实体数量
config.TopKChunks = 15       // 返回的文档块数量
config.PPRDamping = 0.3      // PPR 阻尼系数
config.PPRMaxIter = 100      // PPR 最大迭代次数
config.PPRTolerance = 1e-6   // PPR 收敛阈值
```

### 传统 RAG 配置

```go
topK := 5  // 返回的文档数量
```

## 可用命令

```bash
make help    # 显示帮助信息
make rag     # 运行传统 RAG 演示
make hippo   # 运行 HippoRAG 演示
make build   # 编译演示程序
make clean   # 清理编译文件
make test    # 运行测试
make deps    # 下载依赖
```

## 环境变量

在 `.env` 文件中配置：

```bash
# OpenAI API 配置
OPENAI_API_KEY=your-api-key-here
OPENAI_BASE_URL=https://api.agicto.cn/v1

# 代理配置（可选）
HTTPS_PROXY=http://127.0.0.1:7890
```

## 详细文档

查看 [DEMO.md](DEMO.md) 了解：
- 完整的工作流程
- 技术实现细节
- 参数调优指南
- 性能对比分析

## 技术栈

- **语言**: Go 1.21+
- **Embedding**: OpenAI text-embedding-3-small
- **LLM**: OpenAI gpt-4o-mini
- **图算法**: Personalized PageRank
- **向量检索**: 余弦相似度

## 核心算法

### Personalized PageRank (PPR)

在知识图谱上传播分数，找到间接相关的文档：

```
初始权重（种子节点）→ 迭代传播 → 收敛 → 文档排序
```

### OpenIE 信息抽取

使用 LLM 提取实体和关系三元组：

```
文档 → LLM → 实体列表 + 三元组 → 知识图谱
```

## 性能对比

| 特性 | 传统 RAG | HippoRAG |
|------|---------|----------|
| **索引速度** | 快 | 慢（需要 OpenIE） |
| **检索速度** | 快 | 中等（需要 PPR） |
| **简单查询** | 好 | 好 |
| **复杂查询** | 差 | 好 |
| **多跳推理** | 不支持 | 支持 |
| **抗噪音能力** | 弱 | 强 |

## License

MIT

## 参考资料

- [HippoRAG 论文](https://arxiv.org/abs/2405.14831)
- [Personalized PageRank](https://en.wikipedia.org/wiki/PageRank#Personalized_PageRank)
