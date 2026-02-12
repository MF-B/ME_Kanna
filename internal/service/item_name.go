package service

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mineCCT/internal/config"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var langCache = make(map[string]map[string]string)
var langCacheMutex sync.RWMutex

func GetItemDisplayName(fullID string) (string, error) {
	parts := strings.Split(fullID, ":")
	if len(parts) != 2 {
		return fullID, fmt.Errorf("invalid id format")
	}
	modID := parts[0]
	itemName := parts[1]

	langMap, err := getLangMap(modID)
	if err != nil {
		return itemName, err
	}
	if name, ok := lookupItemName(langMap, modID, itemName); ok {
		return name, nil
	}
	return itemName, nil
}

func lookupItemName(langMap map[string]string, modID string, itemName string) (string, bool) {
	if langMap == nil {
		return "", false
	}
	keys := []string{
		"item." + modID + "." + itemName,
		"block." + modID + "." + itemName,
		"fluid." + modID + "." + itemName,
		"entity." + modID + "." + itemName,
	}
	for _, key := range keys {
		if value, ok := langMap[key]; ok {
			return value, true
		}
	}
	return "", false
}

func getLangMap(modID string) (map[string]string, error) {
	langCacheMutex.RLock()
	if cached, ok := langCache[modID]; ok {
		langCacheMutex.RUnlock()
		return cached, nil
	}
	langCacheMutex.RUnlock()

	cachePath := filepath.Join(config.LangCacheDir(), modID+".json")
	if cached, err := readLangFromDisk(cachePath); err == nil {
		storeLangMap(modID, cached)
		return cached, nil
	}

	var data []byte
	var err error
	if modID == "minecraft" {
		data, err = readVanillaLangFromRepo()
	} else {
		data, err = readLangFromModJar(modID)
	}
	if err != nil {
		return nil, err
	}

	langMap := make(map[string]string)
	if err := json.Unmarshal(data, &langMap); err != nil {
		return nil, err
	}

	_ = writeLangToDisk(cachePath, data)
	storeLangMap(modID, langMap)
	return langMap, nil
}

func storeLangMap(modID string, langMap map[string]string) {
	langCacheMutex.Lock()
	langCache[modID] = langMap
	langCacheMutex.Unlock()
}

func ensureLangCacheDir() error {
	return os.MkdirAll(config.LangCacheDir(), 0o755)
}

func readLangFromDisk(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	langMap := make(map[string]string)
	if err := json.Unmarshal(data, &langMap); err != nil {
		return nil, err
	}
	return langMap, nil
}

func writeLangToDisk(path string, data []byte) error {
	if err := ensureLangCacheDir(); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func readVanillaLangFromRepo() ([]byte, error) {
	for _, path := range config.VanillaLangCandidates() {
		data, err := os.ReadFile(path)
		if err == nil {
			return data, nil
		}
	}

	if err := ensureVanillaRepo(); err != nil {
		return nil, err
	}

	for _, path := range config.VanillaLangCandidates() {
		data, err := os.ReadFile(path)
		if err == nil {
			return data, nil
		}
	}

	return nil, fmt.Errorf("vanilla lang not found in %s", config.VanillaAssetsPath())
}

func ensureVanillaRepo() error {
	vanillaPath := config.VanillaAssetsPath()
	if _, err := os.Stat(vanillaPath); err == nil {
		return nil
	}

	parent := filepath.Dir(vanillaPath)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "clone", "https://gitee.com/MFBg7r/minecraft-assets", vanillaPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %v (%s)", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func readLangFromModJar(modID string) ([]byte, error) {
	modsPath := config.ModsPath()
	files, err := os.ReadDir(modsPath)
	if err != nil {
		return nil, fmt.Errorf("read mods dir failed: %v", err)
	}

	langFiles := []string{
		"assets/" + modID + "/lang/zh_ch.json",
		"assets/" + modID + "/lang/zh_cn.json",
		"assets/" + modID + "/lang/en_us.json",
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := strings.ToLower(file.Name())
		if !strings.HasSuffix(name, ".jar") || !strings.Contains(name, strings.ToLower(modID)) {
			continue
		}

		zipPath := filepath.Join(modsPath, file.Name())
		r, err := zip.OpenReader(zipPath)
		if err != nil {
			continue
		}
		defer r.Close()

		for _, target := range langFiles {
			for _, f := range r.File {
				if !strings.EqualFold(f.Name, target) {
					continue
				}
				rc, err := f.Open()
				if err != nil {
					return nil, err
				}
				defer rc.Close()
				data, err := io.ReadAll(rc)
				if err != nil {
					return nil, err
				}
				return data, nil
			}
		}
	}

	return nil, fmt.Errorf("lang file not found")
}
