package service

import (
	"ME_Kanna/internal/config"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// 内存缓存：读取过的图片存这里，下次直接返回，速度极快
var iconCache = make(map[string][]byte)
var cacheMutex sync.RWMutex
var exportIndex map[string]string
var exportIndexOnce sync.Once

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

	// 3. 去 icon-exports-x32 按需读取
	imgData, err := findImageInExports(modID, itemName)
	if err != nil {
		return nil, err
	}

	// 4. 写缓存 (Cache Write)
	cacheMutex.Lock()
	iconCache[fullID] = imgData
	cacheMutex.Unlock()

	return imgData, nil
}

func findImageInExports(modID, itemName string) ([]byte, error) {
	baseDir := config.IconExportDir()
	primary := filepath.Join(baseDir, toExportFileName(modID, itemName))
	if data, err := os.ReadFile(primary); err == nil {
		return data, nil
	}

	// 兜底：扫描索引一次，支持导出器附带 NBT 后缀等变体文件名
	key := strings.ToLower(modID + "__" + itemName)
	if path, ok := getExportIndex()[key]; ok {
		if data, err := os.ReadFile(path); err == nil {
			return data, nil
		}
	}

	return nil, fmt.Errorf("image not found")
}

func toExportFileName(modID, itemName string) string {
	return strings.ToLower(modID + "__" + itemName + ".png")
}

func getExportIndex() map[string]string {
	exportIndexOnce.Do(func() {
		exportIndex = buildExportIndex()
	})
	return exportIndex
}

func buildExportIndex() map[string]string {
	index := make(map[string]string)
	baseDir := config.IconExportDir()
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return index
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".png") {
			continue
		}
		stem := strings.TrimSuffix(name, filepath.Ext(name))
		stem = strings.ToLower(stem)
		// 去掉导出器附带的属性后缀，例如 __{'mod:energy':250000}
		if i := strings.Index(stem, "__{"); i > 0 {
			stem = stem[:i]
		}
		if _, exists := index[stem]; !exists {
			index[stem] = filepath.Join(baseDir, name)
		}
	}
	return index
}
