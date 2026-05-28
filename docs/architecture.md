# 架构说明

Solvify-Agent 采用模板化分层结构。`internal/app.App` 是全局应用结构体，集中持有项目配置、日志、路由、服务和 HTTP Server。

```text
cmd/server
  -> internal/app.App
     -> pkg/config
     -> pkg/logger
     -> internal/api
        -> internal/middleware
        -> pkg/response
        -> pkg/errors
     -> internal/service
     -> internal/agent
        -> internal/rag
        -> internal/tool
        -> internal/llm
```

## 启动流程

1. `cmd/server/main.go` 创建 `app.NewApp()`
2. `App.Initialize()` 加载配置、初始化日志、装配依赖、初始化路由和 HTTP Server
3. `App.Run()` 启动服务并监听退出信号
4. 收到退出信号后，`App` 使用 `server.shutdown_timeout_seconds` 控制优雅关闭超时时间

## 模块边界

- `cmd/server` 只负责创建、初始化并运行 App
- `pkg/config` 负责配置结构、加载、环境变量覆盖和全局配置访问
- `internal/app` 负责持有全局配置，并编排依赖初始化、路由初始化、服务运行和优雅关闭
- `internal/api` 负责 Gin 路由注册、请求解析和响应编码
- `internal/middleware` 负责 Gin 请求日志和 panic 恢复
- `internal/service` 负责业务用例封装
- `internal/agent` 负责编排 RAG、Tool、LLM
- `internal/rag` 提供内存知识检索示例
- `internal/tool` 提供可校验参数的工具调用示例
- `internal/llm` 封装 Eino 消息模型和 Mock LLM
- `pkg/logger` 提供 zap + lumberjack 日志初始化
- `pkg/errors` 提供业务错误码和业务错误封装
- `pkg/response` 提供统一 API 响应结构

## 请求链路

```text
HTTP Client
  -> Gin Router
  -> middleware.Recovery
  -> middleware.Logger
  -> api.Router.ask
  -> service.ChatService
  -> agent.KnowledgeAgent
     -> rag.Retriever
     -> tool.Tool
     -> llm.Client
  -> response.Success 或 response.BizError
```

## 配置结构

服务监听配置统一放在 `server` 节点：

```yaml
server:
  host: ""
  port: 8080
  shutdown_timeout_seconds: 10
```

`server.host` 为空时表示监听所有地址，最终监听地址由 `pkg/config.ServerConfig.Addr()` 生成。

## 响应规范

成功响应：

```json
{
  "code": 0,
  "message": "成功",
  "data": {}
}
```

错误响应：

```json
{
  "code": 4003,
  "message": "工具调用失败"
}
```

## 设计取舍

- 路由统一使用 Gin，不再使用原生 `ServeMux`
- `net/http` 只作为底层 HTTP Server 承载 Gin Engine
- 日志使用 zap，文件轮转使用 lumberjack
- LLM 当前为 Mock 实现，但消息结构使用 Eino `schema.Message`
- RAG 当前为内存检索实现，便于本地直接运行和测试
