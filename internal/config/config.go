package config

// ModsPath 定义模组文件夹路径
const ModsPath = "/home/mf1bzz/mineCCT/mods"

type FactoryOverride struct {
	Name        string
	Icon        string // 代表图标
	PrimaryItem string
}

// 可选覆盖：不填也能即插即用
var FactoryRegistry = map[string]FactoryOverride{
	"iron_farm": {
		Name:        "刷铁机",
		Icon:        "minecraft:iron_ingot",
		PrimaryItem: "minecraft:iron_ingot",
	},
}
