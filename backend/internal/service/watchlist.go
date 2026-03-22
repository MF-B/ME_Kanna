package service

import (
	"ME_Kanna/internal/model"
	"ME_Kanna/internal/store"
	"log"
)

// UpdatePriorityWatchlist 更新 Priority 名单并下发给采集端
func UpdatePriorityWatchlist(items []string) {
	n := store.Network
	n.Lock()
	n.PriorityWatchlist = make(map[string]bool, len(items))
	for _, item := range items {
		n.PriorityWatchlist[item] = true
	}
	n.Unlock()

	pushWatchlist(model.MsgCmdSetPriority, items)
	log.Printf("[Watchlist] priority updated: %d items", len(items))
}

// UpdateRoutineWatchlist 更新 Routine 名单并下发给采集端
func UpdateRoutineWatchlist(items []string) {
	n := store.Network
	n.Lock()
	n.RoutineWatchlist = make([]string, len(items))
	copy(n.RoutineWatchlist, items)
	n.Unlock()

	pushWatchlist(model.MsgCmdSetRoutine, items)
	log.Printf("[Watchlist] routine updated: %d items", len(items))
}

// GetPriorityWatchlist 返回当前 Priority 名单
func GetPriorityWatchlist() []string {
	n := store.Network
	n.RLock()
	defer n.RUnlock()

	items := make([]string, 0, len(n.PriorityWatchlist))
	for item := range n.PriorityWatchlist {
		items = append(items, item)
	}
	return items
}

// GetRoutineWatchlist 返回当前 Routine 名单
func GetRoutineWatchlist() []string {
	n := store.Network
	n.RLock()
	defer n.RUnlock()

	result := make([]string, len(n.RoutineWatchlist))
	copy(result, n.RoutineWatchlist)
	return result
}

// SyncRoutineFromRules 根据当前 AutoCraftRules 自动重建 RoutineWatchlist 并推送
func SyncRoutineFromRules() {
	n := store.Network
	n.RLock()

	items := make([]string, 0, len(n.AutoCraftRules))
	for itemId, rule := range n.AutoCraftRules {
		if rule.IsActive {
			items = append(items, itemId)
		}
	}

	n.RUnlock()

	UpdateRoutineWatchlist(items)
}

// PushWatchlistsToCollector 向指定采集端推送当前名单 (用于新连接注册时)
func PushWatchlistsToCollector(conn *store.SafeConn) {
	priority := GetPriorityWatchlist()
	routine := GetRoutineWatchlist()

	if len(priority) > 0 {
		pushWatchlistTo(conn, model.MsgCmdSetPriority, priority)
	}
	if len(routine) > 0 {
		pushWatchlistTo(conn, model.MsgCmdSetRoutine, routine)
	}
}

// pushWatchlist 向所有采集端广播名单更新
func pushWatchlist(msgType string, items []string) {
	dataBytes, err := jsonMarshal(model.WatchlistData{Items: items})
	if err != nil {
		log.Printf("[Watchlist] marshal error: %v", err)
		return
	}

	msg := model.Envelope{
		Type: msgType,
		Data: dataBytes,
	}

	store.Pool.BroadcastToCollectors(msg)
}

// pushWatchlistTo 向指定连接推送名单
func pushWatchlistTo(conn *store.SafeConn, msgType string, items []string) {
	dataBytes, err := jsonMarshal(model.WatchlistData{Items: items})
	if err != nil {
		return
	}

	msg := model.Envelope{
		Type: msgType,
		Data: dataBytes,
	}
	_ = conn.WriteJSON(msg)
}
