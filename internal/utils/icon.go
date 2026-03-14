package utils

import (
	"log"
	"os"
	"strings"
)

var iconIndex = make(map[string]string)

func InitIconIndex(dirPath string) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		log.Printf("⚠️ [Asset] 无法读取图标目录: %v", err)
		return
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".png") {
			continue
		}

		fileName := file.Name() // 例: industrialforegoing__speed_addon_tier_2__{...}.png

		// 1. 去掉 .png
		base := strings.TrimSuffix(fileName, ".png")

		// 2. 去掉可能存在的 "__{...}" 后缀
		if idx := strings.Index(base, "__{"); idx != -1 {
			base = base[:idx]
		}

		// 3. 把第一个 "__" 换成 ":"
		itemID := strings.Replace(base, "__", ":", 1)

		iconIndex[itemID] = "/icons/" + fileName
	}

	log.Printf("[Asset Registry] 成功建立静态图标索引，共扫描加载 %d 个物品！", len(iconIndex))
}

func GetIconURL(itemID string) string {
	if url, exists := iconIndex[itemID]; exists {
		return url
	}
	return "/icons/unknown.png"
}
