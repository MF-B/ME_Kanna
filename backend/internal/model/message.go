package model

import "encoding/json"

// ========================
// 消息信封
// ========================

// Envelope 统一消息信封, 所有 WS 消息都使用 {"type": "...", "data": {...}}
type Envelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// 消息类型常量
const (
	// 采集端 → 后端 (上报事件)
	MsgEvtTick       = "evt_tick"
	MsgEvtCrafting   = "evt_crafting"
	MsgEvtProduction = "evt_production"

	// 后端 → 采集端 (下发指令)
	MsgCmdCraftItem   = "cmd_craft_item"
	MsgCmdSetPriority = "cmd_set_priority"
	MsgCmdSetRoutine  = "cmd_set_routine"

	// 后端 → 前端 (聚合推送)
	MsgEvtSync = "evt_sync"
	// evt_crafting 复用同一常量，在两条 WS 上独立使用
)

// ========================
// 采集端 → 后端 消息体
// ========================

// TickItem evt_tick 中的单个物品
type TickItem struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// TickCPU evt_tick 中的 CPU 原始数据
type TickCPU struct {
	CoProcessors int          `json:"coProcessors"`
	Storage      int          `json:"storage"`
	CraftingJob  *CraftingJob `json:"craftingJob"` // null = 空闲
}

// TickData evt_tick 的 data 字段
type TickData struct {
	Items   []TickItem   `json:"items"`
	CPUs    []TickCPU    `json:"cpus"`
	Energy  EnergyStats  `json:"energy"`
	Storage StorageStats `json:"storage"`
}

// CraftingEventData evt_crafting 的 data 字段
type CraftingEventData struct {
	JobID   int    `json:"jobId"`
	IsError bool   `json:"isError"`
	Message string `json:"message"`
}

// ProductionData evt_production 的 data 字段
type ProductionData struct {
	FactoryId   string `json:"factoryId"`
	FactoryName string `json:"factoryName"`
	ItemId      string `json:"itemId"`
	Delta       int64  `json:"delta"`
}

// ========================
// 后端 → 采集端 消息体
// ========================

// CraftItemData cmd_craft_item 的 data 字段
type CraftItemData struct {
	Item   string `json:"item"`
	Amount int64  `json:"amount"`
}

// WatchlistData cmd_set_priority / cmd_set_routine 的 data 字段
type WatchlistData struct {
	Items []string `json:"items"`
}

// ========================
// 后端 → 前端 消息体
// ========================

// SyncData evt_sync 的 data 字段 (聚合后推送给前端)
type SyncData struct {
	Items          []TickItem        `json:"items"`
	CPUs           []TickCPU         `json:"cpus"`
	Energy         EnergyStats       `json:"energy"`
	Storage        StorageStats      `json:"storage"`
	CraftingStatus map[string]string `json:"craftingStatus"` // itemId -> 五态之一
}
