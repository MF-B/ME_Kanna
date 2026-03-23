package store

import (
	"ME_Kanna/internal/model"
	"ME_Kanna/internal/utils"
	"sync"

	"github.com/gorilla/websocket"
)

// ========================
// WebSocket 连接管理
// ========================

// SafeConn 包装 WebSocket 连接，提供并发安全的写操作
type SafeConn struct {
	Conn *websocket.Conn
	mu   sync.Mutex
}

// WriteJSON 线程安全的 JSON 写
func (sc *SafeConn) WriteJSON(v interface{}) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.Conn.WriteJSON(v)
}

// Close 关闭底层连接
func (sc *SafeConn) Close() error {
	return sc.Conn.Close()
}

// WrapConn 将原始 websocket.Conn 包装为 SafeConn
func WrapConn(conn *websocket.Conn) *SafeConn {
	return &SafeConn{Conn: conn}
}

// ConnPool 管理所有 WebSocket 连接
type ConnPool struct {
	mu              sync.RWMutex
	CollectorConns  map[string]*SafeConn // deviceId -> conn
	FrontendClients map[*SafeConn]bool   // 前端连接集合
}

// ========================
// AE2Network 聚合根
// ========================

// AE2Network 系统核心状态, 所有采集数据的内存缓存
type AE2Network struct {
	mu sync.RWMutex

	// 监控名单
	PriorityWatchlist map[string]bool // 前端视口 + 活跃任务, 全量高频查
	RoutineWatchlist  []string        // 补货名单, Lua 分片轮转查

	// 物品
	Registry  map[string]model.ItemInfo  // 静态字典 (懒加载)
	Inventory map[string]model.ItemState // 动态库存

	// CPU & 合成
	CPUs       []model.CPUState       // AE2 CPU 状态快照
	ActiveJobs map[int]*model.JobInfo // 平台活跃任务

	// 系统状态
	Energy  model.EnergyStats
	Storage model.StorageStats

	// 工厂
	Factories map[string]*model.FactoryState

	// 自动合成规则
	AutoCraftRules map[string]*model.AutoCraftRule

	// 可合成物品字典 (按需扫描获取)
	Craftables map[string]int64
}

// ========================
// 全局实例
// ========================

// Global 全局连接池
var Pool = &ConnPool{
	CollectorConns:  make(map[string]*SafeConn),
	FrontendClients: make(map[*SafeConn]bool),
}

// Network 全局 AE2 网络状态
var Network = &AE2Network{
	PriorityWatchlist: make(map[string]bool),
	RoutineWatchlist:  make([]string, 0),
	Registry:          make(map[string]model.ItemInfo),
	Inventory:         make(map[string]model.ItemState),
	CPUs:              make([]model.CPUState, 0),
	ActiveJobs:        make(map[int]*model.JobInfo),
	Factories:         make(map[string]*model.FactoryState),
	AutoCraftRules:    make(map[string]*model.AutoCraftRule),
	Craftables:        make(map[string]int64),
}

// ========================
// AE2Network 方法
// ========================

// RLock 获取读锁
func (n *AE2Network) RLock() { n.mu.RLock() }

// RUnlock 释放读锁
func (n *AE2Network) RUnlock() { n.mu.RUnlock() }

// Lock 获取写锁
func (n *AE2Network) Lock() { n.mu.Lock() }

// Unlock 释放写锁
func (n *AE2Network) Unlock() { n.mu.Unlock() }

// GetItemInfo 获取物品静态信息, 未缓存时自动解析并写入 Registry
func (n *AE2Network) GetItemInfo(itemId string) model.ItemInfo {
	n.mu.RLock()
	info, exists := n.Registry[itemId]
	n.mu.RUnlock()

	if exists {
		return info
	}

	// 写锁 + 双重检查
	n.mu.Lock()
	defer n.mu.Unlock()

	if info, exists = n.Registry[itemId]; exists {
		return info
	}

	name, err := utils.GetItemDisplayName(itemId)
	if err != nil {
		name = itemId
	}
	icon := utils.GetIconURL(itemId)

	newItem := model.ItemInfo{
		ItemId: itemId,
		Name:   name,
		Icon:   icon,
	}
	n.Registry[itemId] = newItem
	return newItem
}

// ========================
// ConnPool 方法
// ========================

// AddCollector 注册采集端连接
func (p *ConnPool) AddCollector(deviceId string, conn *SafeConn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.CollectorConns[deviceId] = conn
}

// RemoveCollector 移除采集端连接
func (p *ConnPool) RemoveCollector(deviceId string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.CollectorConns, deviceId)
}

// GetCollector 获取指定采集端连接
func (p *ConnPool) GetCollector(deviceId string) *SafeConn {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.CollectorConns[deviceId]
}

// GetAnyCollector 获取任意一个采集端连接 (用于下发指令)
func (p *ConnPool) GetAnyCollector() *SafeConn {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, conn := range p.CollectorConns {
		return conn
	}
	return nil
}

// AddFrontend 注册前端连接
func (p *ConnPool) AddFrontend(conn *SafeConn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.FrontendClients[conn] = true
}

// RemoveFrontend 移除前端连接
func (p *ConnPool) RemoveFrontend(conn *SafeConn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.FrontendClients, conn)
}

// BroadcastToFrontend 向所有前端客户端广播消息
func (p *ConnPool) BroadcastToFrontend(v interface{}) {
	p.mu.RLock()
	clients := make([]*SafeConn, 0, len(p.FrontendClients))
	for client := range p.FrontendClients {
		clients = append(clients, client)
	}
	p.mu.RUnlock()

	for _, client := range clients {
		_ = client.WriteJSON(v)
	}
}

// BroadcastToCollectors 向所有采集端广播消息
func (p *ConnPool) BroadcastToCollectors(v interface{}) {
	p.mu.RLock()
	conns := make([]*SafeConn, 0, len(p.CollectorConns))
	for _, conn := range p.CollectorConns {
		conns = append(conns, conn)
	}
	p.mu.RUnlock()

	for _, conn := range conns {
		_ = conn.WriteJSON(v)
	}
}
