package service

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"mineCCT/internal/model"
	"mineCCT/internal/store"
)

var autoCraftState = struct {
	mu          sync.RWMutex
	craftables  map[string]model.CraftableItem
	lastUpdated int64
	tasks       map[string]*model.AutoCraftTask
	lastTrigger map[string]int64
	deviceID    string
}{
	craftables:  make(map[string]model.CraftableItem),
	tasks:       make(map[string]*model.AutoCraftTask),
	lastTrigger: make(map[string]int64),
}

const autoCraftCooldownSeconds int64 = 8

func ProcessCraftablesUpdate(deviceID string, list []model.CraftableItem) {
	next := make(map[string]model.CraftableItem)

	for _, item := range list {
		itemID := strings.TrimSpace(item.ItemID)
		if itemID == "" {
			continue
		}
		itemName := strings.TrimSpace(item.ItemName)
		if itemName == "" {
			itemName = itemID
		}
		next[itemID] = model.CraftableItem{
			ItemID:      itemID,
			ItemName:    itemName,
			Fingerprint: strings.TrimSpace(item.Fingerprint),
			Count:       item.Count,
		}
	}

	autoCraftState.mu.Lock()
	autoCraftState.craftables = next
	autoCraftState.lastUpdated = time.Now().Unix()
	autoCraftState.mu.Unlock()

	// Fix #10: 统一通过 SetMainDeviceID 设置
	if strings.TrimSpace(deviceID) != "" {
		SetMainDeviceID(strings.TrimSpace(deviceID))
	}

	// Fix #1: 在锁内拷贝客户端列表，解锁后再发送
	go func() {
		items, _ := GetCraftablesSnapshot()
		payload := model.IncomingMessage{
			Type:       "craftables",
			Craftables: items,
		}

		s := store.Global
		s.Mutex.RLock()
		clients := make([]*store.SafeConn, 0, len(s.WebClients))
		for client := range s.WebClients {
			clients = append(clients, client)
		}
		s.Mutex.RUnlock()

		for _, client := range clients {
			_ = client.WriteJSON(payload)
		}
	}()
}

func GetCraftablesSnapshot() ([]model.CraftableItem, int64) {
	autoCraftState.mu.RLock()
	items := make([]model.CraftableItem, 0, len(autoCraftState.craftables))
	lastUpdated := autoCraftState.lastUpdated
	for _, item := range autoCraftState.craftables {
		items = append(items, item)
	}
	autoCraftState.mu.RUnlock()

	sort.Slice(items, func(i, j int) bool {
		return items[i].ItemID < items[j].ItemID
	})

	for index := range items {
		displayName, err := GetItemDisplayName(items[index].ItemID)
		if err == nil && strings.TrimSpace(displayName) != "" && displayName != items[index].ItemID {
			// 语言文件解析成功且不是 fallback 到 itemID，覆盖
			items[index].ItemName = displayName
		}
		// 否则保留 Lua 原始 displayName 作为兜底
	}

	return items, lastUpdated
}

func RequestCraftablesRefresh(targetID, requestID string) bool {
	selectedTarget := strings.TrimSpace(targetID)
	if selectedTarget == "" {
		autoCraftState.mu.RLock()
		selectedTarget = strings.TrimSpace(autoCraftState.deviceID)
		autoCraftState.mu.RUnlock()
	}

	if selectedTarget == "" {
		return false
	}

	s := store.Global
	s.Mutex.RLock()
	targetConn := s.DeviceConns[selectedTarget]
	s.Mutex.RUnlock()

	if targetConn == nil {
		return false
	}

	err := targetConn.WriteJSON(model.Command{Type: "cmd_craftables", RequestID: requestID})
	return err == nil
}

func BuildRecipeSnapshot(itemID string) *model.RecipeSnapshot {
	return BuildRecipeTree(itemID, 0)
}

func ListAutoCraftTasks() []*model.AutoCraftTask {
	autoCraftState.mu.RLock()
	list := make([]*model.AutoCraftTask, 0, len(autoCraftState.tasks))
	for _, task := range autoCraftState.tasks {
		copied := *task
		list = append(list, &copied)
	}
	autoCraftState.mu.RUnlock()

	sort.Slice(list, func(i, j int) bool {
		return list[i].ItemID < list[j].ItemID
	})
	return list
}

func UpsertAutoCraftTask(task model.AutoCraftTask) (*model.AutoCraftTask, error) {
	normalized, err := normalizeAutoCraftTask(task)
	if err != nil {
		return nil, err
	}

	autoCraftState.mu.RLock()
	_, exists := autoCraftState.tasks[normalized.ItemID]
	autoCraftState.mu.RUnlock()

	if !exists {
		if _, _, err := EnsureWhitelistItems([]string{normalized.ItemID}); err != nil {
			return nil, err
		}
	}

	autoCraftState.mu.Lock()
	autoCraftState.tasks[normalized.ItemID] = normalized
	autoCraftState.mu.Unlock()

	copied := *normalized
	return &copied, nil
}

func DeleteAutoCraftTask(itemID string) bool {
	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return false
	}

	autoCraftState.mu.Lock()
	_, exists := autoCraftState.tasks[itemID]
	if exists {
		delete(autoCraftState.tasks, itemID)
		delete(autoCraftState.lastTrigger, itemID)
	}
	autoCraftState.mu.Unlock()
	return exists
}

func SetAutoCraftTaskActive(itemID string, isActive bool) (*model.AutoCraftTask, bool) {
	itemID = strings.TrimSpace(itemID)
	autoCraftState.mu.Lock()
	task, exists := autoCraftState.tasks[itemID]
	if exists {
		task.IsActive = isActive
		if !isActive {
			delete(autoCraftState.lastTrigger, itemID)
		}
	}
	autoCraftState.mu.Unlock()

	if !exists {
		return nil, false
	}
	copied := *task
	return &copied, true
}

func EvaluateAutoCraftTasks(deviceID string, inventory map[string]int64) {
	if len(inventory) == 0 {
		return
	}

	// Fix #10: 统一通过 SetMainDeviceID 设置
	if strings.TrimSpace(deviceID) != "" {
		SetMainDeviceID(strings.TrimSpace(deviceID))
	}

	now := time.Now().Unix()

	autoCraftState.mu.Lock()
	commands := make([]model.Command, 0)
	for _, task := range autoCraftState.tasks {
		if task == nil || !task.IsActive {
			continue
		}
		currentCount := inventory[task.ItemID]
		if currentCount >= task.MinThreshold {
			continue
		}
		lastTriggeredAt := autoCraftState.lastTrigger[task.ItemID]
		if now-lastTriggeredAt < autoCraftCooldownSeconds {
			continue
		}
		craftCount := task.MaxThreshold - currentCount
		if craftCount <= 0 {
			continue
		}
		autoCraftState.lastTrigger[task.ItemID] = now
		commands = append(commands, model.Command{
			Type:   "craft",
			ItemID: task.ItemID,
			Count:  craftCount,
		})
	}

	targetDeviceID := autoCraftState.deviceID
	autoCraftState.mu.Unlock()

	for _, command := range commands {
		_ = dispatchCraftCommand(targetDeviceID, command)
	}
}

func normalizeAutoCraftTask(task model.AutoCraftTask) (*model.AutoCraftTask, error) {
	itemID := strings.TrimSpace(task.ItemID)
	if itemID == "" {
		return nil, errors.New("itemId is required")
	}
	if task.MinThreshold <= 0 {
		return nil, errors.New("minThreshold must be > 0")
	}
	if task.MaxThreshold < task.MinThreshold {
		return nil, errors.New("maxThreshold must be >= minThreshold")
	}

	itemName := strings.TrimSpace(task.ItemName)
	if itemName == "" {
		resolvedName, err := GetItemDisplayName(itemID)
		if err == nil && strings.TrimSpace(resolvedName) != "" {
			itemName = resolvedName
		} else {
			itemName = itemID
		}
	}

	recipe := task.RecipeSnapshot
	if recipe == nil {
		recipe = BuildRecipeSnapshot(itemID)
	}

	return &model.AutoCraftTask{
		ItemID:         itemID,
		ItemName:       itemName,
		MinThreshold:   task.MinThreshold,
		MaxThreshold:   task.MaxThreshold,
		IsActive:       task.IsActive,
		RecipeSnapshot: recipe,
	}, nil
}

func dispatchCraftCommand(targetDeviceID string, command model.Command) error {
	s := store.Global
	s.Mutex.RLock()

	var targetConn *store.SafeConn
	if targetDeviceID != "" {
		if conn, ok := s.DeviceConns[targetDeviceID]; ok {
			targetConn = conn
		}
	}
	if targetConn == nil {
		for _, conn := range s.DeviceConns {
			targetConn = conn
			break
		}
	}
	s.Mutex.RUnlock()

	if targetConn == nil {
		return errors.New("no device connection")
	}

	return targetConn.WriteJSON(command)
}
