.PHONY: help demo1 demo2 demo3 demo4 build-all clean

# 加载 .env 文件
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

help: ## 显示帮助信息
	@echo "HippoRAG 演示命令"
	@echo "================"
	@echo ""
	@echo "使用方法: make <命令>"
	@echo ""
	@echo "可用命令:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "测试问题: 爱因斯坦出生于哪个世纪？"
	@echo ""

demo1: ## 命令1: 传统 RAG 检索 + LLM 生成答案
	@echo "运行命令1: 传统 RAG 检索 + LLM 生成答案"
	@go run cmd/demo1_traditional_rag_qa/main.go

demo2: ## 命令2: 传统 RAG 仅检索（不生成答案）
	@echo "运行命令2: 传统 RAG 仅检索"
	@go run cmd/demo2_traditional_rag_retrieve/main.go

demo3: ## 命令3: HippoRAG 检索 + LLM 生成答案
	@echo "运行命令3: HippoRAG 检索 + LLM 生成答案"
	@go run cmd/demo3_hipporag_qa/main.go

demo4: ## 命令4: HippoRAG 仅检索（不生成答案）
	@echo "运行命令4: HippoRAG 仅检索"
	@go run cmd/demo4_hipporag_retrieve/main.go

build-all: ## 编译所有演示程序
	@echo "编译所有演示程序..."
	@mkdir -p bin
	@go build -o bin/demo1 cmd/demo1_traditional_rag_qa/main.go
	@go build -o bin/demo2 cmd/demo2_traditional_rag_retrieve/main.go
	@go build -o bin/demo3 cmd/demo3_hipporag_qa/main.go
	@go build -o bin/demo4 cmd/demo4_hipporag_retrieve/main.go
	@echo "✓ 编译完成，可执行文件在 bin/ 目录"

clean: ## 清理编译文件
	@rm -rf bin/
	@echo "✓ 清理完成"

test: ## 运行测试
	@go test ./...

deps: ## 下载依赖
	@go mod download
	@go mod tidy
