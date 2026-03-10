package service

import (
	"ME_Kanna/internal/store"
	"log"
)

// RegisterDevice 处理设备上线逻辑：保存连接、识别主设备
func RegisterDevice(deviceID string, deviceName string, ws *store.SafeConn) {
	if deviceID == "" {
		return
	}

	s := store.Global
	s.Mutex.Lock()
	s.DeviceConns[deviceID] = ws
	s.Mutex.Unlock()

	log.Printf("[Device] 从设备注册: %s (Name: %s)", deviceID, deviceName)

	// 主设备识别逻辑 — 通过统一函数加锁写入
	if deviceName == "Main Storage" {
		SetMainDeviceID(deviceID)
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
