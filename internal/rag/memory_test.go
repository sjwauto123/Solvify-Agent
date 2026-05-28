package rag

import (
	"context"
	"testing"
)

// TestMemoryRetrieverReturnsFallbackWhenNoDocumentMatches 验证未命中时返回兜底信息
func TestMemoryRetrieverReturnsFallbackWhenNoDocumentMatches(t *testing.T) {
	t.Parallel()

	retriever := NewMemoryRetriever([]Document{
		{ID: "agent", Title: "Agent", Content: "Agent 负责规划和工具调用"},
	})

	got, err := retriever.Retrieve(context.Background(), Query{Question: "数据库连接池如何配置"})
	if err != nil {
		t.Fatalf("retrieve: %v", err)
	}

	if got.Hit {
		t.Fatal("expected no hit")
	}
	if got.Fallback == "" {
		t.Fatal("expected fallback message")
	}
}

// TestMemoryRetrieverFindsRelevantDocument 验证关键词命中时返回相关文档
func TestMemoryRetrieverFindsRelevantDocument(t *testing.T) {
	t.Parallel()

	retriever := NewMemoryRetriever([]Document{
		{ID: "rag", Title: "RAG", Content: "RAG 会先检索知识库再组织回答"},
	})

	got, err := retriever.Retrieve(context.Background(), Query{Question: "RAG 怎么回答"})
	if err != nil {
		t.Fatalf("retrieve: %v", err)
	}

	if !got.Hit {
		t.Fatal("expected hit")
	}
	if len(got.Documents) != 1 || got.Documents[0].ID != "rag" {
		t.Fatalf("unexpected documents: %+v", got.Documents)
	}
}
