package model

import (
	"encoding/json"
)

type EnergyStats struct {
	EnergyStored       float64 `json:"energyStored"`
	EnergyMax          float64 `json:"energyMax"`
	EnergyUsage        float64 `json:"energyUsage"`
	AverageEnergyInput float64 `json:"averageEnergyInput"`
	NetEnergyRate      float64 `json:"netEnergyRate"`
}

type StorageStats struct {
	// 物品存储 (ME Drives)
	ItemTotal int64 `json:"itemTotal"`
	ItemUsed  int64 `json:"itemUsed"`
	// 计算字段：剩余空间 (Total - Used)
	ItemAvailable int64 `json:"itemAvailable,omitempty"`

	// 外部物品存储 (Storage Bus)
	ItemExternalTotal int64 `json:"itemExternalTotal"`
	ItemExternalUsed  int64 `json:"itemExternalUsed"`
	// 计算字段
	ItemExternalAvailable int64 `json:"itemExternalAvailable,omitempty"`

	// 流体存储
	FluidTotal int64 `json:"fluidTotal"`
	FluidUsed  int64 `json:"fluidUsed"`
	// 计算字段
	FluidAvailable int64 `json:"fluidAvailable,omitempty"`
}

// SystemStats 存储系统运行时的状态（高频变动）
type SystemStats struct {
	LastUpdated int64            `json:"lastUpdated"`
	EnergyStats EnergyStats      `json:"energyStats"`
	Storage     StorageStats     `json:"storage"`
	Inventory   map[string]int64 `json:"inventory,omitempty"`
}

// LuaReport 对应 AE2 发上来的库存快照
type LuaReport struct {
	RawItems map[string]int64 `json:"items"`
	IsActive bool             `json:"active"`
	Name     string           `json:"name"`
	Energy   *EnergyStats     `json:"energy,omitempty"`
	Storage  *StorageStats    `json:"storage,omitempty"`
}

// FactoryData 发给 Vue 前端的最终数据
type FactoryData struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	NameLocked  bool                    `json:"nameLocked"`
	ItemID      string                  `json:"itemId"`      // 卡片图标(兼容旧字段)
	PrimaryItem string                  `json:"primaryItem"` // 主显示物品
	Items       map[string]*FactoryItem `json:"items"`
	IsActive    bool                    `json:"isActive"`
	LastUpdated int64                   `json:"lastUpdated"`
}

type FactoryItem struct {
	ItemID   string  `json:"itemId"`
	Count    int64   `json:"count"`
	ProdRate float64 `json:"prodRate"`
	Visible  bool    `json:"visible"`
	Order    int     `json:"order"`
}

type FactoryItemSetting struct {
	ItemID  string `json:"itemId"`
	Visible bool   `json:"visible"`
	Order   int    `json:"order"`
}

type RecipeSnapshot struct {
	ItemID   string            `json:"itemId"`
	ItemName string            `json:"itemName"`
	Count    int64             `json:"count"`
	Children []*RecipeSnapshot `json:"children,omitempty"`
}

type AutoCraftTask struct {
	ItemID         string          `json:"itemId"`
	ItemName       string          `json:"itemName"`
	MinThreshold   int64           `json:"minThreshold"`
	MaxThreshold   int64           `json:"maxThreshold"`
	IsActive       bool            `json:"isActive"`
	RecipeSnapshot *RecipeSnapshot `json:"recipeSnapshot,omitempty"`
}

type CraftableItem struct {
	ItemID   string `json:"itemId"`
	ItemName string `json:"itemName"`
	Count    int64  `json:"count,omitempty"`
}

type BridgeDebugResult struct {
	Ok        bool                   `json:"ok"`
	Error     string                 `json:"error,omitempty"`
	DataType  string                 `json:"dataType,omitempty"`
	DataValue interface{}            `json:"dataValue,omitempty"`
	Data      interface{}            `json:"data,omitempty"`
	Summary   map[string]interface{} `json:"summary,omitempty"`
}

type BridgeDebugPayload struct {
	Timestamp int64                        `json:"timestamp"`
	Results   map[string]BridgeDebugResult `json:"results"`
}

// IncomingMessage 统一接收 Lua 消息
type IncomingMessage struct {
	Type             string              `json:"type"`
	Data             LuaReport           `json:"data"`
	ID               string              `json:"id"`
	RequestID        string              `json:"requestId"`
	Name             string              `json:"name"`
	Delta            int64               `json:"delta"`
	ItemID           string              `json:"itemId"`
	Craftables       []CraftableItem     `json:"craftables"`
	Debug            *BridgeDebugPayload `json:"debug,omitempty"`
	WhitelistVersion json.RawMessage     `json:"whitelist_version"`
}

// Command 控制指令
type Command struct {
	Target      string               `json:"target"`
	Action      string               `json:"action"`
	Type        string               `json:"type,omitempty"`
	RequestID   string               `json:"requestId,omitempty"`
	ItemID      string               `json:"itemId,omitempty"`
	Count       int64                `json:"count,omitempty"`
	Name        string               `json:"name,omitempty"`
	PrimaryItem string               `json:"primaryItem,omitempty"`
	Items       []FactoryItemSetting `json:"items,omitempty"`
}
