package service

import (
	"mineCCT/internal/config"
	"mineCCT/internal/model"
	"mineCCT/internal/store"
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

var factoryWindows = make(map[string][]WindowEntry)
var lastPacketTimes = make(map[string]time.Time)
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
		lastTime, ok := lastPacketTimes[id]
		if !ok { continue }
		
		// 超过 3 分钟没收到海龟数据，速率归零
		if now.Sub(lastTime).Minutes() > 3 {
			if factory.ProdRate > 0 {
				factory.ProdRate = 0
				hasUpdate = true
				delete(factoryWindows, id)
			}
		}
	}
	if hasUpdate {
		BroadcastToWeb()
	}
}

// ========================================================
// 核心处理逻辑
// ========================================================

func getItemMultiplier(factoryID string, itemID string) float64 {
	conf, ok := config.FactoryRegistry[factoryID]
	if !ok { return 1.0 }
	rate, found := conf.Rates[itemID]
	if !found { return 0 }
	return rate
}

// ProcessFlowUpdate 处理海龟产能 (仅计算速率)
func ProcessFlowUpdate(msg model.IncomingMessage) {
	s := store.Global
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	id := msg.ID
	mul := getItemMultiplier(id, msg.Item)
	if mul == 0 { return }

	factory, exists := s.Factories[id]
	if !exists {
		factory = &model.FactoryData{ID: id, Name: id}
		s.Factories[id] = factory
	}

	// 1. 计算窗口数据
	now := time.Now()
	lastTime, seenBefore := lastPacketTimes[id]
	lastPacketTimes[id] = now

	if !seenBefore { return }

	duration := now.Sub(lastTime).Seconds()
	if duration < 1.0 { return } // 忽略过快的数据抖动

	currentAmount := float64(msg.Delta) * mul

	// 2. 更新滑动窗口
	entry := WindowEntry{Amount: currentAmount, Duration: duration}
	window := factoryWindows[id]
	window = append(window, entry)
	if len(window) > WindowSize {
		window = window[1:]
	}
	factoryWindows[id] = window

	// 3. 计算平均速率
	var totalAmount float64
	var totalDuration float64
	for _, e := range window {
		totalAmount += e.Amount
		totalDuration += e.Duration
	}

	if totalDuration > 0 {
		factory.ProdRate = (totalAmount / totalDuration) * 3600.0
	}

	BroadcastToWeb()
}

// ProcessInventoryUpdate 处理 AE 库存 (Hub 分发模式)
func ProcessInventoryUpdate(data map[string]model.LuaReport) {
	s := store.Global
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	now := time.Now()

	updateSystemEnergy(now, data, s)
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

func updateFactories(now time.Time, data map[string]model.LuaReport, s *store.StateManager) {
	// 汇总所有来源的库存到临时大仓库
	globalInventory := make(map[string]int64)
	for _, report := range data {
		for item, count := range report.RawItems {
			globalInventory[item] += count
		}
	}

	// 根据配置表，主动为每个工厂分配库存
	for factoryID, conf := range config.FactoryRegistry {
		var factoryTotalInv float64 = 0
		for itemID, rate := range conf.Rates {
			if count, exists := globalInventory[itemID]; exists {
				factoryTotalInv += float64(count) * rate
			}
		}

		factory, exists := s.Factories[factoryID]
		if !exists {
			factory = &model.FactoryData{ID: factoryID, IsActive: true}
			s.Factories[factoryID] = factory
		}

		factory.Name = conf.Name
		factory.ItemId = conf.Icon
		factory.Count = int64(factoryTotalInv)
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
		factory.ProdRate = 0
		factory.LastUpdated = time.Now().Unix()
	}

	// 3. 立即广播新状态
	BroadcastToWeb()
}
