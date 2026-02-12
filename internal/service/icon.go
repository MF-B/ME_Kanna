package service

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"mineCCT/internal/config"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// 内存缓存：读取过的图片存这里，下次直接返回，速度极快
var iconCache = make(map[string][]byte)
var cacheMutex sync.RWMutex

// GetIconImage 对外接口
func GetIconImage(fullID string) ([]byte, error) {
	// 1. 查缓存 (Cache Hit)
	cacheMutex.RLock()
	if data, ok := iconCache[fullID]; ok {
		cacheMutex.RUnlock()
		return data, nil
	}
	cacheMutex.RUnlock()

	// 2. 解析 ID
	parts := strings.Split(fullID, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid id format")
	}
	modID := parts[0]
	itemName := parts[1]

	// 3. 去 mods 文件夹里找 (Disk Scan)
	imgData, err := findImageInMods(modID, itemName)
	if err != nil {
		return nil, err
	}

	// 4. 写缓存 (Cache Write)
	cacheMutex.Lock()
	iconCache[fullID] = imgData
	cacheMutex.Unlock()

	return imgData, nil
}

// 核心逻辑：遍历 JAR 包找图片
func findImageInMods(modID, itemName string) ([]byte, error) {
	if modID == "minecraft" {
		if data, err := findImageInVanillaRepo(itemName); err == nil {
			return data, nil
		}
	}

	modsPath := config.ModsPath()
	// 读取 mods 目录
	files, err := os.ReadDir(modsPath)
	if err != nil {
		return nil, fmt.Errorf("read mods dir failed: %v", err)
	}

	// 遍历每一个 .jar 文件
	for _, file := range files {
		if file.IsDir() { continue }
		name := strings.ToLower(file.Name())

		if strings.HasSuffix(name, ".jar") {
			zipPath := filepath.Join(modsPath, file.Name())
			r, err := zip.OpenReader(zipPath)
			if err != nil { continue }
			defer r.Close()

			// 可能的路径 (方块、物品、流体)
			possiblePaths := []string{
				"assets/" + modID + "/textures/item/" + itemName + ".png",
				"assets/" + modID + "/textures/block/" + itemName + ".png",
				"assets/" + modID + "/textures/fluid/" + itemName + ".png",
				"assets/" + modID + "/textures/item/" + itemName + "_item.png",
			}

			for _, targetPath := range possiblePaths {
				for _, f := range r.File {
					// 忽略大小写匹配
					if strings.EqualFold(f.Name, targetPath) {
						rc, err := f.Open()
						if err != nil { return nil, err }
						defer rc.Close()

						// 读出来返回
						buf := new(bytes.Buffer)
						_, err = io.Copy(buf, rc)
						return buf.Bytes(), nil
					}
				}
			}
		}
	}
	return nil, fmt.Errorf("image not found")
}

func findImageInVanillaRepo(itemName string) ([]byte, error) {
	base := config.VanillaTexturesRoot()
	possiblePaths := []string{
		filepath.Join(base, "item", itemName+".png"),
		filepath.Join(base, "block", itemName+".png"),
		filepath.Join(base, "fluid", itemName+".png"),
	}
	for _, path := range possiblePaths {
		data, err := os.ReadFile(path)
		if err == nil {
			return data, nil
		}
	}
	return nil, fmt.Errorf("vanilla texture not found")
}
