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
	SystemStatus SystemStats
}

type SystemStats struct {
	EnergyStored       float64 `json:"energyStored"`
	EnergyMax          float64 `json:"energyMax"`
	EnergyUsage        float64 `json:"energyUsage"`
	AverageEnergyInput float64 `json:"averageEnergyInput"`
	NetEnergyRate      float64 `json:"netEnergyRate"`
	LastUpdated        int64   `json:"lastUpdated"`
}

// Global 全局单例
var Global = &StateManager{
	Factories:  make(map[string]*model.FactoryData),
	DeviceConns: make(map[string]*websocket.Conn),
	WebClients: make(map[*websocket.Conn]bool),
	SystemStatus: SystemStats{EnergyMax: 1},
}
