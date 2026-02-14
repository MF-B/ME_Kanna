package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mineCCT/internal/model"
	"mineCCT/internal/service"
	"mineCCT/internal/store"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func parseWhitelistVersion(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		return str
	}
	var num float64
	if err := json.Unmarshal(raw, &num); err == nil {
		return fmt.Sprintf("%v", num)
	}
	return ""
}

func HandleMinecraft(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	var currentDeviceID string

	log.Println("[MC] New Device Connecting...")

	isRegistered := false

	for {
		var msg model.IncomingMessage
		if err := ws.ReadJSON(&msg); err != nil {
			break
		}

		if msg.ID != "" && !isRegistered {
			if msg.Name == "Main Storage" {
				service.SetMainDeviceID(msg.ID)
			}
			currentDeviceID = msg.ID
			s := store.Global
			s.Mutex.Lock()
			s.DeviceConns[currentDeviceID] = ws
			s.Mutex.Unlock()
			isRegistered = true // 标记已注册，下次跳过
			log.Printf("[MC] Device Registered: %s", currentDeviceID)
		}

		clientVersion := parseWhitelistVersion(msg.WhitelistVersion)
		items, serverVersion := service.GetWhitelistSnapshot()
		if serverVersion != "" && clientVersion != serverVersion {
			_ = ws.WriteJSON(gin.H{
				"type":    "config_sync",
				"data":    items,
				"version": serverVersion,
			})
		}

		switch msg.Type {
		case "update":
			service.ProcessInventoryUpdate(msg.ID, msg.Data)
		case "production_flow":
			service.ProcessFlowUpdate(msg)
		case "craftables":
			service.ProcessCraftablesUpdate(msg.ID, msg.Craftables)
		}
	}

	if currentDeviceID != "" {
		s := store.Global
		s.Mutex.Lock()
		delete(s.DeviceConns, currentDeviceID)
		s.Mutex.Unlock()
		log.Printf("[MC] Device [%s] Disconnected", currentDeviceID)
	}
}

func HandleWeb(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	s := store.Global

	// 注册 Web 客户端
	s.Mutex.Lock()
	s.WebClients[ws] = true
	s.Mutex.Unlock()

	service.BroadcastToWeb()

	for {
		var cmd model.Command
		if err := ws.ReadJSON(&cmd); err != nil {
			s.Mutex.Lock()
			delete(s.WebClients, ws)
			s.Mutex.Unlock()
			break
		}

		// ==========================================
		// 1. 处理状态更新 (注意：这里不要加全局锁！)
		// ==========================================
		switch cmd.Action {
		case "stop":
			// 关机：调用 Service 层的重置函数
			// 注意：ResetFactoryStats 内部自己会加锁，所以这里绝对不能加 s.Mutex.Lock()，否则死锁！
			service.ResetFactoryStats(cmd.Target)

		case "start":
			// 开机：简单的状态更新，需要手动加锁
			s.Mutex.Lock()
			if factory, exists := s.Factories[cmd.Target]; exists {
				factory.IsActive = true
			}
			s.Mutex.Unlock()
			// 广播让前端变绿
			service.BroadcastToWeb()
		case "update_factory_items":
			service.UpdateFactoryItemSettings(cmd.Target, cmd.PrimaryItem, cmd.Items)
		case "update_factory_name":
			service.UpdateFactoryName(cmd.Target, cmd.Name)
		}

		// ==========================================
		// 2. 转发指令给海龟 (需要加锁读取 DeviceConns)
		// ==========================================
		if cmd.Action == "start" || cmd.Action == "stop" {
			s.Mutex.Lock()
			if targetConn, ok := s.DeviceConns[cmd.Target]; ok {
				targetConn.WriteJSON(cmd)
				log.Printf("Command forwarded to [%s]: %s", cmd.Target, cmd.Action)
			} else {
				log.Printf("Target [%s] offline, command dropped.", cmd.Target)
			}
			s.Mutex.Unlock()
		}
	}
}

func HandleIcon(c *gin.Context) {
	fullID := c.Param("id")
	data, err := service.GetIconImage(fullID)
	if err != nil {
		c.Status(404)
		return
	}
	c.Data(200, "image/png", data)
}

func HandleItemName(c *gin.Context) {
	fullID := c.Param("id")
	name, err := service.GetItemDisplayName(fullID)
	if err != nil {
		c.JSON(200, gin.H{"id": fullID, "name": name})
		return
	}
	c.JSON(200, gin.H{"id": fullID, "name": name})
}

// HandleConfig 读取持久化白名单
func HandleConfig(c *gin.Context) {
	items, version, _, err := service.EnsureWhitelistFromFactories()
	if err != nil {
		items, version = service.GetWhitelistSnapshot()
	}
	c.JSON(200, gin.H{"monitored_items": items, "version": version})
}

type whitelistUpdateRequest struct {
	Items []string `json:"items"`
}

type autoCraftTaskPatchRequest struct {
	IsActive bool `json:"isActive"`
}

func HandleConfigUpdate(c *gin.Context) {
	_, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	var req whitelistUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid json"})
		return
	}

	version, err := service.UpdateWhitelist(req.Items)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to save whitelist"})
		return
	}

	updated, _ := service.GetWhitelistSnapshot()
	payload := gin.H{"type": "config_sync", "data": updated, "version": version}

	connections := make([]*websocket.Conn, 0)
	s := store.Global
	s.Mutex.RLock()
	for _, conn := range s.DeviceConns {
		connections = append(connections, conn)
	}
	s.Mutex.RUnlock()

	for _, conn := range connections {
		_ = conn.WriteJSON(payload)
	}
	c.JSON(200, gin.H{"monitored_items": updated, "version": version})
}

func HandleAutoCraftables(c *gin.Context) {
	requestID := fmt.Sprintf("%d", time.Now().UnixNano())
	target := c.Query("target")
	requested := service.RequestCraftablesRefresh(target, requestID)
	items, lastUpdated := service.GetCraftablesSnapshot()
	c.JSON(200, gin.H{
		"items":       items,
		"requested":   requested,
		"lastUpdated": lastUpdated,
	})
}

func HandleAutoCraftRecipe(c *gin.Context) {
	itemID := c.Query("itemId")
	if itemID == "" {
		c.JSON(400, gin.H{"error": "missing itemId"})
		return
	}

	recipe := service.BuildRecipeSnapshot(itemID)
	if recipe == nil {
		c.JSON(404, gin.H{"error": "recipe not found"})
		return
	}

	c.JSON(200, recipe)
}

func HandleAutoCraftTasks(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		c.JSON(200, gin.H{"items": service.ListAutoCraftTasks()})
		return
	}

	var task model.AutoCraftTask
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(400, gin.H{"error": "invalid json"})
		return
	}

	created, err := service.UpsertAutoCraftTask(task)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, created)
}

func HandleAutoCraftTaskDelete(c *gin.Context) {
	itemID := c.Param("itemId")
	if itemID == "" {
		c.JSON(400, gin.H{"error": "missing itemId"})
		return
	}

	if !service.DeleteAutoCraftTask(itemID) {
		c.JSON(404, gin.H{"error": "task not found"})
		return
	}

	c.JSON(200, gin.H{"ok": true})
}

func HandleAutoCraftTaskPatch(c *gin.Context) {
	itemID := c.Param("itemId")
	if itemID == "" {
		c.JSON(400, gin.H{"error": "missing itemId"})
		return
	}

	var req autoCraftTaskPatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid json"})
		return
	}

	task, ok := service.SetAutoCraftTaskActive(itemID, req.IsActive)
	if !ok {
		c.JSON(404, gin.H{"error": "task not found"})
		return
	}

	c.JSON(200, task)
}
