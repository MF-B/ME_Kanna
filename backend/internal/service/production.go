package service

import (
	"ME_Kanna/internal/model"
	"ME_Kanna/internal/store"
	"log"
	"time"
)

// ProcessProduction 处理海龟上报的 evt_production 数据
func ProcessProduction(data *model.ProductionData) {
	n := store.Network
	n.Lock()
	defer n.Unlock()

	factory, exists := n.Factories[data.FactoryId]
	if !exists {
		factory = &model.FactoryState{
			FactoryId:   data.FactoryId,
			FactoryName: data.FactoryName,
			Items:       make(map[string]float64),
			IsActive:    true,
		}
		n.Factories[data.FactoryId] = factory
	}

	factory.FactoryName = data.FactoryName
	factory.IsActive = true
	factory.LastUpdated = time.Now().Unix()

	// 简单速率估算: delta 作为即时速率 (后续可以改为滑动窗口)
	// 暂时直接用 delta 值作为每次上报的产量, 实际速率计算留给后续迭代
	factory.Items[data.ItemId] = float64(data.Delta)

	log.Printf("[Production] factory=%s item=%s delta=%d", data.FactoryId, data.ItemId, data.Delta)
}
