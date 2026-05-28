package errors

import "fmt"

// BizError 描述业务错误
type BizError struct {
	Code    int
	Message string
	Err     error
}

// Error 实现 error 接口
func (e *BizError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 返回原始错误
func (e *BizError) Unwrap() error {
	return e.Err
}

// New 创建业务错误
func New(code int, message string) *BizError {
	return &BizError{
		Code:    code,
		Message: message,
	}
}

// NewWithErr 创建带原始错误的业务错误
func NewWithErr(code int, message string, err error) *BizError {
	return &BizError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// NewDefault 使用默认错误消息创建业务错误
func NewDefault(code int) *BizError {
	return New(code, GetMessage(code))
}

// WrapDefault 使用默认错误消息包装原始错误
func WrapDefault(code int, err error) *BizError {
	return NewWithErr(code, GetMessage(code), err)
}
