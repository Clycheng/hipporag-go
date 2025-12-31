package utils

// text.go - 文本处理工具
// 用途：文档分块、文本清洗等预处理功能
// 主要功能：
// - ChunkText: 将长文档切分成固定大小的块（带重叠）
// - CleanText: 清理文本中的多余空白字符

import (
	"strings"
	"unicode"
)

// ChunkText 将文本分割成固定大小的块，支持重叠
// text: 输入文本
// chunkSize: 每块的字符数
// overlap: 块之间的重叠字符数
// 返回：文本块数组
func ChunkText(text string, chunkSize, overlap int) []string {
	if chunkSize <= 0 {
		return []string{text}
	}

	text = CleanText(text)
	if len(text) <= chunkSize {
		return []string{text}
	}

	var chunks []string
	start := 0

	for start < len(text) {
		end := start + chunkSize
		if end > len(text) {
			end = len(text)
		}

		chunks = append(chunks, text[start:end])

		if end == len(text) {
			break
		}

		// 下一块从 (当前位置 + chunkSize - overlap) 开始
		start += chunkSize - overlap
	}

	return chunks
}

// CleanText 清理文本中的多余空白字符
// 将多个连续空白字符替换为单个空格，去除首尾空白
func CleanText(text string) string {
	// 替换所有空白字符为空格
	text = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return ' '
		}
		return r
	}, text)

	// 合并多个连续空格
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}

	return strings.TrimSpace(text)
}
