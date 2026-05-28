package rag

import (
	"context"
	"strings"
)

// Document 描述知识库文档片段
type Document struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// Query 描述检索请求
type Query struct {
	Question string
	TopK     int
}

// Result 描述检索结果
type Result struct {
	Hit       bool       `json:"hit"`
	Documents []Document `json:"documents"`
	Fallback  string     `json:"fallback,omitempty"`
}

// Retriever 定义知识检索边界
type Retriever interface {
	Retrieve(ctx context.Context, query Query) (Result, error)
}

// MemoryRetriever 使用内存文档模拟 RAG 检索
type MemoryRetriever struct {
	documents []Document
}

// NewMemoryRetriever 创建内存检索器
func NewMemoryRetriever(documents []Document) *MemoryRetriever {
	copied := make([]Document, len(documents))
	copy(copied, documents)
	return &MemoryRetriever{documents: copied}
}

// Retrieve 按关键词命中内存文档并在未命中时返回兜底提示
func (r *MemoryRetriever) Retrieve(ctx context.Context, query Query) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	question := strings.TrimSpace(query.Question)
	if question == "" {
		return Result{Fallback: "问题为空，无法检索知识库"}, nil
	}

	topK := query.TopK
	if topK <= 0 {
		topK = 3
	}

	matches := make([]Document, 0, topK)
	for _, doc := range r.documents {
		if matchesQuestion(question, doc) {
			matches = append(matches, doc)
			if len(matches) >= topK {
				break
			}
		}
	}

	if len(matches) == 0 {
		return Result{
			Hit:      false,
			Fallback: "知识库暂未命中，将使用通用推理和工具结果回答",
		}, nil
	}

	return Result{
		Hit:       true,
		Documents: matches,
	}, nil
}

// SeedDocuments 返回 Demo 默认知识库
func SeedDocuments() []Document {
	return []Document{
		{ID: "agent-flow", Title: "Agent 执行链路", Content: "知识助理先理解问题，再执行 RAG 检索，可选调用工具，最后组织 LLM 回答"},
		{ID: "tool-validation", Title: "Tool 参数校验", Content: "工具调用前必须校验参数，避免模型生成的非法参数直接进入业务系统"},
		{ID: "rag-fallback", Title: "RAG 未命中兜底", Content: "RAG 未命中时应显式记录并走通用回答兜底，避免直接返回空结果"},
	}
}

// matchesQuestion 判断问题是否与文档存在简单关键词交集
func matchesQuestion(question string, doc Document) bool {
	text := strings.ToLower(doc.Title + " " + doc.Content)
	for _, token := range tokenize(question) {
		if strings.Contains(text, strings.ToLower(token)) {
			return true
		}
	}
	return false
}

// tokenize 将问题拆成用于 Demo 检索的关键词
func tokenize(question string) []string {
	replacer := strings.NewReplacer("？", " ", "?", " ", "，", " ", ",", " ", "。", " ", ".", " ", "+", " ")
	fields := strings.Fields(replacer.Replace(question))
	tokens := make([]string, 0, len(fields))
	for _, field := range fields {
		token := strings.TrimSpace(field)
		if token != "" {
			tokens = append(tokens, token)
			tokens = append(tokens, bigrams(token)...)
		}
	}
	return tokens
}

// bigrams 为连续中文短语补充二元片段以支持简单命中
func bigrams(token string) []string {
	runes := []rune(token)
	if len(runes) < 2 {
		return nil
	}

	result := make([]string, 0, len(runes)-1)
	for i := 0; i < len(runes)-1; i++ {
		result = append(result, string(runes[i:i+2]))
	}
	return result
}
