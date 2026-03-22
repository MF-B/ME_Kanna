package service

import (
	"ME_Kanna/internal/model"
	"ME_Kanna/internal/store"
	"log"
)

// ProcessTick 处理采集端上报的 evt_tick 数据
func ProcessTick(data *model.TickData) {
	n := store.Network
	n.Lock()
	defer n.Unlock()

	// 更新物品库存
	for _, item := range data.Items {
		n.Inventory[item.Name] = model.ItemState{Count: item.Count}
	}

	// 更新 CPU 状态
	n.CPUs = make([]model.CPUState, len(data.CPUs))
	for i, cpu := range data.CPUs {
		var job *model.CraftingJob
		if cpu.CraftingJob != nil {
			job = &model.CraftingJob{
				Name:  cpu.CraftingJob.Name,
				Count: cpu.CraftingJob.Count,
			}
		}
		n.CPUs[i] = model.CPUState{
			Storage:      cpu.Storage,
			CoProcessors: cpu.CoProcessors,
			Job:          job,
		}
	}

	// 更新能源
	n.Energy = data.Energy

	// 更新存储
	n.Storage = data.Storage

	log.Printf("[Tick] items=%d cpus=%d", len(data.Items), len(data.CPUs))

	// 触发自动合成评估 (在锁外执行)
	go EvaluateAutocraft()

	// 触发广播 (在锁外执行)
	go BroadcastSync()
}
