package api

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes 统一注册所有外部接口
func RegisterRoutes(r *gin.Engine) {
	// 1. WebSocket 长连接
	r.GET("/ws/minecraft", HandleMinecraft)
	r.GET("/ws/web", HandleWeb)

	// 2. 咱们的全新聚合接口！（干掉旧的 icon 和 item-name）
	r.GET("/api/item/:id", HandleItemInfo)

	// 3. 其他业务接口
	r.GET("/config/whitelist", HandleConfig)
	r.POST("/config/whitelist", HandleConfigUpdate)
	r.PUT("/config/whitelist", HandleConfigUpdate)
	r.GET("/autocraft/craftables", HandleAutoCraftables)
	r.GET("/autocraft/recipe", HandleAutoCraftRecipe)
	r.GET("/autocraft/patterns", HandlePatterns)
	r.GET("/autocraft/tasks", HandleAutoCraftTasks)
	r.POST("/autocraft/tasks", HandleAutoCraftTasks)
	r.DELETE("/autocraft/tasks/:itemId", HandleAutoCraftTaskDelete)
	r.PATCH("/autocraft/tasks/:itemId", HandleAutoCraftTaskPatch)

	// 4. 静态资源挂载
	r.Static("/lua", "./lua_scripts")
}
