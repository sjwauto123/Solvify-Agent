package llm

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"
)

// Client 定义模型调用边界
type Client interface {
	Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error)
}

// GenerateRequest 描述一次 LLM 生成请求
type GenerateRequest struct {
	Messages []*schema.Message
	Model    string
}

// GenerateResponse 描述一次 LLM 生成响应
type GenerateResponse struct {
	Message *schema.Message
	Model   string
}

// MockClient 提供可离线运行的 Eino 消息模型示例
type MockClient struct {
	model string
}

// NewMockClient 创建 Mock LLM 客户端
func NewMockClient(model string) *MockClient {
	return &MockClient{model: model}
}

// Generate 基于输入消息生成可预测回答
func (c *MockClient) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	if err := ctx.Err(); err != nil {
		return GenerateResponse{}, err
	}
	if len(req.Messages) == 0 {
		return GenerateResponse{}, errors.New("messages are required")
	}

	question := lastUserContent(req.Messages)
	if strings.TrimSpace(question) == "" {
		return GenerateResponse{}, errors.New("user question is required")
	}

	model := req.Model
	if model == "" {
		model = c.model
	}

	content := fmt.Sprintf("基于当前上下文，我会按知识检索、工具结果和问题意图综合回答：%s", question)
	return GenerateResponse{
		Message: schema.AssistantMessage(content, nil),
		Model:   model,
	}, nil
}

// lastUserContent 提取最后一条用户消息内容
func lastUserContent(messages []*schema.Message) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i] != nil && messages[i].Role == schema.User {
			return messages[i].Content
		}
	}
	return ""
}
