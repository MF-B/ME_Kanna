package config

// ModsPath 定义模组文件夹路径
const ModsPath = "/home/mf1bzz/mineCCT/mods"

// 这里定义每个工厂的：显示名字、代表图标
var FactoryRegistry = map[string]struct {
	Name string
	Icon string // 代表图标
	Rates map[string]float64
}{
	"iron_farm": {
		Name: "刷铁机",
		Icon: "minecraft:iron_ingot",
		Rates: map[string]float64{
			"minecraft:iron_ingot":           1.0,
			"minecraft:iron_nugget":          1.0 / 9.0,
			"minecraft:iron_block":           9.0,
			"allthecompressed:iron_block_1x": 81.0,
		},
	},
}
