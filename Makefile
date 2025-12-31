.PHONY: help rag hippo build-all clean

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

rag: ## 传统 RAG（向量检索 + LLM）
	@echo "运行传统 RAG 演示"
	@go run cmd/demo1_traditional_rag_qa/main.go

hippo: ## HippoRAG（知识图谱 + PPR + LLM）
	@echo "运行 HippoRAG 演示"
	@go run cmd/demo3_hipporag_qa/main.go

build-all: ## 编译所有演示程序
	@echo "编译演示程序..."
	@mkdir -p bin
	@go build -o bin/rag cmd/demo1_traditional_rag_qa/main.go
	@go build -o bin/hippo cmd/demo3_hipporag_qa/main.go
	@echo "✓ 编译完成，可执行文件在 bin/ 目录"

clean: ## 清理编译文件
	@rm -rf bin/
	@echo "✓ 清理完成"

test: ## 运行测试
	@go test ./...

deps: ## 下载依赖
	@go mod download
	@go mod tidy
