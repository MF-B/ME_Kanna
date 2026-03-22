package service

import (
	"ME_Kanna/internal/model"
	"ME_Kanna/internal/store"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"sync"
	"time"
)

// jsonMarshal 是 json.Marshal 的包级别辅助函数 (供 broadcast.go 等使用)
func jsonMarshal(v interface{}) (json.RawMessage, error) {
	return json.Marshal(v)
}

// ========================
// 合成事件处理
// ========================

// ProcessCraftingEvent 处理采集端上报的 evt_crafting 事件
func ProcessCraftingEvent(data *model.CraftingEventData) {
	n := store.Network
	n.Lock()

	job, exists := n.ActiveJobs[data.JobID]
	if exists {
		if data.IsError {
			job.Status = model.JobStatusError
		} else {
			// 根据消息内容判断状态
			msg := strings.ToUpper(data.Message)
			if strings.Contains(msg, "CRAFTING_STARTED") || strings.Contains(msg, "CRAFTING") {
				job.Status = model.JobStatusCrafting
			}
			// 完成时移除任务
			if strings.Contains(msg, "FINISHED") || strings.Contains(msg, "COMPLETED") {
				delete(n.ActiveJobs, data.JobID)
			}
		}
	}

	n.Unlock()

	// 转发给前端
	BroadcastCraftingEvent(data)

	log.Printf("[CraftEvent] jobId=%d isError=%v msg=%s", data.JobID, data.IsError, data.Message)
}

// ========================
// craftingStatus 五态聚合
// ========================

// ComputeCraftingStatus 计算 Priority 名单中有 AutoCraftRule 的物品的合成状态
// 调用方需持有 n.RLock()
func ComputeCraftingStatus(n *store.AE2Network) map[string]string {
	result := make(map[string]string)

	for itemId, rule := range n.AutoCraftRules {
		// 只处理 Priority 名单中的物品
		if !n.PriorityWatchlist[itemId] {
			continue
		}

		if !rule.IsActive {
			result[itemId] = model.CraftingStatusDisabled
			continue
		}

		// 检查是否有活跃任务
		hasJob := false
		jobStatus := ""
		for _, job := range n.ActiveJobs {
			if job.ItemId == itemId {
				hasJob = true
				jobStatus = job.Status
				break
			}
		}

		if hasJob {
			switch jobStatus {
			case model.JobStatusError:
				result[itemId] = model.CraftingStatusError
			case model.JobStatusCrafting:
				result[itemId] = model.CraftingStatusCrafting
			default:
				result[itemId] = model.CraftingStatusPending
			}
			continue
		}

		// 无活跃任务，检查库存
		state, exists := n.Inventory[itemId]
		if exists && state.Count < rule.MinThreshold {
			result[itemId] = model.CraftingStatusPending
		} else {
			result[itemId] = model.CraftingStatusIdle
		}
	}

	return result
}

// ========================
// 自动合成引擎
// ========================

var autocraftMu sync.Mutex
var autocraftCooldown = make(map[string]int64)

const autocraftCooldownSec int64 = 8

// EvaluateAutocraft 遍历 AutoCraftRules, 对库存不足的物品触发合成
func EvaluateAutocraft() {
	n := store.Network
	n.RLock()

	now := time.Now().Unix()
	var commands []model.CraftItemData

	for itemId, rule := range n.AutoCraftRules {
		if !rule.IsActive {
			continue
		}

		state, exists := n.Inventory[itemId]
		if !exists || state.Count >= rule.MinThreshold {
			continue
		}

		// 检查冷却
		autocraftMu.Lock()
		lastTriggered := autocraftCooldown[itemId]
		if now-lastTriggered < autocraftCooldownSec {
			autocraftMu.Unlock()
			continue
		}
		autocraftCooldown[itemId] = now
		autocraftMu.Unlock()

		amount := rule.MaxThreshold - state.Count
		if amount <= 0 {
			continue
		}

		commands = append(commands, model.CraftItemData{
			Item:   itemId,
			Amount: amount,
		})
	}

	n.RUnlock()

	// 在锁外下发指令
	for _, cmd := range commands {
		if err := DispatchCraftCommand(cmd); err != nil {
			log.Printf("[Autocraft] dispatch failed: %v", err)
		}
	}
}

// DispatchCraftCommand 向采集端下发合成指令
func DispatchCraftCommand(cmd model.CraftItemData) error {
	dataBytes, err := jsonMarshal(cmd)
	if err != nil {
		return err
	}

	msg := model.Envelope{
		Type: model.MsgCmdCraftItem,
		Data: dataBytes,
	}

	conn := store.Pool.GetAnyCollector()
	if conn == nil {
		return errors.New("no collector connected")
	}

	log.Printf("[Craft] dispatching: %s x%d", cmd.Item, cmd.Amount)
	return conn.WriteJSON(msg)
}

// ========================
// AutoCraftRule CRUD
// ========================

// ListAutoCraftRules 返回所有自动合成规则
func ListAutoCraftRules() []*model.AutoCraftRule {
	n := store.Network
	n.RLock()
	defer n.RUnlock()

	rules := make([]*model.AutoCraftRule, 0, len(n.AutoCraftRules))
	for _, rule := range n.AutoCraftRules {
		copied := *rule
		rules = append(rules, &copied)
	}
	return rules
}

// CreateAutoCraftRule 创建自动合成规则
func CreateAutoCraftRule(rule model.AutoCraftRule) (*model.AutoCraftRule, error) {
	if strings.TrimSpace(rule.ItemId) == "" {
		return nil, errors.New("itemId is required")
	}
	if rule.MinThreshold <= 0 {
		return nil, errors.New("minThreshold must be > 0")
	}
	if rule.MaxThreshold < rule.MinThreshold {
		return nil, errors.New("maxThreshold must be >= minThreshold")
	}

	n := store.Network
	n.Lock()
	defer n.Unlock()

	newRule := &model.AutoCraftRule{
		ItemId:       strings.TrimSpace(rule.ItemId),
		MinThreshold: rule.MinThreshold,
		MaxThreshold: rule.MaxThreshold,
		IsActive:     rule.IsActive,
	}
	n.AutoCraftRules[newRule.ItemId] = newRule

	copied := *newRule

	// 同步 RoutineWatchlist (在锁外执行)
	go SyncRoutineFromRules()

	return &copied, nil
}

// DeleteAutoCraftRule 删除自动合成规则
func DeleteAutoCraftRule(itemId string) bool {
	n := store.Network
	n.Lock()
	defer n.Unlock()

	if _, exists := n.AutoCraftRules[itemId]; !exists {
		return false
	}
	delete(n.AutoCraftRules, itemId)

	autocraftMu.Lock()
	delete(autocraftCooldown, itemId)
	autocraftMu.Unlock()

	// 同步 RoutineWatchlist
	go SyncRoutineFromRules()

	return true
}

// PatchAutoCraftRule 部分更新自动合成规则
func PatchAutoCraftRule(itemId string, patch map[string]interface{}) (*model.AutoCraftRule, bool) {
	n := store.Network
	n.Lock()
	defer n.Unlock()

	rule, exists := n.AutoCraftRules[itemId]
	if !exists {
		return nil, false
	}

	if v, ok := patch["isActive"]; ok {
		if b, ok := v.(bool); ok {
			rule.IsActive = b
		}
	}
	if v, ok := patch["minThreshold"]; ok {
		if f, ok := v.(float64); ok {
			rule.MinThreshold = int64(f)
		}
	}
	if v, ok := patch["maxThreshold"]; ok {
		if f, ok := v.(float64); ok {
			rule.MaxThreshold = int64(f)
		}
	}

	copied := *rule

	// 同步 RoutineWatchlist (isActive 可能变化)
	go SyncRoutineFromRules()

	return &copied, true
}
