package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger 创建 Gin 请求日志中间件
func Logger(log *zap.Logger) gin.HandlerFunc {
	if log == nil {
		log = zap.NewNop()
	}

	return func(ctx *gin.Context) {
		start := time.Now()
		path := ctx.Request.URL.Path
		query := ctx.Request.URL.RawQuery

		ctx.Next()

		fields := []zap.Field{
			zap.String("method", ctx.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", ctx.Writer.Status()),
			zap.String("ip", ctx.ClientIP()),
			zap.String("user_agent", ctx.Request.UserAgent()),
			zap.Duration("latency", time.Since(start)),
		}
		if len(ctx.Errors) > 0 {
			fields = append(fields, zap.String("errors", ctx.Errors.String()))
		}

		log.Info("HTTP 请求完成", fields...)
	}
}
