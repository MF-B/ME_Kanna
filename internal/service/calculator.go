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
// 速率计算相关 (滑动窗口)
// ========================================================

type WindowEntry struct {
	Amount   float64
	Duration float64
}

var factoryWindows = make(map[string]map[string][]WindowEntry)
var lastPacketTimes = make(map[string]map[string]time.Time)
const WindowSize = 5

// ========================================================
// 后台任务
// ========================================================

func StartBackgroundTasks() {
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for range ticker.C {
			checkStaleFactories()
		}
	}()
}

// 清理长期无响应的工厂速率
func checkStaleFactories() {
	s := store.Global
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	now := time.Now()
	hasUpdate := false

	for id, factory := range s.Factories {
		itemTimes, ok := lastPacketTimes[id]
		if !ok { continue }

		for itemID, lastTime := range itemTimes {
			// 超过 3 分钟没收到海龟数据，速率归零
			if now.Sub(lastTime).Minutes() > 3 {
				if factory.Items != nil {
					if item, exists := factory.Items[itemID]; exists && item.ProdRate > 0 {
						item.ProdRate = 0
						hasUpdate = true
					}
				}
				if factoryWindows[id] != nil {
					delete(factoryWindows[id], itemID)
				}
				delete(itemTimes, itemID)
			}
		}

		if len(itemTimes) == 0 {
			delete(lastPacketTimes, id)
		}
	}
	if hasUpdate {
		BroadcastToWeb()
	}
}

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
			factory.ItemId = conf.Icon
		}
	}
	if factory.Name == "" {
		factory.Name = factoryID
	}
	if factory.PrimaryItem == "" {
		factory.PrimaryItem = factory.ItemId
	}
	if factory.ItemId == "" {
		factory.ItemId = factory.PrimaryItem
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

func ensureFactoryItem(factory *model.FactoryData, itemID string) *model.FactoryItem {
	if factory.Items == nil {
		factory.Items = make(map[string]*model.FactoryItem)
	}
	item, exists := factory.Items[itemID]
	if !exists {
		item = &model.FactoryItem{ItemId: itemID, Visible: true}
		factory.Items[itemID] = item
		if factory.PrimaryItem == "" {
			factory.PrimaryItem = itemID
			if factory.ItemId == "" {
				factory.ItemId = itemID
			}
		}
	}
	return item
}

// ProcessFlowUpdate 处理海龟产能 (仅计算速率)
func ProcessFlowUpdate(msg model.IncomingMessage) {
	s := store.Global
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	id := msg.ID
	if id == "" || msg.Item == "" { return }

	factory := ensureFactory(s, id, msg.Name)
	factory.IsActive = true
	item := ensureFactoryItem(factory, msg.Item)

	// 1. 计算窗口数据
	now := time.Now()
	if lastPacketTimes[id] == nil {
		lastPacketTimes[id] = make(map[string]time.Time)
	}
	lastTime, seenBefore := lastPacketTimes[id][msg.Item]
	lastPacketTimes[id][msg.Item] = now

	if !seenBefore { return }

	duration := now.Sub(lastTime).Seconds()
	if duration < 1.0 { return } // 忽略过快的数据抖动

	currentAmount := float64(msg.Delta)

	// 2. 更新滑动窗口
	entry := WindowEntry{Amount: currentAmount, Duration: duration}
	if factoryWindows[id] == nil {
		factoryWindows[id] = make(map[string][]WindowEntry)
	}
	window := factoryWindows[id][msg.Item]
	window = append(window, entry)
	if len(window) > WindowSize {
		window = window[1:]
	}
	factoryWindows[id][msg.Item] = window

	// 3. 计算平均速率
	var totalAmount float64
	var totalDuration float64
	for _, e := range window {
		totalAmount += e.Amount
		totalDuration += e.Duration
	}

	if totalDuration > 0 {
		item.ProdRate = (totalAmount / totalDuration) * 3600.0
	}
	if factory.PrimaryItem == "" {
		factory.PrimaryItem = msg.Item
	}
	if factory.ItemId == "" {
		factory.ItemId = factory.PrimaryItem
	}
	factory.LastUpdated = now.Unix()

	BroadcastToWeb()
}

// ProcessInventoryUpdate 处理 AE 库存 (Hub 分发模式)
func ProcessInventoryUpdate(data map[string]model.LuaReport) {
	s := store.Global
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	now := time.Now()

	updateSystemEnergy(now, data, s)
	updateSystemStorage(now, data, s)
	updateFactories(now, data, s)

	BroadcastToWeb()
}

func updateSystemEnergy(now time.Time, data map[string]model.LuaReport, s *store.StateManager) {
	for _, report := range data {
		if report.Energy == nil {
			break
		}
		if report.Energy.EnergyMax > 0 {
			s.SystemStatus.EnergyStored = report.Energy.EnergyStored
			s.SystemStatus.EnergyMax = report.Energy.EnergyMax
			s.SystemStatus.EnergyUsage = report.Energy.EnergyUsage
			s.SystemStatus.AverageEnergyInput = report.Energy.AverageEnergyInput
			s.SystemStatus.NetEnergyRate = report.Energy.AverageEnergyInput - report.Energy.EnergyUsage
			s.SystemStatus.LastUpdated = now.Unix()
		}
		break
	}
}

func updateSystemStorage(now time.Time, data map[string]model.LuaReport, s *store.StateManager) {
	for _, report := range data {
		if report.Storage == nil {
			break
		}

		s.SystemStatus.Storage = *report.Storage
		s.SystemStatus.LastUpdated = now.Unix()
		break
	}
}

func updateFactories(now time.Time, data map[string]model.LuaReport, s *store.StateManager) {
	// 汇总所有来源的库存到临时大仓库
	globalInventory := make(map[string]int64)
	for _, report := range data {
		for item, count := range report.RawItems {
			globalInventory[item] += count
		}
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
			factory.PrimaryItem = factory.ItemId
		}
		if factory.ItemId == "" {
			factory.ItemId = factory.PrimaryItem
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

	// 1. 清除滑动窗口数据 (防止下次开机时计算错误)
	delete(factoryWindows, id)
	delete(lastPacketTimes, id)

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
	defer s.Mutex.Unlock()

	factory, exists := s.Factories[id]
	if !exists {
		return
	}
	if factory.Items == nil {
		factory.Items = make(map[string]*model.FactoryItem)
	}

	for _, setting := range settings {
		item, ok := factory.Items[setting.ItemId]
		if !ok {
			item = &model.FactoryItem{ItemId: setting.ItemId}
			factory.Items[setting.ItemId] = item
		}
		item.Visible = setting.Visible
		item.Order = setting.Order
	}

	if primaryItem != "" {
		factory.PrimaryItem = primaryItem
		if conf, ok := config.FactoryRegistry[id]; ok && conf.Icon != "" {
			factory.ItemId = conf.Icon
		} else {
			factory.ItemId = primaryItem
		}
	}

	BroadcastToWeb()
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
