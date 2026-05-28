package middleware

import (
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"solvify-agent/pkg/response"
)

// Recovery 创建 Gin panic 恢复中间件
func Recovery(log *zap.Logger) gin.HandlerFunc {
	if log == nil {
		log = zap.NewNop()
	}

	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error("请求发生 panic",
					zap.Any("error", err),
					zap.String("stack", string(debug.Stack())),
					zap.String("path", ctx.Request.URL.Path),
					zap.String("method", ctx.Request.Method),
					zap.String("ip", ctx.ClientIP()),
				)

				response.InternalError(ctx, "服务内部错误")
				ctx.Abort()
			}
		}()

		ctx.Next()
	}
}
