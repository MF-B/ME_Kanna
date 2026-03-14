package main

import (
	"ME_Kanna/internal/api" // 引入 api 包
	"ME_Kanna/internal/service"
	"log"

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

	api.RegisterRoutes(r)

	port := ":8080"
	log.Printf("MineDock 启动, 监听: [::]%s", port)

	if err := r.Run("[::]" + port); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}
