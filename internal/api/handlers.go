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

	for {
		var msg model.IncomingMessage
		if err := ws.ReadJSON(&msg); err != nil {
			break
		}

		if msg.ID != "" {
			currentDeviceID = msg.ID
			s := store.Global
			s.Mutex.Lock()
			s.DeviceConns[currentDeviceID] = ws
			s.Mutex.Unlock()
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

		if msg.Type == "update" {
			service.ProcessInventoryUpdate(msg.Data)
		} else if msg.Type == "production_flow" {
			service.ProcessFlowUpdate(msg)
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
		if cmd.Action == "stop" {
			// 关机：调用 Service 层的重置函数
			// 注意：ResetFactoryStats 内部自己会加锁，所以这里绝对不能加 s.Mutex.Lock()，否则死锁！
			service.ResetFactoryStats(cmd.Target)

		} else if cmd.Action == "start" {
			// 开机：简单的状态更新，需要手动加锁
			s.Mutex.Lock()
			if factory, exists := s.Factories[cmd.Target]; exists {
				factory.IsActive = true
			}
			s.Mutex.Unlock()
			// 广播让前端变绿
			service.BroadcastToWeb()
		} else if cmd.Action == "update_factory_items" {
			service.UpdateFactoryItemSettings(cmd.Target, cmd.PrimaryItem, cmd.Items)
		} else if cmd.Action == "update_factory_name" {
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
	MonitoredItems []string `json:"monitored_items"`
	Items          []string `json:"items"`
}

func HandleConfigUpdate(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	var req whitelistUpdateRequest
	_ = json.Unmarshal(body, &req)

	items := req.MonitoredItems
	if len(items) == 0 {
		items = req.Items
	}
	if len(items) == 0 {
		var list []string
		if err := json.Unmarshal(body, &list); err == nil {
			items = list
		}
	}

	version, err := service.UpdateWhitelist(items)
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
