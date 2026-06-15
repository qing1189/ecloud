package auth

import (
	"ecloud_computer_auto_boot/pkg/store"
	"errors"
	"time"
)

var (
	ErrInvalidPassword = errors.New("密码错误")
	ErrUserNotFound    = errors.New("用户不存在")
)

// AuthStore 认证信息存储
type AuthStore struct {
	Admin     AdminAuth `json:"admin"`
	SecretKey string    `json:"secret_key"` // JWT 密钥
}

// AdminAuth 管理员认证信息
type AdminAuth struct {
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
}

// Manager 认证管理器
type Manager struct {
	secretKey string
}

// NewManager 创建认证管理器
func NewManager() *Manager {
	return &Manager{}
}

// Init 初始化认证信息（首次启动生成密码）
func (m *Manager) Init() (string, error) {
	// 检查是否已存在
	if store.FileExists(store.AuthFile) {
		// 加载密钥
		var authStore AuthStore
		if err := store.ReadJSON(store.AuthFile, &authStore); err != nil {
			return "", err
		}
		m.secretKey = authStore.SecretKey
		return "", nil // 已初始化，返回空密码
	}

	// 首次初始化：生成随机密码
	password, err := GenerateRandomPassword(12)
	if err != nil {
		return "", err
	}

	// 生成密钥
	secretKey, err := GenerateSecretKey()
	if err != nil {
		return "", err
	}
	m.secretKey = secretKey

	// 哈希密码（这里用简单的方式，稍后用 bcrypt 替换）
	passwordHash := hashPassword(password)

	authStore := AuthStore{
		Admin: AdminAuth{
			PasswordHash: passwordHash,
			CreatedAt:    time.Now(),
		},
		SecretKey: secretKey,
	}

	if err := store.WriteJSON(store.AuthFile, authStore); err != nil {
		return "", err
	}

	return password, nil
}

// VerifyPassword 验证密码
func (m *Manager) VerifyPassword(password string) error {
	var authStore AuthStore
	if err := store.ReadJSON(store.AuthFile, &authStore); err != nil {
		return err
	}

	if authStore.Admin.PasswordHash == "" {
		return ErrUserNotFound
	}

	if !checkPassword(password, authStore.Admin.PasswordHash) {
		return ErrInvalidPassword
	}

	return nil
}

// ChangePassword 修改密码
func (m *Manager) ChangePassword(oldPassword, newPassword string) error {
	// 验证旧密码
	if err := m.VerifyPassword(oldPassword); err != nil {
		return err
	}

	// 更新密码
	var authStore AuthStore
	if err := store.ReadJSON(store.AuthFile, &authStore); err != nil {
		return err
	}

	authStore.Admin.PasswordHash = hashPassword(newPassword)

	return store.WriteJSON(store.AuthFile, authStore)
}

// GetSecretKey 获取 JWT 密钥
func (m *Manager) GetSecretKey() string {
	return m.secretKey
}

// 简单的密码哈希（临时实现，后续用 bcrypt）
func hashPassword(password string) string {
	// TODO: 替换为 bcrypt
	return base64Encode(password)
}

func checkPassword(password, hash string) bool {
	// TODO: 替换为 bcrypt
	return base64Encode(password) == hash
}

func base64Encode(s string) string {
	return "hash_" + s // 临时简单实现
}
