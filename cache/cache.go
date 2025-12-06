package cache

import (
	"encoding/json"
	"fmt"
	"loto-suite/backend/logging"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type cacheEntry struct {
	Data      json.RawMessage `json:"data"`
	ExpiresAt time.Time       `json:"expires_at"`
}

type cache struct {
	entries  map[string]*cacheEntry
	mutex    sync.RWMutex
	cacheDir string
}

var (
	globalCache *cache
	once        sync.Once
)

func getCache() *cache {
	once.Do(func() {
		if _, filename, _, ok := runtime.Caller(0); ok {
			cachingDir := filepath.Dir(filename)
			cacheDir := filepath.Join(cachingDir, "cache")

			globalCache = &cache{
				entries:  make(map[string]*cacheEntry),
				cacheDir: cacheDir,
			}

			_ = os.MkdirAll(cacheDir, 0755)
			globalCache.loadFromDisk()
		}
	})

	return globalCache
}

func (c *cache) getCacheKey(game string, month string, year string) string {
	return fmt.Sprintf("%s_%s_%s", game, month, year)
}

func (c *cache) getCacheFilePath(key string) string {
	return filepath.Join(c.cacheDir, key+".json")
}

func Get(gameId string, month string, year string) (json.RawMessage, bool) {
	cache := getCache()
	if cache == nil {
		return nil, false
	}

	return getCache().get(gameId, month, year)
}

func (c *cache) get(gameId string, month string, year string) (json.RawMessage, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	key := c.getCacheKey(gameId, month, year)
	entry, exists := c.entries[key]

	if !exists {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		delete(c.entries, key)
		os.Remove(c.getCacheFilePath(key))
		return nil, false
	}

	return entry.Data, true
}

func Set(gameId string, month string, year string, data json.RawMessage, ttl time.Duration) {
	getCache().set(gameId, month, year, data, ttl)
}

func (c *cache) set(gameId string, month string, year string, data json.RawMessage, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	key := c.getCacheKey(gameId, month, year)
	now := time.Now()

	entry := &cacheEntry{
		Data:      data,
		ExpiresAt: now.Add(ttl),
	}

	c.entries[key] = entry

	c.saveToDisk(key, entry)
}

func (c *cache) saveToDisk(key string, entry *cacheEntry) {
	filePath := c.getCacheFilePath(key)
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		logging.ErrorBe(err.Error())
		return
	}

	os.WriteFile(filePath, data, 0644)
}

func (c *cache) loadFromDisk() {
	files, err := os.ReadDir(c.cacheDir)
	if err != nil {
		logging.ErrorBe(err.Error())
		return
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			key := file.Name()[:len(file.Name())-5]
			c.loadEntryFromDisk(key)
		}
	}
}

func (c *cache) loadEntryFromDisk(key string) {
	filePath := c.getCacheFilePath(key)
	data, err := os.ReadFile(filePath)
	if err != nil {
		logging.ErrorBe(err.Error())
		return
	}

	var entry cacheEntry
	err = json.Unmarshal(data, &entry)
	if err != nil {
		logging.ErrorBe(err.Error())
		os.Remove(filePath)
		return
	}

	if time.Now().After(entry.ExpiresAt) {
		os.Remove(filePath)
		return
	}

	c.entries[key] = &entry
}

func ClearCache() {
	getCache().clear()
}

func (c *cache) clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.entries = make(map[string]*cacheEntry)

	files, err := os.ReadDir(c.cacheDir)
	if err != nil {
		logging.ErrorBe(err.Error())
		return
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			os.Remove(filepath.Join(c.cacheDir, file.Name()))
		}
	}
}
