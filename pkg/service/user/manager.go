package user

import (
	"database/sql"
	"ecloud_computer_auto_boot/pkg/store"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrUserNotFound      = errors.New("用户不存在")
	ErrUserExists        = errors.New("用户名已存在")
	ErrInvalidUsername   = errors.New("用户名格式不正确")
	ErrInvalidRole       = errors.New("角色必须为 admin 或 user")
	ErrCannotDeleteAdmin = errors.New("不能删除最后一个管理员账号")
)

// Manager 用户管理器
type Manager struct {
	mu sync.RWMutex
}

// NewManager 创建用户管理器
func NewManager() *Manager {
	return &Manager{}
}

// userRow 数据库行
type userRow struct {
	ID           string
	Username     string
	PasswordHash string
	Role         string
	DisplayName  string
	Email        string
	CreatedAt    string
	UpdatedAt    string
	LastLoginAt  string
}

// toUser 转换为 User 结构体
func (r *userRow) toUser() User {
	return User{
		ID:           r.ID,
		Username:     r.Username,
		PasswordHash: r.PasswordHash,
		Role:         r.Role,
		DisplayName:  r.DisplayName,
		Email:        r.Email,
		CreatedAt:    parseTime(r.CreatedAt),
		UpdatedAt:    parseTime(r.UpdatedAt),
		LastLoginAt:  parseTime(r.LastLoginAt),
	}
}

// parseTime 解析 SQLite 时间字符串
func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		return time.Time{}
	}
	return t
}

type scannable interface {
	Scan(dest ...interface{}) error
}

// scanUser 扫描一行数据
func scanUser(row scannable) (*userRow, error) {
	r := &userRow{}
	err := row.Scan(&r.ID, &r.Username, &r.PasswordHash, &r.Role,
		&r.DisplayName, &r.Email, &r.CreatedAt, &r.UpdatedAt, &r.LastLoginAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return r, err
}

// validateUsername 验证用户名格式（3-32位，仅字母数字下划线）
func validateUsername(username string) error {
	if len(username) < 3 || len(username) > 32 {
		return ErrInvalidUsername
	}
	for _, c := range username {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return ErrInvalidUsername
		}
	}
	return nil
}

// CreateUser 创建用户
func (m *Manager) CreateUser(username, passwordHash, role, displayName, email string) (*User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 验证用户名
	if err := validateUsername(username); err != nil {
		return nil, err
	}

	// 验证角色
	if role != "admin" && role != "user" {
		return nil, ErrInvalidRole
	}

	// 检查用户名是否已存在（不区分大小写）
	var count int
	err := store.DB().QueryRow("SELECT COUNT(*) FROM users WHERE LOWER(username) = LOWER(?)", username).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrUserExists
	}

	// 生成 ID
	id := fmt.Sprintf("user_%d", time.Now().UnixNano())
	now := time.Now()

	user := User{
		ID:           id,
		Username:     username,
		PasswordHash: passwordHash,
		Role:         role,
		DisplayName:  displayName,
		Email:        email,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	_, err = store.DB().Exec(`INSERT INTO users
		(id, username, password_hash, role, display_name, email, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		user.ID, user.Username, user.PasswordHash, user.Role,
		user.DisplayName, user.Email,
		user.CreatedAt.Format("2006-01-02 15:04:05"),
		user.UpdatedAt.Format("2006-01-02 15:04:05"))

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByID 根据 ID 获取用户
func (m *Manager) GetUserByID(id string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	row := store.DB().QueryRow(`SELECT id, username, password_hash, role, display_name, email,
		created_at, updated_at, last_login_at FROM users WHERE id = ?`, id)

	r, err := scanUser(row)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, ErrUserNotFound
	}

	user := r.toUser()
	return &user, nil
}

// GetUserByUsername 根据用户名获取用户（不区分大小写）
func (m *Manager) GetUserByUsername(username string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	row := store.DB().QueryRow(`SELECT id, username, password_hash, role, display_name, email,
		created_at, updated_at, last_login_at FROM users WHERE LOWER(username) = LOWER(?)`, username)

	r, err := scanUser(row)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, ErrUserNotFound
	}

	user := r.toUser()
	return &user, nil
}

// ListUsers 列出所有用户
func (m *Manager) ListUsers() ([]SafeUser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rows, err := store.DB().Query(`SELECT id, username, password_hash, role, display_name, email,
		created_at, updated_at, last_login_at FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []SafeUser
	for rows.Next() {
		r, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		if r != nil {
			user := r.toUser()
			users = append(users, user.ToSafeUser())
		}
	}

	if users == nil {
		users = []SafeUser{}
	}
	return users, rows.Err()
}

// UpdateUser 更新用户信息
func (m *Manager) UpdateUser(id string, updater func(*User) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 查询当前用户
	row := store.DB().QueryRow(`SELECT id, username, password_hash, role, display_name, email,
		created_at, updated_at, last_login_at FROM users WHERE id = ?`, id)

	r, err := scanUser(row)
	if err != nil {
		return err
	}
	if r == nil {
		return ErrUserNotFound
	}

	user := r.toUser()

	// 执行更新
	if err := updater(&user); err != nil {
		return err
	}

	// 更新时间
	user.UpdatedAt = time.Now()

	_, err = store.DB().Exec(`UPDATE users SET
		username=?, password_hash=?, role=?, display_name=?, email=?, updated_at=?, last_login_at=?
		WHERE id=?`,
		user.Username, user.PasswordHash, user.Role, user.DisplayName, user.Email,
		user.UpdatedAt.Format("2006-01-02 15:04:05"),
		user.LastLoginAt.Format("2006-01-02 15:04:05"),
		id)

	return err
}

// UpdatePassword 更新用户密码
func (m *Manager) UpdatePassword(id string, newPasswordHash string) error {
	return m.UpdateUser(id, func(u *User) error {
		u.PasswordHash = newPasswordHash
		return nil
	})
}

// UpdateLastLogin 更新最后登录时间
func (m *Manager) UpdateLastLogin(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, err := store.DB().Exec("UPDATE users SET last_login_at = datetime('now','localtime') WHERE id = ?", id)
	return err
}

// DeleteUser 删除用户
func (m *Manager) DeleteUser(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查用户是否存在
	var role string
	err := store.DB().QueryRow("SELECT role FROM users WHERE id = ?", id).Scan(&role)
	if err == sql.ErrNoRows {
		return ErrUserNotFound
	}
	if err != nil {
		return err
	}

	// 如果是管理员，检查是否为最后一个管理员
	if role == "admin" {
		var adminCount int
		err := store.DB().QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&adminCount)
		if err != nil {
			return err
		}
		if adminCount <= 1 {
			return ErrCannotDeleteAdmin
		}
	}

	// 开启事务
	tx, err := store.DB().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 删除用户关联的云账号
	_, err = tx.Exec("DELETE FROM accounts WHERE user_id = ?", id)
	if err != nil {
		return err
	}

	// 删除用户
	result, err := tx.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return err
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return ErrUserNotFound
	}

	return tx.Commit()
}

// CountAdmins 统计管理员数量
func (m *Manager) CountAdmins() (int, error) {
	var count int
	err := store.DB().QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&count)
	return count, err
}

// HasUsers 检查是否存在用户
func (m *Manager) HasUsers() (bool, error) {
	var count int
	err := store.DB().QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count > 0, err
}
