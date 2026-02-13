package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"mineCCT/internal/config"
	"mineCCT/internal/store"
)

type whitelistFile struct {
	MonitoredItems []string `json:"monitored_items"`
}

type whitelistState struct {
	mu      sync.RWMutex
	items   []string
	version string
	path    string
}

var whitelist = &whitelistState{}

func InitWhitelist() error {
	whitelist.path = config.WhitelistFilePath()
	return loadWhitelistFromFile(whitelist.path)
}

func GetWhitelistSnapshot() ([]string, string) {
	whitelist.mu.RLock()
	defer whitelist.mu.RUnlock()

	items := make([]string, len(whitelist.items))
	copy(items, whitelist.items)
	return items, whitelist.version
}

func EnsureWhitelistFromFactories() ([]string, string, bool, error) {
	items, version := GetWhitelistSnapshot()
	if len(items) > 0 {
		return items, version, false, nil
	}

	s := store.Global
	seen := make(map[string]bool)
	collected := make([]string, 0)

	s.Mutex.RLock()
	for _, factory := range s.Factories {
		for itemID := range factory.Items {
			if itemID == "" || seen[itemID] {
				continue
			}
			seen[itemID] = true
			collected = append(collected, itemID)
		}
	}
	s.Mutex.RUnlock()

	if len(collected) == 0 {
		return items, version, false, nil
	}

	newVersion, err := UpdateWhitelist(collected)
	if err != nil {
		return items, version, false, err
	}

	updated, _ := GetWhitelistSnapshot()
	return updated, newVersion, true, nil
}

func UpdateWhitelist(items []string) (string, error) {
	normalized := normalizeWhitelist(items)
	version := computeWhitelistHash(normalized)
	path := whitelist.path
	if path == "" {
		path = config.WhitelistFilePath()
		whitelist.mu.Lock()
		whitelist.path = path
		whitelist.mu.Unlock()
	}

	if err := saveWhitelistToFile(path, normalized); err != nil {
		return "", err
	}

	whitelist.mu.Lock()
	whitelist.items = normalized
	whitelist.version = version
	whitelist.mu.Unlock()

	return version, nil
}

func loadWhitelistFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			_, err = UpdateWhitelist([]string{})
			return err
		}
		return err
	}

	if len(data) == 0 {
		_, err = UpdateWhitelist([]string{})
		return err
	}

	var file whitelistFile
	if err := json.Unmarshal(data, &file); err != nil || file.MonitoredItems == nil {
		var list []string
		if err := json.Unmarshal(data, &list); err != nil {
			return err
		}
		_, err = UpdateWhitelist(list)
		return err
	}

	_, err = UpdateWhitelist(file.MonitoredItems)
	return err
}

func saveWhitelistToFile(path string, items []string) error {
	payload := whitelistFile{MonitoredItems: items}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func normalizeWhitelist(items []string) []string {
	if items == nil {
		return []string{}
	}

	seen := make(map[string]bool)
	normalized := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		normalized = append(normalized, trimmed)
	}

	sort.Strings(normalized)
	return normalized
}

func computeWhitelistHash(items []string) string {
	payload := whitelistFile{MonitoredItems: items}
	data, _ := json.Marshal(payload)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
