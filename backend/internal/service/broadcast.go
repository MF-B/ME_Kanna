package service

import (
	"ME_Kanna/internal/model"
	"ME_Kanna/internal/store"
	"log"
)

// BroadcastSync 构建 SyncData 并广播给所有前端客户端
func BroadcastSync() {
	n := store.Network
	n.RLock()

	// 构建 items
	items := make([]model.TickItem, 0, len(n.Inventory))
	for name, state := range n.Inventory {
		items = append(items, model.TickItem{
			Name:  name,
			Count: state.Count,
		})
	}

	// 构建 cpus
	cpus := make([]model.TickCPU, len(n.CPUs))
	for i, cpu := range n.CPUs {
		var job *model.CraftingJob
		if cpu.Job != nil {
			j := *cpu.Job
			job = &j
		}
		cpus[i] = model.TickCPU{
			CoProcessors: cpu.CoProcessors,
			Storage:      cpu.Storage,
			CraftingJob:  job,
		}
	}

	// 复制能源和存储
	energy := n.Energy
	storage := n.Storage

	// 计算 craftingStatus
	craftingStatus := ComputeCraftingStatus(n)

	n.RUnlock()

	// 构建推送消息
	syncMsg := model.Envelope{
		Type: model.MsgEvtSync,
	}
	syncData := model.SyncData{
		Items:          items,
		CPUs:           cpus,
		Energy:         energy,
		Storage:        storage,
		CraftingStatus: craftingStatus,
	}

	// 序列化 data 字段
	dataBytes, err := jsonMarshal(syncData)
	if err != nil {
		log.Printf("[Broadcast] marshal error: %v", err)
		return
	}
	syncMsg.Data = dataBytes

	store.Pool.BroadcastToFrontend(syncMsg)
}

// BroadcastCraftingEvent 转发合成事件给前端
func BroadcastCraftingEvent(event *model.CraftingEventData) {
	dataBytes, err := jsonMarshal(event)
	if err != nil {
		return
	}

	msg := model.Envelope{
		Type: model.MsgEvtCrafting,
		Data: dataBytes,
	}

	store.Pool.BroadcastToFrontend(msg)
}
