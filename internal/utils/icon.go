package utils

import (
	"fmt"
	"strings"
)

// 根据物品ID生成对应的图标URL
func GetIconUrl(fullID string) string {
	parts := strings.Split(fullID, ":")
	if len(parts) != 2 {
		return "/icons/unknown.png"
	}

	modID := strings.ToLower(parts[0])
	itemName := strings.ToLower(parts[1])

	return fmt.Sprintf("/icons/%s__%s.png", modID, itemName)
}
