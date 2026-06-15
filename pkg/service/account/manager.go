package account

import (
	"encoding/base64"
	"ecloud_computer_auto_boot/pkg/store"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrAccountNotFound = errors.New("账号不存在")
	ErrAccountExists   = errors.New("账号已存在")
)

// Manager 账号管理器
type Manager struct {
	mu sync.RWMutex
}

// NewManager 创建账号管理器
func NewManager() *Manager {
	return &Manager{}
}

// LoadAccounts 加载所有账号
func (m *Manager) LoadAccounts() ([]Account, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var accountStore AccountStore
	if err := store.ReadJSON(store.AccountsFile, &accountStore); err != nil {
		return nil, err
	}

	return accountStore.Accounts, nil
}

// SaveAccounts 保存所有账号
func (m *Manager) SaveAccounts(accounts []Account) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	accountStore := AccountStore{
		Accounts: accounts,
		Version:  1,
	}

	return store.WriteJSON(store.AccountsFile, accountStore)
}

// GetAccount 获取单个账号
func (m *Manager) GetAccount(id string) (*Account, error) {
	accounts, err := m.LoadAccounts()
	if err != nil {
		return nil, err
	}

	for _, acc := range accounts {
		if acc.ID == id {
			return &acc, nil
		}
	}

	return nil, ErrAccountNotFound
}

// AddAccount 添加账号
func (m *Manager) AddAccount(acc Account) error {
	accounts, err := m.LoadAccounts()
	if err != nil {
		return err
	}

	// 检查ID是否已存在
	for _, existing := range accounts {
		if existing.ID == acc.ID {
			return ErrAccountExists
		}
	}

	// 生成ID
	if acc.ID == "" {
		acc.ID = fmt.Sprintf("acc_%d", time.Now().Unix())
	}

	// 设置时间
	now := time.Now()
	acc.CreatedAt = now
	acc.UpdatedAt = now

	// 加密密码（简单Base64编码）
	if acc.Password != "" {
		acc.Password = base64.StdEncoding.EncodeToString([]byte(acc.Password))
	}
	if acc.SecretKey != "" {
		acc.SecretKey = base64.StdEncoding.EncodeToString([]byte(acc.SecretKey))
	}

	accounts = append(accounts, acc)
	return m.SaveAccounts(accounts)
}

// UpdateAccount 更新账号
func (m *Manager) UpdateAccount(id string, updater func(*Account) error) error {
	accounts, err := m.LoadAccounts()
	if err != nil {
		return err
	}

	found := false
	for i := range accounts {
		if accounts[i].ID == id {
			// 解密密码
			if accounts[i].Password != "" {
				if decoded, err := base64.StdEncoding.DecodeString(accounts[i].Password); err == nil {
					accounts[i].Password = string(decoded)
				}
			}
			if accounts[i].SecretKey != "" {
				if decoded, err := base64.StdEncoding.DecodeString(accounts[i].SecretKey); err == nil {
					accounts[i].SecretKey = string(decoded)
				}
			}

			// 执行更新
			if err := updater(&accounts[i]); err != nil {
				return err
			}

			// 重新加密
			if accounts[i].Password != "" {
				accounts[i].Password = base64.StdEncoding.EncodeToString([]byte(accounts[i].Password))
			}
			if accounts[i].SecretKey != "" {
				accounts[i].SecretKey = base64.StdEncoding.EncodeToString([]byte(accounts[i].SecretKey))
			}

			accounts[i].UpdatedAt = time.Now()
			found = true
			break
		}
	}

	if !found {
		return ErrAccountNotFound
	}

	return m.SaveAccounts(accounts)
}

// DeleteAccount 删除账号
func (m *Manager) DeleteAccount(id string) error {
	accounts, err := m.LoadAccounts()
	if err != nil {
		return err
	}

	newAccounts := make([]Account, 0, len(accounts))
	found := false
	for _, acc := range accounts {
		if acc.ID != id {
			newAccounts = append(newAccounts, acc)
		} else {
			found = true
		}
	}

	if !found {
		return ErrAccountNotFound
	}

	return m.SaveAccounts(newAccounts)
}

// GetAccountCredentials 获取账号凭证（解密）
func (m *Manager) GetAccountCredentials(id string) (username, password, accessKey, secretKey string, err error) {
	acc, err := m.GetAccount(id)
	if err != nil {
		return "", "", "", "", err
	}

	// 解密密码
	if acc.Password != "" {
		decoded, err := base64.StdEncoding.DecodeString(acc.Password)
		if err != nil {
			return "", "", "", "", fmt.Errorf("解密密码失败: %w", err)
		}
		password = string(decoded)
	}

	// 解密SecretKey
	if acc.SecretKey != "" {
		decoded, err := base64.StdEncoding.DecodeString(acc.SecretKey)
		if err != nil {
			return "", "", "", "", fmt.Errorf("解密SecretKey失败: %w", err)
		}
		secretKey = string(decoded)
	}

	return acc.Username, password, acc.AccessKey, secretKey, nil
}

// ListAccounts 列出所有账号（安全版本）
func (m *Manager) ListAccounts() ([]SafeAccount, error) {
	accounts, err := m.LoadAccounts()
	if err != nil {
		return nil, err
	}

	safeAccounts := make([]SafeAccount, len(accounts))
	for i, acc := range accounts {
		safeAccounts[i] = acc.ToSafeAccount()
	}

	return safeAccounts, nil
}
