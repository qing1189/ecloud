package account

import (
	"time"
)

// Account 账号信息
type Account struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Type          string        `json:"type"` // "public" 或 "business"
	Username      string        `json:"username,omitempty"`
	Password      string        `json:"password,omitempty"`
	AccessKey     string        `json:"access_key,omitempty"`
	SecretKey     string        `json:"secret_key,omitempty"`
	PoolID        string        `json:"pool_id,omitempty"`
	MonitorConfig MonitorConfig `json:"monitor_config"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// MonitorConfig 监控配置
type MonitorConfig struct {
	Enabled  bool     `json:"enabled"`
	Interval int      `json:"interval"` // 秒
	Machines []string `json:"machines"` // 监控的机器ID列表，空表示全部
}

// AccountStore 账号存储结构
type AccountStore struct {
	Accounts []Account `json:"accounts"`
	Version  int       `json:"version"`
}

// ToSafeAccount 转换为安全的账号信息（脱敏）
func (a *Account) ToSafeAccount() SafeAccount {
	return SafeAccount{
		ID:            a.ID,
		Name:          a.Name,
		Type:          a.Type,
		Username:      a.Username,
		MonitorConfig: a.MonitorConfig,
		CreatedAt:     a.CreatedAt,
		UpdatedAt:     a.UpdatedAt,
	}
}

// SafeAccount 安全的账号信息（不包含敏感信息）
type SafeAccount struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Type          string        `json:"type"`
	Username      string        `json:"username,omitempty"`
	MonitorConfig MonitorConfig `json:"monitor_config"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}
