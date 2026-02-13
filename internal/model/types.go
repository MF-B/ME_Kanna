package model

import "encoding/json"

type EnergyStats struct {
	EnergyStored       float64 `json:"energyStored"`
	EnergyMax          float64 `json:"energyMax"`
	EnergyUsage        float64 `json:"energyUsage"`
	AverageEnergyInput float64 `json:"averageEnergyInput"`
}

type StorageStats struct {
	ItemTotal                 int64 `json:"itemTotal"`
	ItemUsed                  int64 `json:"itemUsed"`
	ItemAvailable             int64 `json:"itemAvailable"`
	ItemExternalTotal         int64 `json:"itemExternalTotal"`
	ItemExternalUsed          int64 `json:"itemExternalUsed"`
	ItemExternalAvailable     int64 `json:"itemExternalAvailable"`
	FluidTotal                int64 `json:"fluidTotal"`
	FluidUsed                 int64 `json:"fluidUsed"`
	FluidAvailable            int64 `json:"fluidAvailable"`
	FluidExternalTotal        int64 `json:"fluidExternalTotal"`
	FluidExternalUsed         int64 `json:"fluidExternalUsed"`
	FluidExternalAvailable    int64 `json:"fluidExternalAvailable"`
	ChemicalTotal             int64 `json:"chemicalTotal"`
	ChemicalUsed              int64 `json:"chemicalUsed"`
	ChemicalAvailable         int64 `json:"chemicalAvailable"`
	ChemicalExternalTotal     int64 `json:"chemicalExternalTotal"`
	ChemicalExternalUsed      int64 `json:"chemicalExternalUsed"`
	ChemicalExternalAvailable int64 `json:"chemicalExternalAvailable"`
}

// LuaReport 对应 AE2 发上来的库存快照
type LuaReport struct {
	RawItems map[string]int64 `json:"raw_items"`
	IsActive bool             `json:"isActive"`
	Name     string           `json:"name"`
	Energy   *EnergyStats     `json:"energy,omitempty"`
	Storage  *StorageStats    `json:"storage,omitempty"`
}

// FactoryData 发给 Vue 前端的最终数据
type FactoryData struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	NameLocked  bool                    `json:"nameLocked"`
	ItemId      string                  `json:"itemId"`      // 卡片图标(兼容旧字段)
	PrimaryItem string                  `json:"primaryItem"` // 主显示物品
	Items       map[string]*FactoryItem `json:"items"`
	IsActive    bool                    `json:"isActive"`
	LastUpdated int64                   `json:"lastUpdated"`
}

type FactoryItem struct {
	ItemId   string  `json:"itemId"`
	Count    int64   `json:"count"`
	ProdRate float64 `json:"prodRate"`
	Visible  bool    `json:"visible"`
	Order    int     `json:"order"`
}

type FactoryItemSetting struct {
	ItemId  string `json:"itemId"`
	Visible bool   `json:"visible"`
	Order   int    `json:"order"`
}

// IncomingMessage 统一接收 Lua 消息
type IncomingMessage struct {
	Type             string               `json:"type"`
	Data             map[string]LuaReport `json:"data"`  // AE2 数据
	ID               string               `json:"id"`    // 海龟 ID
	Name             string               `json:"name"`  // 海龟名称
	Delta            int64                `json:"delta"` // 海龟 增量
	Item             string               `json:"item"`  // 海龟 物品
	WhitelistVersion json.RawMessage      `json:"whitelist_version"`
}

// Command 控制指令
type Command struct {
	Target      string               `json:"target"`
	Action      string               `json:"action"`
	Name        string               `json:"name,omitempty"`
	PrimaryItem string               `json:"primaryItem,omitempty"`
	Items       []FactoryItemSetting `json:"items,omitempty"`
}
