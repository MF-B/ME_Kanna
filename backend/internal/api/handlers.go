package api

import (
	"ME_Kanna/internal/model"
	"ME_Kanna/internal/service"
	"ME_Kanna/internal/store"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// ========================
// WebSocket: 采集端
// ========================

// HandleCollector 处理采集端 WebSocket 连接
func HandleCollector(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[Collector] upgrade failed: %v", err)
		return
	}
	ws := store.WrapConn(conn)
	defer ws.Close()

	// 从查询参数获取 deviceId
	deviceId := c.Query("deviceId")
	if deviceId == "" {
		deviceId = "default"
	}

	// 注册连接
	store.Pool.AddCollector(deviceId, ws)
	defer store.Pool.RemoveCollector(deviceId)
	log.Printf("[Collector] connected: %s", deviceId)

	// 推送当前名单
	service.PushWatchlistsToCollector(ws)

	// 要求采集端上报初始的可合成列表
	scanCmd := model.Envelope{
		Type: model.MsgCmdScanCraftables,
		Data: []byte("{}"),
	}
	_ = ws.WriteJSON(scanCmd)
	log.Printf("[Collector] sent initial scan command")

	// 读取消息循环
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[Collector] disconnected: %s (%v)", deviceId, err)
			break
		}

		var envelope model.Envelope
		if err := json.Unmarshal(raw, &envelope); err != nil {
			log.Printf("[Collector] invalid message: %v", err)
			continue
		}

		handleCollectorMessage(&envelope)
	}
}

// handleCollectorMessage 按 type 分发采集端消息
func handleCollectorMessage(env *model.Envelope) {
	switch env.Type {
	case model.MsgEvtTick:
		var data model.TickData
		if err := json.Unmarshal(env.Data, &data); err != nil {
			log.Printf("[Collector] evt_tick parse error: %v", err)
			return
		}
		service.ProcessTick(&data)

	case model.MsgEvtCrafting:
		var data model.CraftingEventData
		if err := json.Unmarshal(env.Data, &data); err != nil {
			log.Printf("[Collector] evt_crafting parse error: %v", err)
			return
		}
		service.ProcessCraftingEvent(&data)

	case model.MsgEvtProduction:
		var data model.ProductionData
		if err := json.Unmarshal(env.Data, &data); err != nil {
			log.Printf("[Collector] evt_production parse error: %v", err)
			return
		}
		service.ProcessProduction(&data)

	case model.MsgEvtCraftables:
		var data model.CraftablesData
		if err := json.Unmarshal(env.Data, &data); err != nil {
			log.Printf("[Collector] evt_craftables parse error: %v", err)
			return
		}
		service.ProcessCraftables(&data)

	default:
		log.Printf("[Collector] unknown type: %s", env.Type)
	}
}

// ========================
// WebSocket: 前端
// ========================

// HandleFrontend 处理前端 WebSocket 连接
func HandleFrontend(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[Frontend] upgrade failed: %v", err)
		return
	}
	ws := store.WrapConn(conn)
	defer ws.Close()

	store.Pool.AddFrontend(ws)
	defer store.Pool.RemoveFrontend(ws)
	log.Printf("[Frontend] connected")

	// 立刻推送一次当前状态
	go service.BroadcastSync()

	// 读取循环 (保持连接存活, 处理前端可能发的消息)
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[Frontend] disconnected: %v", err)
			break
		}
		// 前端消息暂不处理, 后续可扩展 (如手动合成请求)
	}
}

// ========================
// HTTP API
// ========================

// HandleItemInfo 获取物品静态信息
func HandleItemInfo(c *gin.Context) {
	itemId := c.Param("itemId")
	if itemId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "itemId is required"})
		return
	}

	info := store.Network.GetItemInfo(itemId)
	c.JSON(http.StatusOK, info)
}

// HandleGetCraftables 获取全网可合成物品列表
func HandleGetCraftables(c *gin.Context) {
	items := service.GetCraftables()
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// HandleTriggerCraftablesScan 触发手动扫描全网可合成物品
func HandleTriggerCraftablesScan(c *gin.Context) {
	service.TriggerCraftablesScan()
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// HandleListAutoCraftTasks 列出所有自动合成规则
func HandleListAutoCraftTasks(c *gin.Context) {
	rules := service.ListAutoCraftRules()
	c.JSON(http.StatusOK, rules)
}

// HandleCreateAutoCraftTask 创建自动合成规则
func HandleCreateAutoCraftTask(c *gin.Context) {
	var rule model.AutoCraftRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	created, err := service.CreateAutoCraftRule(rule)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// HandleDeleteAutoCraftTask 删除自动合成规则
func HandleDeleteAutoCraftTask(c *gin.Context) {
	itemId := c.Param("itemId")
	if service.DeleteAutoCraftRule(itemId) {
		c.JSON(http.StatusOK, gin.H{"deleted": true})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "rule not found"})
	}
}

// HandlePatchAutoCraftTask 部分更新自动合成规则
func HandlePatchAutoCraftTask(c *gin.Context) {
	itemId := c.Param("itemId")

	var patch map[string]interface{}
	if err := c.ShouldBindJSON(&patch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, found := service.PatchAutoCraftRule(itemId, patch)
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "rule not found"})
		return
	}

	c.JSON(http.StatusOK, updated)
}
