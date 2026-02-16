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
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	if err := service.InitWhitelist(); err != nil {
		log.Printf("Whitelist init failed: %v", err)
	}

	// 注册路由 - 全部委托给 api 包处理
	r.GET("/ws/minecraft", api.HandleMinecraft)
	r.GET("/ws/web", api.HandleWeb)
	r.GET("/icon/:id", api.HandleIcon)
	r.GET("/item-name/:id", api.HandleItemName)
	r.GET("/config/whitelist", api.HandleConfig)
	r.GET("/debug/bridge", api.HandleBridgeDebug)
	r.POST("/debug/craft", api.HandleDebugCraft)
	r.POST("/config/whitelist", api.HandleConfigUpdate)
	r.PUT("/config/whitelist", api.HandleConfigUpdate)
	r.GET("/autocraft/craftables", api.HandleAutoCraftables)
	r.GET("/autocraft/recipe", api.HandleAutoCraftRecipe)
	r.GET("/autocraft/tasks", api.HandleAutoCraftTasks)
	r.POST("/autocraft/tasks", api.HandleAutoCraftTasks)
	r.DELETE("/autocraft/tasks/:itemId", api.HandleAutoCraftTaskDelete)
	r.PATCH("/autocraft/tasks/:itemId", api.HandleAutoCraftTaskPatch)

	r.Static("/lua", "./lua_scripts")

	port := ":8080"
	log.Printf("MineDock 模块化重构版启动! 监听: [::]%s", port)

	if err := r.Run("[::]" + port); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}
