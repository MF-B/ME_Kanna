package api

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes 统一注册所有外部接口
func RegisterRoutes(r *gin.Engine) {
	// 1. WebSocket 长连接
	r.GET("/ws/collector", HandleCollector)
	r.GET("/ws/frontend", HandleFrontend)

	// 2. HTTP API
	r.GET("/api/itemInfo/:itemId", HandleItemInfo)
	r.GET("/api/autocraft/tasks", HandleListAutoCraftTasks)
	r.POST("/api/autocraft/tasks", HandleCreateAutoCraftTask)
	r.DELETE("/api/autocraft/tasks/:itemId", HandleDeleteAutoCraftTask)
	r.PATCH("/api/autocraft/tasks/:itemId", HandlePatchAutoCraftTask)

	r.GET("/api/craftables", HandleGetCraftables)
	r.POST("/api/craftables/scan", HandleTriggerCraftablesScan)

	// 3. 静态资源挂载
	r.Static("/lua", "../collector")
	r.Static("/icons", "../.minecraft/icon-exports-x32")
}
