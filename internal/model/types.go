package model

type EnergyStats struct {
	EnergyStored       float64 `json:"energyStored"`
	EnergyMax          float64 `json:"energyMax"`
	EnergyUsage        float64 `json:"energyUsage"`
	AverageEnergyInput float64 `json:"averageEnergyInput"`
}

// LuaReport 对应 AE2 发上来的库存快照
type LuaReport struct {
	RawItems map[string]int64 `json:"raw_items"`
	IsActive bool             `json:"isActive"`
	Name     string           `json:"name"`
	Energy   *EnergyStats     `json:"energy,omitempty"`
}

// FactoryData 发给 Vue 前端的最终数据
type FactoryData struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	ItemId      string  `json:"itemId"`     // 卡片图标
	Count       int64   `json:"count"`      // 库存总量
	IsActive    bool    `json:"isActive"`
	ProdRate    float64 `json:"prodRate"`   // 纯生产率
	LastUpdated int64   `json:"lastUpdated"`
}

// IncomingMessage 统一接收 Lua 消息
type IncomingMessage struct {
	Type  string              `json:"type"`
	Data  map[string]LuaReport `json:"data"` // AE2 数据
	ID    string              `json:"id"`    // 海龟 ID
	Delta int64               `json:"delta"` // 海龟 增量
	Item  string              `json:"item"`  // 海龟 物品
}

// Command 控制指令
type Command struct {
	Target string `json:"target"`
	Action string `json:"action"`
}
