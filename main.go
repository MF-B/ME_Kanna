package main

import (
	"log"
	"mineCCT/internal/api" // 引入 api 包
	"mineCCT/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 启动后台任务 (清理长期无响应工厂的产率)
	service.StartBackgroundTasks()

	// 注册路由 - 全部委托给 api 包处理
	r.GET("/ws/minecraft", api.HandleMinecraft)
	r.GET("/ws/web", api.HandleWeb)
	r.GET("/icon/:id", api.HandleIcon)
	r.GET("/item-name/:id", api.HandleItemName)
	r.GET("/config/whitelist", api.HandleConfig)

	r.Static("/lua", "./lua_scripts")

	port := ":8080"
	log.Printf("MineDock 模块化重构版启动! 监听: [::]%s", port)

	if err := r.Run("[::]" + port); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}
