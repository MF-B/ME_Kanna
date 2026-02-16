package service

import (
	"encoding/json"
	"fmt"
	"log"
	"mineCCT/internal/model"
	"mineCCT/internal/store"
	"strings"
	"sync"
	"time"
)

// ========== Pattern 数据类型 ==========

type PatternItem struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Count       int64  `json:"count"`
	Fingerprint string `json:"fingerprint,omitempty"`
}

type Pattern struct {
	PatternType   string        `json:"patternType"`
	PrimaryOutput PatternItem   `json:"primaryOutput"`
	Inputs        []PatternItem `json:"inputs"`
	Outputs       []PatternItem `json:"outputs"`
}

// ========== Pattern 缓存 ==========

var patternState = struct {
	mu          sync.RWMutex
	patterns    []Pattern
	lastUpdated time.Time
	deviceID    string
}{}

func ProcessPatternsUpdate(deviceID string, rawPatterns json.RawMessage) {
	var patterns []Pattern
	if err := json.Unmarshal(rawPatterns, &patterns); err != nil {
		log.Printf("[Patterns] parse error: %v", err)
		return
	}

	patternState.mu.Lock()
	patternState.patterns = patterns
	patternState.lastUpdated = time.Now()
	if strings.TrimSpace(deviceID) != "" {
		patternState.deviceID = deviceID
	}
	patternState.mu.Unlock()

	log.Printf("[Patterns] cached %d patterns from device %s", len(patterns), deviceID)
}

func RequestPatternsRefresh(targetID, requestID string, filter map[string]interface{}) bool {
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

	cmd := map[string]interface{}{
		"type":      "cmd_get_patterns",
		"requestId": requestID,
	}
	if filter != nil {
		cmd["filter"] = filter
	}

	_ = targetConn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	err := targetConn.WriteJSON(cmd)
	_ = targetConn.SetWriteDeadline(time.Time{})
	return err == nil
}

func GetPatternsSnapshot() ([]Pattern, time.Time) {
	patternState.mu.RLock()
	defer patternState.mu.RUnlock()

	copied := make([]Pattern, len(patternState.patterns))
	copy(copied, patternState.patterns)
	return copied, patternState.lastUpdated
}

// BuildRecipeTree 构建某物品的配方树（递归展开子配方）
func BuildRecipeTree(itemID string, depth int) *model.RecipeSnapshot {
	itemID = strings.TrimSpace(itemID)
	if itemID == "" || depth > 10 {
		return nil
	}

	name, err := GetItemDisplayName(itemID)
	if err != nil || strings.TrimSpace(name) == "" {
		name = itemID
	}

	node := &model.RecipeSnapshot{
		ItemID:   itemID,
		ItemName: name,
		Count:    1,
	}

	// 在缓存的 patterns 中查找以此物品为主产出的配方
	patternState.mu.RLock()
	patterns := patternState.patterns
	patternState.mu.RUnlock()

	var matched *Pattern
	for i := range patterns {
		if patterns[i].PrimaryOutput.Name == itemID {
			matched = &patterns[i]
			break
		}
	}

	if matched == nil {
		return node // 没有配方，是原材料
	}

	node.Count = matched.PrimaryOutput.Count

	// 递归展开子配方
	for _, input := range matched.Inputs {
		child := BuildRecipeTree(input.Name, depth+1)
		if child != nil {
			child.Count = input.Count
			node.Children = append(node.Children, child)
		}
	}

	return node
}

// HandlePatternsRequest 处理来自前端的配方请求：先刷新再返回
func HandlePatternsRequest(itemID string) ([]Pattern, bool) {
	requestID := fmt.Sprintf("%d", time.Now().UnixNano())

	var filter map[string]interface{}
	if itemID != "" {
		filter = map[string]interface{}{
			"output": map[string]interface{}{
				"name": itemID,
			},
		}
	}

	requested := RequestPatternsRefresh("", requestID, filter)
	patterns, _ := GetPatternsSnapshot()
	return patterns, requested
}
