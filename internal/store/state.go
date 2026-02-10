package store

import (
	"mineCCT/internal/model"
	"sync"

	"github.com/gorilla/websocket"
)

type StateManager struct {
	Factories  map[string]*model.FactoryData
	Mutex      sync.RWMutex
	DeviceConns map[string]*websocket.Conn
	WebClients map[*websocket.Conn]bool
}

// Global 全局单例
var Global = &StateManager{
	Factories:  make(map[string]*model.FactoryData),
	DeviceConns: make(map[string]*websocket.Conn),
	WebClients: make(map[*websocket.Conn]bool),
}
