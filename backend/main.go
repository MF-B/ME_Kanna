package main

import (
	"ME_Kanna/internal/api"
	"ME_Kanna/internal/utils"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// CORS 中间件
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

	// 初始化图标索引
	utils.InitIconIndex("../.minecraft/icon-exports-x32")

	// 注册路由
	api.RegisterRoutes(r)

	port := ":8080"
	log.Printf("ME_Kanna 启动, 监听: [::]%s", port)

	if err := r.Run("[::]" + port); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}
