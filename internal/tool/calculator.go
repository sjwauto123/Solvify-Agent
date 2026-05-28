package tool

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const calculatorName = "calculator"

// Calculator 提供受控的加法计算工具示例
type Calculator struct{}

// NewCalculator 创建计算器工具
func NewCalculator() *Calculator {
	return &Calculator{}
}

// Name 返回工具名称
func (c *Calculator) Name() string {
	return calculatorName
}

// Call 校验参数并执行加法表达式计算
func (c *Calculator) Call(ctx context.Context, req CallRequest) (CallResult, error) {
	if err := ctx.Err(); err != nil {
		return CallResult{}, err
	}
	if err := ValidateName(req, calculatorName); err != nil {
		return CallResult{}, err
	}

	expression, ok := req.Args["expression"].(string)
	if !ok || strings.TrimSpace(expression) == "" {
		return CallResult{}, errors.New("calculator expression is required")
	}

	value, err := evalAddition(expression)
	if err != nil {
		return CallResult{}, err
	}

	return CallResult{
		Name:    calculatorName,
		Content: strconv.Itoa(value),
	}, nil
}

// evalAddition 解析仅允许整数加法的表达式
func evalAddition(expression string) (int, error) {
	parts := strings.Split(expression, "+")
	if len(parts) < 2 {
		return 0, errors.New("calculator only supports integer addition")
	}

	total := 0
	for _, part := range parts {
		value, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return 0, fmt.Errorf("invalid calculator expression: %w", err)
		}
		total += value
	}
	return total, nil
}
