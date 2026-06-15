package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

const (
	StoreDir       = "store"
	AccountsFile   = "accounts.json"
	AuthFile       = "auth.json"
	LogsFile       = "logs.json"
)

var (
	storePath string
	mu        sync.RWMutex
)

// Init 初始化存储目录
func Init(basePath string) error {
	if basePath == "" {
		basePath = "."
	}
	storePath = filepath.Join(basePath, StoreDir)

	// 创建存储目录
	if err := os.MkdirAll(storePath, 0755); err != nil {
		return err
	}

	return nil
}

// GetStorePath 获取存储路径
func GetStorePath() string {
	return storePath
}

// ReadJSON 读取 JSON 文件（线程安全）
func ReadJSON(filename string, v interface{}) error {
	mu.RLock()
	defer mu.RUnlock()

	path := filepath.Join(storePath, filename)

	// 文件不存在返回空
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	return json.Unmarshal(data, v)
}

// WriteJSON 写入 JSON 文件（线程安全）
func WriteJSON(filename string, v interface{}) error {
	mu.Lock()
	defer mu.Unlock()

	path := filepath.Join(storePath, filename)

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// FileExists 检查文件是否存在
func FileExists(filename string) bool {
	path := filepath.Join(storePath, filename)
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
