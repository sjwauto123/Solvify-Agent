package api

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"solvify-agent/internal/agent"
	"solvify-agent/internal/middleware"
	"solvify-agent/internal/service"
	"solvify-agent/pkg/response"
)

// Router 聚合 HTTP 路由依赖
type Router struct {
	service *service.ChatService
	logger  *zap.Logger
	engine  *gin.Engine
}

// NewRouter 创建 Gin 路由并注册接口
func NewRouter(chatService *service.ChatService, logger *zap.Logger) *Router {
	if logger == nil {
		logger = zap.NewNop()
	}

	engine := gin.New()
	router := &Router{
		service: chatService,
		logger:  logger,
		engine:  engine,
	}

	engine.Use(middleware.Recovery(logger))
	engine.Use(middleware.Logger(logger))
	router.routes(engine)
	return router
}

// Engine 返回 Gin 引擎实例
func (r *Router) Engine() *gin.Engine {
	return r.engine
}

// routes 注册 Gin HTTP 路由
func (r *Router) routes(engine *gin.Engine) {
	engine.GET("/health", r.health)
	v1 := engine.Group("/api/v1")
	v1.POST("/ask", r.ask)
}

// health 返回服务健康状态
func (r *Router) health(ctx *gin.Context) {
	response.Success(ctx, gin.H{
		"status":  "ok",
		"service": "solvify-agent",
	})
}

// ask 处理知识助理问答请求
func (r *Router) ask(ctx *gin.Context) {
	var input agent.Request
	if err := ctx.ShouldBindJSON(&input); err != nil {
		response.BadRequest(ctx, "请求体格式错误")
		return
	}

	output, err := r.service.Ask(ctx.Request.Context(), input)
	if err != nil {
		response.BizError(ctx, err)
		return
	}

	response.Success(ctx, output)
}
