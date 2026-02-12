package api

import (
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

func HandleMinecraft(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil { return }
	defer ws.Close()

	var currentDeviceID string

	log.Println("[MC] New Device Connecting...")

	for {
		var msg model.IncomingMessage
		if err := ws.ReadJSON(&msg); err != nil { break }

		if msg.ID != "" {
			currentDeviceID = msg.ID
			s := store.Global
			s.Mutex.Lock()
			s.DeviceConns[currentDeviceID] = ws
			s.Mutex.Unlock()
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
	if err != nil { return }
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

// HandleConfig 动态生成白名单
func HandleConfig(c *gin.Context) {
	list := make([]string, 0)
	seen := make(map[string]bool)

	s := store.Global
	s.Mutex.RLock()
	for _, factory := range s.Factories {
		for itemID := range factory.Items {
			if !seen[itemID] {
				list = append(list, itemID)
				seen[itemID] = true
			}
		}
	}
	s.Mutex.RUnlock()

	c.JSON(200, gin.H{"monitored_items": list})
}
