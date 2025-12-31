.PHONY: help rag hippo build clean test deps

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
	@echo "  help            显示帮助信息"
	@echo "  rag             运行传统 RAG 演示"
	@echo "  hippo           运行 HippoRAG 演示（完整版）"
	@echo "  build           编译演示程序"
	@echo "  clean           清理编译文件"
	@echo "  test            运行测试"
	@echo "  deps            下载依赖"
	@echo ""

rag: ## 运行传统 RAG 演示
	@go run cmd/traditional_rag/main.go

hippo: ## 运行 HippoRAG 演示（完整版）
	@go run cmd/hipporag/main.go

build: ## 编译演示程序
	@echo "编译演示程序..."
	@mkdir -p bin
	@go build -o bin/traditional_rag cmd/traditional_rag/main.go
	@go build -o bin/hipporag cmd/hipporag/main.go
	@echo "✓ 编译完成，可执行文件在 bin/ 目录"

clean: ## 清理编译文件
	@rm -rf bin/
	@echo "✓ 清理完成"

test: ## 运行测试
	@go test ./...

deps: ## 下载依赖
	@go mod download
	@go mod tidy
