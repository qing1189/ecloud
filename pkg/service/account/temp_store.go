package account

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"ecloud_computer_auto_boot/pkg/ecloud"
)

// TempAccount 临时账号（等待验证）
type TempAccount struct {
	ID        string         `json:"id"`
	Client    *ecloud.Client `json:"-"`
	Account   *Account       `json:"account"`
	CreatedAt time.Time      `json:"createdAt"`
}

var (
	tempStore = make(map[string]*TempAccount)
	tempMutex sync.RWMutex
)

// SaveTempAccount 保存临时账号
func SaveTempAccount(client *ecloud.Client, account *Account) string {
	tempMutex.Lock()
	defer tempMutex.Unlock()

	id := "temp_" + generateTempID()
	tempStore[id] = &TempAccount{
		ID:        id,
		Client:    client,
		Account:   account,
		CreatedAt: time.Now(),
	}

	// 5分钟后自动清理
	time.AfterFunc(5*time.Minute, func() {
		DeleteTempAccount(id)
	})

	return id
}

// GetTempAccount 获取临时账号
func GetTempAccount(id string) (*TempAccount, error) {
	tempMutex.RLock()
	defer tempMutex.RUnlock()

	temp, ok := tempStore[id]
	if !ok {
		return nil, errors.New("临时账号不存在或已过期")
	}

	return temp, nil
}

// DeleteTempAccount 删除临时账号
func DeleteTempAccount(id string) {
	tempMutex.Lock()
	defer tempMutex.Unlock()

	delete(tempStore, id)
}

// generateTempID 生成临时ID
func generateTempID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
