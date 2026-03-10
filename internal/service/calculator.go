package service

import (
	"ME_Kanna/internal/config"
	"ME_Kanna/internal/model"
	"ME_Kanna/internal/store"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ========================================================
// 速率计算相关 (10 分钟环形桶)
// ========================================================

type RateBucket struct {
	Minute int64
	Amount float64
}

var factoryBuckets = make(map[string]map[string][]RateBucket)

const RateBucketCount = 10

// ========================================================
// 核心处理逻辑
// ========================================================

func SetMainDeviceID(id string) {
	autoCraftState.mu.Lock()
	defer autoCraftState.mu.Unlock()

	// 只有当 ID 变了的时候才打印日志，防止刷屏
	if autoCraftState.deviceID != id {
		autoCraftState.deviceID = id
		log.Printf("[AutoCraft] 主计算机锁定: %s (Main Storage)", id)
	}
}

func applyFactoryOverrides(factoryID string, factory *model.FactoryData) {
	if factory == nil {
		return
	}
	if conf, ok := config.FactoryRegistry[factoryID]; ok {
		if conf.Name != "" && !factory.NameLocked {
			factory.Name = conf.Name
		}
		if conf.PrimaryItem != "" {
			factory.PrimaryItem = conf.PrimaryItem
		}
		if conf.Icon != "" {
			factory.ItemID = conf.Icon
		}
	}
	if factory.Name == "" {
		factory.Name = factoryID
	}
	if factory.PrimaryItem == "" {
		factory.PrimaryItem = factory.ItemID
	}
	if factory.ItemID == "" {
		factory.ItemID = factory.PrimaryItem
	}
}

func ensureFactory(s *store.StateManager, id string, name string) *model.FactoryData {
	factory, exists := s.Factories[id]
	if !exists {
		factory = &model.FactoryData{ID: id, IsActive: true, Items: make(map[string]*model.FactoryItem)}
		s.Factories[id] = factory
	}
	if factory.Items == nil {
		factory.Items = make(map[string]*model.FactoryItem)
	}
	if name != "" && !factory.NameLocked {
		factory.Name = name
	}
	applyFactoryOverrides(id, factory)
	return factory
}

func ensureFactoryItem(factory *model.FactoryData, itemID string) (*model.FactoryItem, bool) {
	if factory.Items == nil {
		factory.Items = make(map[string]*model.FactoryItem)
	}
	item, exists := factory.Items[itemID]
	if !exists {
		item = &model.FactoryItem{ItemID: itemID, Visible: true}
		factory.Items[itemID] = item
		if factory.PrimaryItem == "" {
			factory.PrimaryItem = itemID
			if factory.ItemID == "" {
				factory.ItemID = itemID
			}
		}
		return item, true
	}
	return item, false
}

// collectAllFactoryItemIDs 收集所有工厂中的物品 ID（去重）
// Fix #6: 避免重复的收集逻辑散落在多处
// 调用者必须持有 s.Mutex 锁
func collectAllFactoryItemIDs(s *store.StateManager) []string {
	seen := make(map[string]bool)
	collected := make([]string, 0)
	for _, factory := range s.Factories {
		for itemID := range factory.Items {
			if itemID == "" || seen[itemID] {
				continue
			}
			seen[itemID] = true
			collected = append(collected, itemID)
		}
	}
	return collected
}

// ProcessFlowUpdate 处理海龟产能 (仅计算速率)
// Fix #9: 统一使用 defer Unlock
func ProcessFlowUpdate(msg model.IncomingMessage) {
	s := store.Global
	s.Mutex.Lock()

	id := msg.ID
	if id == "" || msg.ItemID == "" {
		s.Mutex.Unlock()
		return
	}

	factory := ensureFactory(s, id, msg.Name)
	factory.IsActive = true
	_, created := ensureFactoryItem(factory, msg.ItemID)

	var collected []string
	if created {
		collected = collectAllFactoryItemIDs(s)
	}

	// 1. 写入当前分钟桶
	now := time.Now()
	currentMinute := now.Unix() / 60
	if factoryBuckets[id] == nil {
		factoryBuckets[id] = make(map[string][]RateBucket)
	}
	buckets := factoryBuckets[id][msg.ItemID]
	if len(buckets) != RateBucketCount {
		buckets = make([]RateBucket, RateBucketCount)
	}
	index := int(currentMinute % RateBucketCount)
	bucket := &buckets[index]
	if bucket.Minute != currentMinute {
		bucket.Minute = currentMinute
		bucket.Amount = 0
	}
	bucket.Amount += float64(msg.Delta)
	factoryBuckets[id][msg.ItemID] = buckets

	// 2. 计算 10 分钟平均速率
	var totalAmount float64
	minMinute := currentMinute - (RateBucketCount - 1)
	for _, b := range buckets {
		if b.Minute >= minMinute && b.Minute <= currentMinute {
			totalAmount += b.Amount
		}
	}
	item, _ := ensureFactoryItem(factory, msg.ItemID)
	item.ProdRate = (totalAmount / float64(RateBucketCount)) * 60.0
	if factory.PrimaryItem == "" {
		factory.PrimaryItem = msg.ItemID
	}
	if factory.ItemID == "" {
		factory.ItemID = factory.PrimaryItem
	}
	factory.LastUpdated = now.Unix()

	// Fix #1: 先拷贝 payload，解锁后再发送
	s.Mutex.Unlock()

	BroadcastToWeb()

	if len(collected) > 0 {
		_, _, _ = EnsureWhitelistItems(collected)
	}
}

// ProcessInventoryUpdate 处理 AE 库存 (Hub 分发模式)
func ProcessInventoryUpdate(deviceID string, report model.LuaReport) {
	EvaluateAutoCraftTasks(deviceID, report.RawItems)

	s := store.Global
	s.Mutex.Lock()

	now := time.Now()
	networkID := resolveNetworkID(deviceID)
	systemStats := ensureNetworkStats(s, networkID)

	updateSystemEnergy(now, report, systemStats)
	updateSystemStorage(now, report, systemStats)
	updateFactories(now, report, s, systemStats)

	s.Mutex.Unlock()

	BroadcastToWeb()
}

func updateSystemEnergy(now time.Time, report model.LuaReport, systemStats *model.SystemStats) {
	if systemStats == nil || report.Energy == nil {
		return
	}
	if report.Energy.EnergyMax > 0 {
		systemStats.EnergyStats.EnergyStored = report.Energy.EnergyStored
		systemStats.EnergyStats.EnergyMax = report.Energy.EnergyMax
		systemStats.EnergyStats.EnergyUsage = report.Energy.EnergyUsage
		systemStats.EnergyStats.AverageEnergyInput = report.Energy.AverageEnergyInput
		systemStats.EnergyStats.NetEnergyRate = report.Energy.AverageEnergyInput - report.Energy.EnergyUsage
		systemStats.LastUpdated = now.Unix()
	}
}

func updateSystemStorage(now time.Time, report model.LuaReport, systemStats *model.SystemStats) {
	if systemStats == nil || report.Storage == nil {
		return
	}

	systemStats.Storage = *report.Storage
	// 计算 Available = Total - Used
	systemStats.Storage.ItemAvailable = systemStats.Storage.ItemTotal - systemStats.Storage.ItemUsed
	systemStats.Storage.ItemExternalAvailable = systemStats.Storage.ItemExternalTotal - systemStats.Storage.ItemExternalUsed
	systemStats.Storage.FluidAvailable = systemStats.Storage.FluidTotal - systemStats.Storage.FluidUsed
	systemStats.LastUpdated = now.Unix()
}

func updateFactories(now time.Time, report model.LuaReport, s *store.StateManager, systemStats *model.SystemStats) {
	globalInventory := report.RawItems
	if globalInventory == nil {
		globalInventory = make(map[string]int64)
	}

	inventorySnapshot := make(map[string]int64, len(globalInventory))
	for itemID, count := range globalInventory {
		inventorySnapshot[itemID] = count
	}
	if systemStats != nil {
		systemStats.Inventory = inventorySnapshot
	}

	for factoryID, factory := range s.Factories {
		if factory.Items == nil {
			factory.Items = make(map[string]*model.FactoryItem)
		}
		for itemID, item := range factory.Items {
			if count, exists := globalInventory[itemID]; exists {
				item.Count = count
			} else {
				item.Count = 0
			}
		}
		applyFactoryOverrides(factoryID, factory)
		if factory.PrimaryItem == "" {
			factory.PrimaryItem = factory.ItemID
		}
		if factory.ItemID == "" {
			factory.ItemID = factory.PrimaryItem
		}
		factory.LastUpdated = now.Unix()
	}
}

// BroadcastToWeb 构建 payload 并广播到所有 Web 客户端
// Fix #1: 在锁内拷贝数据，解锁后再发送，避免 WriteJSON 阻塞持锁
func BroadcastToWeb() {
	s := store.Global

	// 锁内：拷贝数据
	s.Mutex.RLock()
	list := make([]*model.FactoryData, 0, len(s.Factories))
	for _, v := range s.Factories {
		list = append(list, v)
	}
	systemStats := selectBroadcastSystemStats(s)
	clients := make([]*store.SafeConn, 0, len(s.WebClients))
	for client := range s.WebClients {
		clients = append(clients, client)
	}
	s.Mutex.RUnlock()

	// 锁外：构建 payload 并发送
	payload := gin.H{
		"type":   "update",
		"data":   list,
		"system": toLegacySystemPayload(systemStats),
	}

	for _, client := range clients {
		_ = client.WriteJSON(payload)
	}
}

func BroadcastCraftResult(msg model.IncomingMessage) {
	s := store.Global
	payload := gin.H{
		"type":    "craft_result",
		"itemId":  msg.ItemID,
		"count":   msg.Count,
		"success": msg.Success,
		"taskId":  msg.TaskID,
		"error":   msg.Error,
	}
	log.Printf("[CraftResult] item=%s count=%d success=%v taskId=%s err=%s",
		msg.ItemID, msg.Count, msg.Success, msg.TaskID, msg.Error)

	s.Mutex.RLock()
	clients := make([]*store.SafeConn, 0, len(s.WebClients))
	for client := range s.WebClients {
		clients = append(clients, client)
	}
	s.Mutex.RUnlock()

	for _, client := range clients {
		_ = client.WriteJSON(payload)
	}
}

func BroadcastCraftStatus(msg model.IncomingMessage) {
	s := store.Global
	payload := gin.H{
		"type":    "craft_status",
		"taskId":  msg.TaskID,
		"error":   msg.Error,
		"message": msg.Message,
	}
	log.Printf("[CraftStatus] taskId=%s error=%v msg=%s",
		msg.TaskID, msg.Error, msg.Message)

	s.Mutex.RLock()
	clients := make([]*store.SafeConn, 0, len(s.WebClients))
	for client := range s.WebClients {
		clients = append(clients, client)
	}
	s.Mutex.RUnlock()

	for _, client := range clients {
		_ = client.WriteJSON(payload)
	}
}

func resolveNetworkID(deviceID string) string {
	trimmed := strings.TrimSpace(deviceID)
	if trimmed != "" {
		return trimmed
	}
	return "default"
}

func ensureNetworkStats(s *store.StateManager, networkID string) *model.SystemStats {
	if s.Networks == nil {
		s.Networks = make(map[string]*model.SystemStats)
	}
	stats := s.Networks[networkID]
	if stats == nil {
		stats = &model.SystemStats{
			EnergyStats: model.EnergyStats{EnergyMax: 1},
			Inventory:   make(map[string]int64),
		}
		s.Networks[networkID] = stats
	}
	if stats.Inventory == nil {
		stats.Inventory = make(map[string]int64)
	}
	if stats.EnergyStats.EnergyMax == 0 {
		stats.EnergyStats.EnergyMax = 1
	}
	return stats
}

func selectBroadcastSystemStats(s *store.StateManager) *model.SystemStats {
	autoCraftState.mu.RLock()
	primaryID := strings.TrimSpace(autoCraftState.deviceID)
	autoCraftState.mu.RUnlock()

	if primaryID != "" {
		if stats := s.Networks[primaryID]; stats != nil {
			return stats
		}
	}

	for _, stats := range s.Networks {
		if stats != nil {
			return stats
		}
	}

	return &model.SystemStats{EnergyStats: model.EnergyStats{EnergyMax: 1}, Inventory: map[string]int64{}}
}

// TODO Fix #8: 前端迁移到 system.energyStats.xxx 后删除冗余的扁平字段
func toLegacySystemPayload(stats *model.SystemStats) gin.H {
	if stats == nil {
		stats = &model.SystemStats{EnergyStats: model.EnergyStats{EnergyMax: 1}, Inventory: map[string]int64{}}
	}
	inventory := stats.Inventory
	if inventory == nil {
		inventory = map[string]int64{}
	}

	return gin.H{
		"lastUpdated":        stats.LastUpdated,
		"energyStats":        stats.EnergyStats,
		"energyStored":       stats.EnergyStats.EnergyStored,
		"energyMax":          stats.EnergyStats.EnergyMax,
		"energyUsage":        stats.EnergyStats.EnergyUsage,
		"averageEnergyInput": stats.EnergyStats.AverageEnergyInput,
		"netEnergyRate":      stats.EnergyStats.NetEnergyRate,
		"storage":            stats.Storage,
		"inventory":          inventory,
	}
}

func ResetFactoryStats(id string) {
	s := store.Global
	s.Mutex.Lock()

	// 1. 清除环形桶数据 (防止下次开机时计算错误)
	delete(factoryBuckets, id)

	// 2. 强制归零
	if factory, exists := s.Factories[id]; exists {
		factory.IsActive = false
		for _, item := range factory.Items {
			item.ProdRate = 0
		}
		factory.LastUpdated = time.Now().Unix()
	}

	// Fix #1: 解锁后再广播
	s.Mutex.Unlock()
	BroadcastToWeb()
}

func UpdateFactoryItemSettings(id string, primaryItem string, settings []model.FactoryItemSetting) {
	s := store.Global
	s.Mutex.Lock()

	factory, exists := s.Factories[id]
	if !exists {
		s.Mutex.Unlock()
		return
	}
	if factory.Items == nil {
		factory.Items = make(map[string]*model.FactoryItem)
	}

	for _, setting := range settings {
		item, ok := factory.Items[setting.ItemID]
		if !ok {
			item = &model.FactoryItem{ItemID: setting.ItemID}
			factory.Items[setting.ItemID] = item
		}
		item.Visible = setting.Visible
		item.Order = setting.Order
	}

	if primaryItem != "" {
		factory.PrimaryItem = primaryItem
		if conf, ok := config.FactoryRegistry[id]; ok && conf.Icon != "" {
			factory.ItemID = conf.Icon
		} else {
			factory.ItemID = primaryItem
		}
	}

	collected := collectAllFactoryItemIDs(s)

	// Fix #1: 解锁后再广播
	s.Mutex.Unlock()
	BroadcastToWeb()

	if len(collected) > 0 {
		_, _, _ = EnsureWhitelistItems(collected)
	}
}

func UpdateFactoryName(id string, name string) {
	s := store.Global
	s.Mutex.Lock()

	factory, exists := s.Factories[id]
	if !exists {
		s.Mutex.Unlock()
		return
	}
	name = strings.TrimSpace(name)
	if name == "" {
		s.Mutex.Unlock()
		return
	}

	factory.Name = name
	factory.NameLocked = true

	// Fix #1: 解锁后再广播
	s.Mutex.Unlock()
	BroadcastToWeb()
}
