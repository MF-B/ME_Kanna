package config

import (
	"os"
	"path/filepath"
)

func MinecraftDir() string {
	if dir := os.Getenv("MINECCT_MINECRAFT_DIR"); dir != "" {
		return dir
	}
	wd, err := os.Getwd()
	if err != nil {
		return ".minecraft"
	}
	return filepath.Join(wd, ".minecraft")
}

func ModsPath() string {
	return filepath.Join(MinecraftDir(), "mods")
}

// 语言缓存直接放到 .minecraft 下
func LangCacheDir() string {
	return filepath.Join(MinecraftDir(), "lang_cache")
}

// 原版资源仓库存放到 .minecraft/vanilla
func VanillaAssetsPath() string {
	return filepath.Join(MinecraftDir(), "vanilla")
}

func VanillaLangPath() string {
	return filepath.Join(VanillaAssetsPath(), "assets", "minecraft", "lang", "zh_cn.json")
}

func VanillaTexturesRoot() string {
	return filepath.Join(VanillaAssetsPath(), "assets", "minecraft", "textures")
}

func VanillaLangCandidates() []string {
	base := filepath.Join(VanillaAssetsPath(), "assets", "minecraft", "lang")
	return []string{
		filepath.Join(base, "zh_cn.json"),
		filepath.Join(base, "zh_ch.json"),
		filepath.Join(base, "en_us.json"),
	}
}

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
