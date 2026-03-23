package service

import (
	"ME_Kanna/internal/model"
	"ME_Kanna/internal/store"
	"log"
)

// ProcessCraftables 处理采集端上报的 evt_craftables 数据
func ProcessCraftables(data *model.CraftablesData) {
	n := store.Network
	n.Lock()
	n.Craftables = make(map[string]int64, len(data.Items))
	for _, item := range data.Items {
		n.Craftables[item.Name] = item.Count
	}
	n.Unlock()

	log.Printf("[Craftables] updated: %d items", len(data.Items))
}

// TriggerCraftablesScan 向所有采集端下发扫描指令
func TriggerCraftablesScan() {
	msg := model.Envelope{
		Type: model.MsgCmdScanCraftables,
		Data: []byte("{}"),
	}
	store.Pool.BroadcastToCollectors(msg)
	log.Printf("[Craftables] scan command broadcasted")
}

// GetCraftables 获取当前缓存的可合成物品列表
func GetCraftables() []model.CraftableItem {
	n := store.Network
	n.RLock()
	defer n.RUnlock()

	result := make([]model.CraftableItem, 0, len(n.Craftables))
	for name, count := range n.Craftables {
		result = append(result, model.CraftableItem{
			Name:  name,
			Count: count,
		})
	}
	return result
}
