package service

import (
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
}{
	craftables: make(map[string]model.CraftableItem),
}

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
			ItemID:   itemID,
			ItemName: itemName,
		}
	}

	autoCraftState.mu.Lock()
	autoCraftState.craftables = next
	autoCraftState.lastUpdated = time.Now().Unix()
	autoCraftState.mu.Unlock()
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
		if err == nil && strings.TrimSpace(displayName) != "" {
			items[index].ItemName = displayName
		}
	}

	if len(items) > 0 {
		return items, lastUpdated
	}

	// fallback：Lua 还没上报时，用白名单兜底，避免前端报错
	whitelistItems, _ := GetWhitelistSnapshot()
	fallback := make([]model.CraftableItem, 0, len(whitelistItems))
	for _, id := range whitelistItems {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		displayName, err := GetItemDisplayName(trimmed)
		if err != nil || strings.TrimSpace(displayName) == "" {
			displayName = trimmed
		}
		fallback = append(fallback, model.CraftableItem{ItemID: trimmed, ItemName: displayName})
	}
	return fallback, lastUpdated
}

func RequestCraftablesRefresh(targetID, requestID string) bool {
	s := store.Global
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()

	if targetID != "" {
		if conn, ok := s.DeviceConns[targetID]; ok {
			_ = conn.WriteJSON(model.Command{Type: "cmd_craftables", RequestID: requestID})
			return true
		}
		return false
	}

	for _, conn := range s.DeviceConns {
		_ = conn.WriteJSON(model.Command{Type: "cmd_craftables", RequestID: requestID})
		return true
	}

	return false
}

func BuildRecipeSnapshot(itemID string) *model.RecipeSnapshot {
	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return nil
	}
	name, err := GetItemDisplayName(itemID)
	if err != nil || strings.TrimSpace(name) == "" {
		name = itemID
	}
	return &model.RecipeSnapshot{
		ItemID:   itemID,
		ItemName: name,
		Count:    1,
	}
}
