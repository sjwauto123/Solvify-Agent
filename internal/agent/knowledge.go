package agent

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"

	"solvify-agent/internal/llm"
	"solvify-agent/internal/rag"
	"solvify-agent/internal/tool"
	apperrors "solvify-agent/pkg/errors"
)

// Request 描述知识助理输入
type Request struct {
	Question string `json:"question"`
	UseRAG   bool   `json:"use_rag"`
	UseTools bool   `json:"use_tools"`
}

// Response 描述知识助理输出
type Response struct {
	Answer      string            `json:"answer"`
	TraceID     string            `json:"trace_id"`
	RAGHit      bool              `json:"rag_hit"`
	Documents   []rag.Document    `json:"documents,omitempty"`
	ToolResults []tool.CallResult `json:"tool_results,omitempty"`
}

// Options 描述知识助理依赖
type Options struct {
	LLM       llm.Client
	Retriever rag.Retriever
	Tools     []tool.Tool
	Logger    *zap.Logger
	Model     string
}

// KnowledgeAgent 编排 RAG、Tool 和 LLM 调用
type KnowledgeAgent struct {
	llm       llm.Client
	retriever rag.Retriever
	tools     map[string]tool.Tool
	logger    *zap.Logger
	model     string
}

// NewKnowledgeAgent 创建知识助理 Agent
func NewKnowledgeAgent(opts Options) *KnowledgeAgent {
	tools := make(map[string]tool.Tool, len(opts.Tools))
	for _, item := range opts.Tools {
		if item != nil {
			tools[item.Name()] = item
		}
	}

	logger := opts.Logger
	if logger == nil {
		logger = zap.NewNop()
	}

	return &KnowledgeAgent{
		llm:       opts.LLM,
		retriever: opts.Retriever,
		tools:     tools,
		logger:    logger,
		model:     opts.Model,
	}
}

// Run 执行知识助理链路并返回最终答案
func (a *KnowledgeAgent) Run(ctx context.Context, req Request) (Response, error) {
	if err := ctx.Err(); err != nil {
		return Response{}, apperrors.WrapDefault(apperrors.CodeAgentRunTimeout, err)
	}
	if strings.TrimSpace(req.Question) == "" {
		return Response{}, apperrors.NewDefault(apperrors.CodeQuestionRequired)
	}
	if a.llm == nil {
		return Response{}, apperrors.New(apperrors.CodeLLMCallFailed, "LLM 客户端未初始化")
	}

	traceID := newTraceID()
	a.logger.Info("Agent 开始执行", zap.String("trace_id", traceID))

	ragResult, err := a.retrieve(ctx, req)
	if err != nil {
		return Response{}, apperrors.WrapDefault(apperrors.CodeRAGFailed, err)
	}

	toolResults, err := a.callTools(ctx, req)
	if err != nil {
		return Response{}, err
	}

	messages := buildMessages(req.Question, ragResult, toolResults)
	generated, err := a.llm.Generate(ctx, llm.GenerateRequest{
		Messages: messages,
		Model:    a.model,
	})
	if err != nil {
		return Response{}, apperrors.WrapDefault(apperrors.CodeLLMCallFailed, err)
	}

	a.logger.Info("Agent 执行完成", zap.String("trace_id", traceID), zap.Bool("rag_hit", ragResult.Hit), zap.Int("tool_count", len(toolResults)))
	return Response{
		Answer:      generated.Message.Content,
		TraceID:     traceID,
		RAGHit:      ragResult.Hit,
		Documents:   ragResult.Documents,
		ToolResults: toolResults,
	}, nil
}

// retrieve 按请求开关执行 RAG 检索
func (a *KnowledgeAgent) retrieve(ctx context.Context, req Request) (rag.Result, error) {
	if !req.UseRAG || a.retriever == nil {
		return rag.Result{Fallback: "RAG 已关闭，将直接使用问题和工具结果回答"}, nil
	}

	result, err := a.retriever.Retrieve(ctx, rag.Query{Question: req.Question, TopK: 3})
	if err != nil {
		return rag.Result{}, fmt.Errorf("retrieve knowledge: %w", err)
	}
	return result, nil
}

// callTools 根据问题意图执行可选工具调用
func (a *KnowledgeAgent) callTools(ctx context.Context, req Request) ([]tool.CallResult, error) {
	if !req.UseTools {
		return nil, nil
	}

	expression := extractAddition(req.Question)
	if expression == "" {
		return nil, nil
	}

	calculator, ok := a.tools["calculator"]
	if !ok {
		return nil, apperrors.New(apperrors.CodeToolNotFound, "calculator 工具未注册")
	}

	result, err := calculator.Call(ctx, tool.CallRequest{
		Name: "calculator",
		Args: map[string]any{
			"expression": expression,
		},
	})
	if err != nil {
		return nil, apperrors.WrapDefault(apperrors.CodeToolCallFailed, err)
	}
	return []tool.CallResult{result}, nil
}

// buildMessages 构造 Eino 消息上下文
func buildMessages(question string, ragResult rag.Result, toolResults []tool.CallResult) []*schema.Message {
	system := "你是 Solvify 知识助理，请基于上下文给出简洁、可靠、可追踪的回答"
	contextLines := []string{"问题：" + question}

	if ragResult.Hit {
		for _, doc := range ragResult.Documents {
			contextLines = append(contextLines, fmt.Sprintf("知识片段[%s]：%s", doc.Title, doc.Content))
		}
	} else if ragResult.Fallback != "" {
		contextLines = append(contextLines, "检索状态："+ragResult.Fallback)
	}

	for _, result := range toolResults {
		contextLines = append(contextLines, fmt.Sprintf("工具结果[%s]：%s", result.Name, result.Content))
	}

	return []*schema.Message{
		schema.SystemMessage(system),
		schema.UserMessage(strings.Join(contextLines, "\n")),
	}
}

var additionPattern = regexp.MustCompile(`\d+(?:\s*\+\s*\d+)+`)

// extractAddition 从用户问题中提取简单加法表达式
func extractAddition(question string) string {
	return additionPattern.FindString(question)
}

// newTraceID 生成简易请求追踪编号
func newTraceID() string {
	return fmt.Sprintf("agent-%d", time.Now().UnixNano())
}
