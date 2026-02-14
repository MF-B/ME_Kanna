package store

import (
	"mineCCT/internal/model"
	"sync"

	"github.com/gorilla/websocket"
)

// SystemMeta 保持不变，用于存角色
type SystemMeta struct {
	DeviceRoles map[string]string
}

// StateManager
type StateManager struct {
	Mutex sync.RWMutex

	// 连接池
	DeviceConns map[string]*websocket.Conn
	WebClients  map[*websocket.Conn]bool

	// 工厂生产数据 (流速)
	Factories map[string]*model.FactoryData

	Networks map[string]*model.SystemStats

	// 元数据
	Meta SystemMeta
}

// Global 初始化
var Global = &StateManager{
	Factories:   make(map[string]*model.FactoryData),
	DeviceConns: make(map[string]*websocket.Conn),
	WebClients:  make(map[*websocket.Conn]bool),

	Networks: make(map[string]*model.SystemStats),

	Meta: SystemMeta{
		DeviceRoles: make(map[string]string),
	},
}

// =======================
// Helper 方法更新
// =======================

// UpdateNetworkStatus 更新指定主设备的状态
func (s *StateManager) UpdateNetworkStatus(deviceID string, stats *model.SystemStats) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	// 直接存入对应的 Map Key 中，互不干扰
	s.Networks[deviceID] = stats
}

// GetNetworkStatus 获取指定主设备的状态
func (s *StateManager) GetNetworkStatus(deviceID string) *model.SystemStats {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()

	return s.Networks[deviceID]
}

// GetAllNetworks 获取所有网络状态 (用于前端概览页)
func (s *StateManager) GetAllNetworks() map[string]*model.SystemStats {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()

	// 注意：为了并发安全，最好这里返回一个深拷贝，或者前端只读
	// 简单起见，这里直接返回 map (在读锁保护下读取瞬间是安全的，但后续使用要注意)
	// 生产环境建议 copy 一份 map
	copyMap := make(map[string]*model.SystemStats)
	for k, v := range s.Networks {
		copyMap[k] = v
	}
	return copyMap
}
