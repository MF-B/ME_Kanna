package service

import (
	"mineCCT/internal/config"
	"mineCCT/internal/model"
	"mineCCT/internal/store"
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

// ProcessFlowUpdate 处理海龟产能 (仅计算速率)
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
	item, created := ensureFactoryItem(factory, msg.ItemID)

	var collected []string
	if created {
		seen := make(map[string]bool)
		collected = make([]string, 0)
		for _, existing := range s.Factories {
			for itemID := range existing.Items {
				if itemID == "" || seen[itemID] {
					continue
				}
				seen[itemID] = true
				collected = append(collected, itemID)
			}
		}
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
	item.ProdRate = (totalAmount / float64(RateBucketCount)) * 60.0
	if factory.PrimaryItem == "" {
		factory.PrimaryItem = msg.ItemID
	}
	if factory.ItemID == "" {
		factory.ItemID = factory.PrimaryItem
	}
	factory.LastUpdated = now.Unix()

	BroadcastToWeb()
	s.Mutex.Unlock()

	if len(collected) > 0 {
		normalized := normalizeWhitelist(collected)
		newVersion := computeWhitelistHash(normalized)
		_, currentVersion := GetWhitelistSnapshot()
		if newVersion != currentVersion {
			_, _ = UpdateWhitelist(normalized)
		}
	}
}

// ProcessInventoryUpdate 处理 AE 库存 (Hub 分发模式)
func ProcessInventoryUpdate(deviceID string, report model.LuaReport) {
	s := store.Global
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	now := time.Now()

	updateSystemEnergy(now, report, s)
	updateSystemStorage(now, report, s)
	updateFactories(now, report, s)

	BroadcastToWeb()
}

func updateSystemEnergy(now time.Time, report model.LuaReport, s *store.StateManager) {
	if report.Energy == nil {
		return
	}
	if report.Energy.EnergyMax > 0 {
		s.SystemStatus.EnergyStored = report.Energy.EnergyStored
		s.SystemStatus.EnergyMax = report.Energy.EnergyMax
		s.SystemStatus.EnergyUsage = report.Energy.EnergyUsage
		s.SystemStatus.AverageEnergyInput = report.Energy.AverageEnergyInput
		s.SystemStatus.NetEnergyRate = report.Energy.AverageEnergyInput - report.Energy.EnergyUsage
		s.SystemStatus.LastUpdated = now.Unix()
	}
}

func updateSystemStorage(now time.Time, report model.LuaReport, s *store.StateManager) {
	if report.Storage == nil {
		return
	}

	s.SystemStatus.Storage = *report.Storage
	s.SystemStatus.LastUpdated = now.Unix()
}

func updateFactories(now time.Time, report model.LuaReport, s *store.StateManager) {
	globalInventory := report.RawItems
	if globalInventory == nil {
		globalInventory = make(map[string]int64)
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

func BroadcastToWeb() {
	s := store.Global
	list := make([]*model.FactoryData, 0)
	for _, v := range s.Factories {
		list = append(list, v)
	}

	payload := gin.H{
		"type":   "update",
		"data":   list,
		"system": s.SystemStatus,
	}

	for client := range s.WebClients {
		// 忽略错误，简单处理
		_ = client.WriteJSON(payload)
	}
}

func ResetFactoryStats(id string) {
	s := store.Global
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

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

	// 3. 立即广播新状态
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

	BroadcastToWeb()
	s.Mutex.Unlock()

	normalized := normalizeWhitelist(collected)
	newVersion := computeWhitelistHash(normalized)
	_, currentVersion := GetWhitelistSnapshot()
	if newVersion != currentVersion {
		_, _ = UpdateWhitelist(normalized)
	}
}

func UpdateFactoryName(id string, name string) {
	s := store.Global
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	factory, exists := s.Factories[id]
	if !exists {
		return
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}

	factory.Name = name
	factory.NameLocked = true

	BroadcastToWeb()
}
