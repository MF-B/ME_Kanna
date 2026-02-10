package main

import (
	"log"
	"mineCCT/internal/api" // 引入 api 包

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// 注册路由 - 全部委托给 api 包处理
	r.GET("/ws/minecraft", api.HandleMinecraft)
	r.GET("/ws/web", api.HandleWeb)
	r.GET("/icon/:id", api.HandleIcon)
	r.GET("/config/whitelist", api.HandleConfig)

	r.Static("/lua", "./lua_scripts")

	port := ":8080"
	log.Printf("MineDock 模块化重构版启动! 监听: [::]%s", port)

	if err := r.Run("[::]" + port); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}
