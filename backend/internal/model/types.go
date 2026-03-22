package model

// ========================
// 物品
// ========================

// ItemInfo 物品静态信息 (字典, 缓存在内存中, 按需懒加载)
type ItemInfo struct {
	ItemId string `json:"itemId"`
	Name   string `json:"name"`
	Icon   string `json:"icon"`
}

// ItemState 物品动态信息 (随每次 evt_tick 更新)
type ItemState struct {
	Count int64 `json:"count"`
}

// ========================
// CPU & 合成任务
// ========================

// CraftingJob CPU 上的合成任务 (来自 AE2 原始数据)
type CraftingJob struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// CPUState CPU 状态 (来自 AE2 原始数据)
type CPUState struct {
	Storage      int          `json:"storage"`
	CoProcessors int          `json:"coProcessors"`
	Job          *CraftingJob `json:"job"` // nil = 空闲
}

// JobInfo 平台下发的合成任务追踪
type JobInfo struct {
	JobID  int    `json:"jobId"`
	ItemId string `json:"itemId"`
	Count  int    `json:"count"`
	Status string `json:"status"` // pending, crafting, error
}

// Job 状态常量
const (
	JobStatusPending  = "pending"
	JobStatusCrafting = "crafting"
	JobStatusError    = "error"
)

// ========================
// 能源 & 存储
// ========================

// EnergyStats 能源统计
type EnergyStats struct {
	Stored   float64 `json:"stored"`
	Capacity float64 `json:"capacity"`
	Usage    float64 `json:"usage"`
	Input    float64 `json:"input"`
}

// StorageStats 存储统计
type StorageStats struct {
	ItemTotal  int64 `json:"itemTotal"`
	ItemUsed   int64 `json:"itemUsed"`
	FluidTotal int64 `json:"fluidTotal"`
	FluidUsed  int64 `json:"fluidUsed"`
}

// ========================
// 工厂产能
// ========================

// FactoryState 工厂产能状态
type FactoryState struct {
	FactoryId   string             `json:"factoryId"`
	FactoryName string             `json:"factoryName"`
	Items       map[string]float64 `json:"items"` // itemId -> 生产速率 (个/分钟)
	IsActive    bool               `json:"isActive"`
	LastUpdated int64              `json:"lastUpdated"`
}

// ========================
// 自动合成
// ========================

// AutoCraftRule 自动合成规则 (用户配置)
type AutoCraftRule struct {
	ItemId       string `json:"itemId"`
	MinThreshold int64  `json:"minThreshold"` // 低于此值触发合成
	MaxThreshold int64  `json:"maxThreshold"` // 合成到此数量
	IsActive     bool   `json:"isActive"`
}

// CraftingStatus 指示灯五态
const (
	CraftingStatusDisabled = "disabled" // 规则存在但未启用
	CraftingStatusIdle     = "idle"     // 补货已开启，库存充足
	CraftingStatusPending  = "pending"  // 需要补货，等待合成
	CraftingStatusCrafting = "crafting" // AE2 正在合成
	CraftingStatusError    = "error"    // 合成失败
)
