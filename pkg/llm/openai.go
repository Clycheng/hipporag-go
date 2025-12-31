package llm

// openai.go - OpenAI LLM 客户端
// 用途：调用 OpenAI API 进行文本生成（用于 OpenIE 和 QA）
// 主要功能：
// - Complete: 文本补全/生成
// - 支持自定义模型、温度等参数

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenAIClient OpenAI LLM 客户端
type OpenAIClient struct {
	apiKey      string
	model       string
	temperature float64
	baseURL     string
	client      *http.Client
}

// NewOpenAIClient 创建 OpenAI 客户端
func NewOpenAIClient(apiKey, model string) *OpenAIClient {
	if model == "" {
		model = "gpt-4o-mini"
	}

	return &OpenAIClient{
		apiKey:      apiKey,
		model:       model,
		temperature: 0.0, // 默认确定性输出
		baseURL:     "https://api.openai.com/v1",
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// SetTemperature 设置温度参数
func (c *OpenAIClient) SetTemperature(temp float64) {
	c.temperature = temp
}

// OpenAI API 请求/响应结构
type completionRequest struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	Temperature float64   `json:"temperature"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type completionResponse struct {
	Choices []struct {
		Message message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// Complete 生成文本补全
func (c *OpenAIClient) Complete(ctx context.Context, prompt string) (string, error) {
	// 构造请求
	reqBody := completionRequest{
		Model: c.model,
		Messages: []message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: c.temperature,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	// 发送请求
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	var compResp completionResponse
	if err := json.Unmarshal(body, &compResp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	// 检查错误
	if compResp.Error != nil {
		return "", fmt.Errorf("openai api error: %s", compResp.Error.Message)
	}

	if len(compResp.Choices) == 0 {
		return "", fmt.Errorf("no completion returned")
	}

	return compResp.Choices[0].Message.Content, nil
}
