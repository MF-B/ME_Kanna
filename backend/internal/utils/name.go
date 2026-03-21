package utils

import (
	"ME_Kanna/internal/config"
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
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

	// 兜底：可能命中了旧缓存或错误语言，强制从源刷新一次
	if refreshed, err := loadLangMapFromSource(modID); err == nil {
		storeLangMap(modID, refreshed)
		cachePath := filepath.Join(config.LangCacheDir(), modID+".json")
		if data, marshalErr := json.Marshal(refreshed); marshalErr == nil {
			_ = writeLangToDisk(cachePath, data)
		}
		if name, ok := lookupItemName(refreshed, modID, itemName); ok {
			return name, nil
		}
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

	langMap, err := loadLangMapFromSource(modID)
	if err != nil {
		return nil, err
	}
	if data, marshalErr := json.Marshal(langMap); marshalErr == nil {
		_ = writeLangToDisk(cachePath, data)
	}
	storeLangMap(modID, langMap)
	return langMap, nil
}

func loadLangMapFromSource(modID string) (map[string]string, error) {
	merged := make(map[string]string)

	var baseErr error
	if modID == "minecraft" {
		if data, err := readVanillaLangFromRepo(); err == nil {
			if baseMap, parseErr := parseLangData(data); parseErr == nil {
				mergeLangMap(merged, baseMap)
			} else {
				baseErr = parseErr
			}
		} else {
			baseErr = err
		}
	} else {
		if data, err := readLangFromModJar(modID); err == nil {
			if baseMap, parseErr := parseLangData(data); parseErr == nil {
				mergeLangMap(merged, baseMap)
			} else {
				baseErr = parseErr
			}
		} else {
			baseErr = err
		}
	}

	if rpMap, err := readLangFromResourcePacks(modID); err == nil {
		mergeLangMap(merged, rpMap)
	}

	if len(merged) == 0 {
		if baseErr != nil {
			return nil, baseErr
		}
		return nil, fmt.Errorf("lang file not found")
	}

	return merged, nil
}

func parseLangData(data []byte) (map[string]string, error) {
	langMap := make(map[string]string)
	if err := json.Unmarshal(data, &langMap); err != nil {
		return nil, err
	}
	return langMap, nil
}

func mergeLangMap(dst map[string]string, src map[string]string) {
	for key, value := range src {
		dst[key] = value
	}
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
		"assets/" + modID + "/lang/zh_cn.json",
		"assets/" + modID + "/lang/zh_ch.json",
		"assets/" + modID + "/lang/en_us.json",
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := strings.ToLower(file.Name())
		if !strings.HasSuffix(name, ".jar") {
			continue
		}

		zipPath := filepath.Join(modsPath, file.Name())
		if data, err := readFileFromZip(zipPath, langFiles); err == nil {
			return data, nil
		}
	}

	return nil, fmt.Errorf("lang file not found")
}

func readLangFromResourcePacks(modID string) (map[string]string, error) {
	packPaths := orderedResourcePackPaths()
	if len(packPaths) == 0 {
		return nil, fmt.Errorf("no resourcepacks")
	}

	langFiles := []string{
		"assets/" + modID + "/lang/zh_cn.json",
		"assets/" + modID + "/lang/zh_ch.json",
		"assets/" + modID + "/lang/en_us.json",
	}

	merged := make(map[string]string)
	found := false

	for _, packPath := range packPaths {
		var data []byte
		var err error
		if strings.HasSuffix(strings.ToLower(packPath), ".zip") {
			data, err = readFileFromZip(packPath, langFiles)
		} else {
			data, err = readFileFromDir(packPath, langFiles)
		}
		if err != nil {
			continue
		}
		langMap, err := parseLangData(data)
		if err != nil {
			continue
		}
		mergeLangMap(merged, langMap)
		found = true
	}

	if !found {
		return nil, fmt.Errorf("resourcepack lang not found")
	}

	return merged, nil
}

func orderedResourcePackPaths() []string {
	dir := config.ResourcePacksDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	available := make(map[string]string)
	for _, entry := range entries {
		name := entry.Name()
		if !(entry.IsDir() || strings.HasSuffix(strings.ToLower(name), ".zip")) {
			continue
		}
		available[name] = filepath.Join(dir, name)
	}

	orderedNames := readResourcePackOrderFromOptions()
	result := make([]string, 0, len(available))
	used := make(map[string]bool)

	for _, name := range orderedNames {
		if path, ok := available[name]; ok {
			result = append(result, path)
			used[name] = true
		}
	}

	remaining := make([]string, 0)
	for name := range available {
		if !used[name] {
			remaining = append(remaining, name)
		}
	}
	sort.Strings(remaining)
	for _, name := range remaining {
		result = append(result, available[name])
	}

	return result
}

func readResourcePackOrderFromOptions() []string {
	data, err := os.ReadFile(config.OptionsFilePath())
	if err != nil {
		return nil
	}
	content := string(data)
	line := ""
	for _, raw := range strings.Split(content, "\n") {
		if strings.HasPrefix(raw, "resourcePacks:") {
			line = raw
			break
		}
	}
	if line == "" {
		return nil
	}

	re := regexp.MustCompile(`"([^"]+)"`)
	matches := re.FindAllStringSubmatch(line, -1)
	result := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		entry := strings.TrimSpace(match[1])
		if entry == "vanilla" || strings.HasPrefix(entry, "fabric") {
			continue
		}
		if after, ok := strings.CutPrefix(entry, "file/"); ok {
			entry = after
		}
		result = append(result, entry)
	}
	return result
}

func readFileFromDir(dirPath string, candidates []string) ([]byte, error) {
	for _, relative := range candidates {
		path := filepath.Join(dirPath, filepath.FromSlash(relative))
		data, err := os.ReadFile(path)
		if err == nil {
			return data, nil
		}
	}
	return nil, fmt.Errorf("file not found in dir")
}

func readFileFromZip(zipPath string, candidates []string) ([]byte, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	for _, target := range candidates {
		for _, f := range r.File {
			if !strings.EqualFold(f.Name, target) {
				continue
			}
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			data, readErr := io.ReadAll(rc)
			_ = rc.Close()
			if readErr != nil {
				return nil, readErr
			}
			return data, nil
		}
	}

	return nil, fmt.Errorf("file not found in zip")
}
