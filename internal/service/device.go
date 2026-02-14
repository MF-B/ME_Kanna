package service

import (
	"log"
	"mineCCT/internal/store"

	"github.com/gorilla/websocket"
)

// RegisterDevice 处理设备上线逻辑：保存连接、识别主设备
func RegisterDevice(deviceID string, deviceName string, ws *websocket.Conn) {
	if deviceID == "" {
		return
	}

	s := store.Global
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.DeviceConns[deviceID] = ws
	log.Printf("[Device] 从设备注册: %s (Name: %s)", deviceID, deviceName)

	// 主设备识别逻辑
	if deviceName == "Main Storage" {
		autoCraftState.deviceID = deviceID
		log.Printf("[Device] 主设备注册: %s", deviceID)
	}
}

// UnregisterDevice 处理设备下线
func UnregisterDevice(deviceID string) {
	if deviceID == "" {
		return
	}
	s := store.Global
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	delete(s.DeviceConns, deviceID)
	log.Printf("[Device] 下线: %s", deviceID)
}
