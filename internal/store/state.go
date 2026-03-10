package store

import (
	"ME_Kanna/internal/model"
	"sync"

	"github.com/gorilla/websocket"
)

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

// SetWriteDeadline 线程安全的写超时设置
func (sc *SafeConn) SetWriteDeadline(t interface{}) {
	// 由调用方在 WriteJSON 前后自行管理
}

// Close 关闭底层连接
func (sc *SafeConn) Close() error {
	return sc.Conn.Close()
}

// WrapConn 将原始 websocket.Conn 包装为 SafeConn
func WrapConn(conn *websocket.Conn) *SafeConn {
	return &SafeConn{Conn: conn}
}

// SystemMeta 保持不变，用于存角色
type SystemMeta struct {
	DeviceRoles map[string]string
}

// StateManager
type StateManager struct {
	Mutex sync.RWMutex

	// 连接池
	DeviceConns map[string]*SafeConn
	WebClients  map[*SafeConn]bool

	// 工厂生产数据 (流速)
	Factories map[string]*model.FactoryData

	Networks map[string]*model.SystemStats

	// 元数据
	Meta SystemMeta
}

// Global 初始化
var Global = &StateManager{
	Factories:   make(map[string]*model.FactoryData),
	DeviceConns: make(map[string]*SafeConn),
	WebClients:  make(map[*SafeConn]bool),

	Networks: make(map[string]*model.SystemStats),

	Meta: SystemMeta{
		DeviceRoles: make(map[string]string),
	},
}
