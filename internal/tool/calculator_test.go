package tool

import (
	"context"
	"testing"
)

// TestCalculatorToolRejectsInvalidExpression 验证计算器会拒绝非法表达式
func TestCalculatorToolRejectsInvalidExpression(t *testing.T) {
	t.Parallel()

	_, err := NewCalculator().Call(context.Background(), CallRequest{
		Name: "calculator",
		Args: map[string]any{
			"expression": "1 + abc",
		},
	})
	if err == nil {
		t.Fatal("expected invalid expression error")
	}
}

// TestCalculatorToolCalculatesAddition 验证计算器可以执行整数加法
func TestCalculatorToolCalculatesAddition(t *testing.T) {
	t.Parallel()

	got, err := NewCalculator().Call(context.Background(), CallRequest{
		Name: "calculator",
		Args: map[string]any{
			"expression": "12 + 30",
		},
	})
	if err != nil {
		t.Fatalf("call calculator: %v", err)
	}

	if got.Content != "42" {
		t.Fatalf("expected 42, got %q", got.Content)
	}
}
