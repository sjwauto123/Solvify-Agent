package service

import (
	"context"
	"fmt"

	"solvify-agent/internal/agent"
)

// ChatService 封装知识助理业务用例
type ChatService struct {
	agent *agent.KnowledgeAgent
}

// NewChatService 创建知识助理业务服务
func NewChatService(agent *agent.KnowledgeAgent) *ChatService {
	return &ChatService{agent: agent}
}

// Ask 调用 Agent 完成一次问答
func (s *ChatService) Ask(ctx context.Context, req agent.Request) (agent.Response, error) {
	if s.agent == nil {
		return agent.Response{}, fmt.Errorf("knowledge agent is not initialized")
	}
	return s.agent.Run(ctx, req)
}
