package tool

import (
	"context"
	"errors"
)

// CallRequest 描述工具调用请求
type CallRequest struct {
	Name string         `json:"name"`
	Args map[string]any `json:"args"`
}

// CallResult 描述工具调用结果
type CallResult struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// Tool 定义 Agent 可调用工具边界
type Tool interface {
	Name() string
	Call(ctx context.Context, req CallRequest) (CallResult, error)
}

// ValidateName 校验请求中的工具名是否匹配当前工具
func ValidateName(req CallRequest, expected string) error {
	if req.Name == "" {
		return errors.New("tool name is required")
	}
	if req.Name != expected {
		return errors.New("tool name mismatch")
	}
	return nil
}
