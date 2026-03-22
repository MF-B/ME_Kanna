package config

import (
	"os"
	"path/filepath"
)

// MinecraftDir 返回 .minecraft 目录的绝对路径
func MinecraftDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return ".minecraft"
	}
	// 后端运行在 backend/ 目录, .minecraft 在项目根目录
	return filepath.Join(wd, "..", ".minecraft")
}

// ModsPath 返回 mods 目录路径
func ModsPath() string {
	return filepath.Join(MinecraftDir(), "mods")
}

// LangCacheDir 语言缓存目录
func LangCacheDir() string {
	return filepath.Join(MinecraftDir(), "lang_cache")
}

// IconExportDir 图标导出目录
func IconExportDir() string {
	return filepath.Join(MinecraftDir(), "icon-exports-x32")
}

// ResourcePacksDir 资源包目录
func ResourcePacksDir() string {
	return filepath.Join(MinecraftDir(), "resourcepacks")
}

// VanillaAssetsPath 原版资源目录
func VanillaAssetsPath() string {
	return filepath.Join(MinecraftDir(), "vanilla")
}

// VanillaLangPath 原版中文语言文件
func VanillaLangPath() string {
	return filepath.Join(VanillaAssetsPath(), "assets", "minecraft", "lang", "zh_cn.json")
}

// VanillaTexturesRoot 原版纹理根目录
func VanillaTexturesRoot() string {
	return filepath.Join(VanillaAssetsPath(), "assets", "minecraft", "textures")
}

// VanillaLangCandidates 语言文件候选列表
func VanillaLangCandidates() []string {
	base := filepath.Join(VanillaAssetsPath(), "assets", "minecraft", "lang")
	return []string{
		filepath.Join(base, "zh_cn.json"),
		filepath.Join(base, "zh_ch.json"),
		filepath.Join(base, "en_us.json"),
	}
}
