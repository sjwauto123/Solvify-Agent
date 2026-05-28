package agent

import (
	"context"
	"strings"
	"testing"

	"solvify-agent/internal/llm"
	"solvify-agent/internal/rag"
	"solvify-agent/internal/tool"
)

// TestKnowledgeAgentRunsRAGToolAndLLM 验证 Agent 串联 RAG、Tool 和 LLM
func TestKnowledgeAgentRunsRAGToolAndLLM(t *testing.T) {
	t.Parallel()

	agent := NewKnowledgeAgent(Options{
		LLM: llm.NewMockClient("mock-model"),
		Retriever: rag.NewMemoryRetriever([]rag.Document{
			{ID: "tool", Title: "工具调用", Content: "工具调用前需要校验参数"},
		}),
		Tools: []tool.Tool{tool.NewCalculator()},
	})

	got, err := agent.Run(context.Background(), Request{
		Question: "工具调用要注意什么？另外计算 1 + 2",
		UseRAG:   true,
		UseTools: true,
	})
	if err != nil {
		t.Fatalf("run agent: %v", err)
	}

	if got.Answer == "" {
		t.Fatal("expected answer")
	}
	if !got.RAGHit {
		t.Fatal("expected rag hit")
	}
	if len(got.ToolResults) != 1 || got.ToolResults[0].Content != "3" {
		t.Fatalf("unexpected tool results: %+v", got.ToolResults)
	}
	if !strings.Contains(got.TraceID, "agent-") {
		t.Fatalf("unexpected trace id: %q", got.TraceID)
	}
}
