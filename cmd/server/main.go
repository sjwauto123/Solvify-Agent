package main

import (
	"fmt"

	"solvify-agent/internal/app"
)

// main 启动 Solvify Agent HTTP 服务
func main() {
	application := app.NewApp()
	if err := application.Initialize(); err != nil {
		panic(fmt.Sprintf("应用初始化失败: %v", err))
	}
	if err := application.Run(); err != nil {
		panic(fmt.Sprintf("应用运行失败: %v", err))
	}
}
