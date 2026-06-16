package account

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
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

// accountRow 扫描数据库行
type accountRow struct {
	ID              string
	Name            string
	Type            string
	Username        string
	Password        string
	AccessKey       string
	SecretKey       string
	PoolID          string
	MonitorEnabled  int
	MonitorInterval int
	MonitorMachines string
	UserID          string
	CreatedAt       string
	UpdatedAt       string
}

// toAccount 数据库行转 Account 结构体
func (r *accountRow) toAccount() Account {
	var machines []string
	json.Unmarshal([]byte(r.MonitorMachines), &machines)

	return Account{
		ID:        r.ID,
		Name:      r.Name,
		Type:      r.Type,
		Username:  r.Username,
		Password:  r.Password,
		AccessKey: r.AccessKey,
		SecretKey: r.SecretKey,
		PoolID:    r.PoolID,
		UserID:    r.UserID,
		MonitorConfig: MonitorConfig{
			Enabled:  r.MonitorEnabled == 1,
			Interval: r.MonitorInterval,
			Machines: machines,
		},
		CreatedAt: parseTime(r.CreatedAt),
		UpdatedAt: parseTime(r.UpdatedAt),
	}
}

// scanAccount 扫描一行数据到 accountRow
func scanAccount(row scannable) (*accountRow, error) {
	r := &accountRow{}
	err := row.Scan(
		&r.ID, &r.Name, &r.Type, &r.Username,
		&r.Password, &r.AccessKey, &r.SecretKey,
		&r.PoolID, &r.MonitorEnabled, &r.MonitorInterval,
		&r.MonitorMachines, &r.UserID, &r.CreatedAt, &r.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return r, err
}

type scannable interface {
	Scan(dest ...interface{}) error
}

// parseTime 解析 SQLite 时间字符串
func parseTime(s string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		return time.Now()
	}
	return t
}

// LoadAccounts 加载所有账号
func (m *Manager) LoadAccounts() ([]Account, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rows, err := store.DB().Query(`SELECT id, name, type, username, password,
		access_key, secret_key, pool_id,
		monitor_enabled, monitor_interval, monitor_machines, user_id,
		created_at, updated_at FROM accounts ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		r, err := scanAccount(rows)
		if err != nil {
			return nil, err
		}
		if r != nil {
			accounts = append(accounts, r.toAccount())
		}
	}
	if accounts == nil {
		accounts = []Account{}
	}
	return accounts, rows.Err()
}

// SaveAccounts 保存所有账号（全量替换）
func (m *Manager) SaveAccounts(accounts []Account) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tx, err := store.DB().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 清空旧数据
	if _, err := tx.Exec("DELETE FROM accounts"); err != nil {
		return err
	}

	stmt, err := tx.Prepare(`INSERT INTO accounts
		(id, name, type, username, password, access_key, secret_key, pool_id,
		 monitor_enabled, monitor_interval, monitor_machines, user_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, acc := range accounts {
		enabled := 0
		if acc.MonitorConfig.Enabled {
			enabled = 1
		}
		machines := store.JSONString(acc.MonitorConfig.Machines)

		_, err := stmt.Exec(
			acc.ID, acc.Name, acc.Type, acc.Username,
			acc.Password, acc.AccessKey, acc.SecretKey, acc.PoolID,
			enabled, acc.MonitorConfig.Interval, machines, acc.UserID,
			acc.CreatedAt.Format("2006-01-02 15:04:05"),
			acc.UpdatedAt.Format("2006-01-02 15:04:05"),
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
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
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查 ID 是否已存在
	var count int
	err := store.DB().QueryRow("SELECT COUNT(*) FROM accounts WHERE id = ?", acc.ID).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrAccountExists
	}

	// 生成ID
	if acc.ID == "" {
		acc.ID = fmt.Sprintf("acc_%d", time.Now().Unix())
	}

	// 设置时间
	now := time.Now()
	acc.CreatedAt = now
	acc.UpdatedAt = now

	// 加密密码
	if acc.Password != "" {
		acc.Password = base64.StdEncoding.EncodeToString([]byte(acc.Password))
	}
	if acc.SecretKey != "" {
		acc.SecretKey = base64.StdEncoding.EncodeToString([]byte(acc.SecretKey))
	}

	enabled := 0
	if acc.MonitorConfig.Enabled {
		enabled = 1
	}
	machines := store.JSONString(acc.MonitorConfig.Machines)

	_, err = store.DB().Exec(`INSERT INTO accounts
		(id, name, type, username, password, access_key, secret_key, pool_id,
		 monitor_enabled, monitor_interval, monitor_machines, user_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		acc.ID, acc.Name, acc.Type, acc.Username,
		acc.Password, acc.AccessKey, acc.SecretKey, acc.PoolID,
		enabled, acc.MonitorConfig.Interval, machines, acc.UserID,
		acc.CreatedAt.Format("2006-01-02 15:04:05"),
		acc.UpdatedAt.Format("2006-01-02 15:04:05"),
	)
	return err
}

// UpdateAccount 更新账号
func (m *Manager) UpdateAccount(id string, updater func(*Account) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 查询当前账号
	var r accountRow
	err := store.DB().QueryRow(`SELECT id, name, type, username, password,
		access_key, secret_key, pool_id,
		monitor_enabled, monitor_interval, monitor_machines, user_id,
		created_at, updated_at FROM accounts WHERE id = ?`, id).Scan(
		&r.ID, &r.Name, &r.Type, &r.Username,
		&r.Password, &r.AccessKey, &r.SecretKey,
		&r.PoolID, &r.MonitorEnabled, &r.MonitorInterval,
		&r.MonitorMachines, &r.UserID, &r.CreatedAt, &r.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return ErrAccountNotFound
	}
	if err != nil {
		return err
	}

	acc := r.toAccount()

	// 解密
	if acc.Password != "" {
		if decoded, err := base64.StdEncoding.DecodeString(acc.Password); err == nil {
			acc.Password = string(decoded)
		}
	}
	if acc.SecretKey != "" {
		if decoded, err := base64.StdEncoding.DecodeString(acc.SecretKey); err == nil {
			acc.SecretKey = string(decoded)
		}
	}

	// 执行更新
	if err := updater(&acc); err != nil {
		return err
	}

	// 重新加密
	if acc.Password != "" {
		acc.Password = base64.StdEncoding.EncodeToString([]byte(acc.Password))
	}
	if acc.SecretKey != "" {
		acc.SecretKey = base64.StdEncoding.EncodeToString([]byte(acc.SecretKey))
	}

	acc.UpdatedAt = time.Now()

	enabled := 0
	if acc.MonitorConfig.Enabled {
		enabled = 1
	}
	machines := store.JSONString(acc.MonitorConfig.Machines)

	_, err = store.DB().Exec(`UPDATE accounts SET
		name=?, type=?, username=?, password=?, access_key=?, secret_key=?, pool_id=?,
		monitor_enabled=?, monitor_interval=?, monitor_machines=?, user_id=?,
		updated_at=?
		WHERE id=?`,
		acc.Name, acc.Type, acc.Username, acc.Password,
		acc.AccessKey, acc.SecretKey, acc.PoolID,
		enabled, acc.MonitorConfig.Interval, machines, acc.UserID,
		acc.UpdatedAt.Format("2006-01-02 15:04:05"), id,
	)
	return err
}

// DeleteAccount 删除账号
func (m *Manager) DeleteAccount(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	result, err := store.DB().Exec("DELETE FROM accounts WHERE id = ?", id)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return ErrAccountNotFound
	}
	return nil
}

// GetAccountCredentials 获取账号凭证（解密）
func (m *Manager) GetAccountCredentials(id string) (username, password, accessKey, secretKey string, err error) {
	acc, err := m.GetAccount(id)
	if err != nil {
		return "", "", "", "", err
	}

	if acc.Password != "" {
		decoded, err := base64.StdEncoding.DecodeString(acc.Password)
		if err != nil {
			return "", "", "", "", fmt.Errorf("解密密码失败: %w", err)
		}
		password = string(decoded)
	}

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

// ListAccountsByUser 列出指定用户的账号
func (m *Manager) ListAccountsByUser(userID string) ([]SafeAccount, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rows, err := store.DB().Query(`SELECT id, name, type, username, password,
		access_key, secret_key, pool_id,
		monitor_enabled, monitor_interval, monitor_machines, user_id,
		created_at, updated_at FROM accounts WHERE user_id = ? ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var safeAccounts []SafeAccount
	for rows.Next() {
		r, err := scanAccount(rows)
		if err != nil {
			return nil, err
		}
		if r != nil {
			acc := r.toAccount()
			safeAccounts = append(safeAccounts, acc.ToSafeAccount())
		}
	}

	if safeAccounts == nil {
		safeAccounts = []SafeAccount{}
	}
	return safeAccounts, rows.Err()
}
