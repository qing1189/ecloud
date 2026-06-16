package auth

import (
	"database/sql"
	"ecloud_computer_auto_boot/pkg/service/user"
	"ecloud_computer_auto_boot/pkg/store"
	"errors"
	"os"
)

var (
	ErrInvalidPassword = errors.New("密码错误")
	ErrUserNotFound    = errors.New("用户不存在")
)

// Manager 认证管理器
type Manager struct {
	secretKey   string
	userManager *user.Manager
}

// NewManager 创建认证管理器
func NewManager(userManager *user.Manager) *Manager {
	return &Manager{
		userManager: userManager,
	}
}

// Init 初始化认证信息（加载 JWT 密钥，检查是否需要创建默认管理员）
func (m *Manager) Init() (needCreateAdmin bool, err error) {
	// 1. 尝试从 auth 表加载 secret_key（兼容旧版本）
	var secretKey string
	err = store.DB().QueryRow("SELECT secret_key FROM auth WHERE id = 1").Scan(&secretKey)

	if err == nil {
		// 旧版本存在 auth 表，加载密钥
		m.secretKey = secretKey
	} else if err == sql.ErrNoRows {
		// auth 表为空，生成新密钥并保存
		secretKey, err = GenerateSecretKey()
		if err != nil {
			return false, err
		}
		m.secretKey = secretKey

		// 保存到 auth 表（保持兼容）
		_, _ = store.DB().Exec(`INSERT INTO auth (id, secret_key, created_at)
			VALUES (1, ?, datetime('now','localtime'))`, secretKey)
	} else {
		return false, err
	}

	// 2. 检查是否存在用户
	hasUsers, err := m.userManager.HasUsers()
	if err != nil {
		return false, err
	}

	// 3. 如果没有用户，标记需要创建默认管理员
	if !hasUsers {
		return true, nil
	}

	return false, nil
}

// Login 用户登录
func (m *Manager) Login(username, password string) (token string, userInfo *user.User, err error) {
	// 根据用户名查找用户
	u, err := m.userManager.GetUserByUsername(username)
	if err != nil {
		if err == user.ErrUserNotFound {
			return "", nil, ErrUserNotFound
		}
		return "", nil, err
	}

	// 验证密码
	if !CheckPassword(password, u.PasswordHash) {
		return "", nil, ErrInvalidPassword
	}

	// 更新最后登录时间
	_ = m.userManager.UpdateLastLogin(u.ID)

	// 生成 Token
	token, err = m.GenerateToken(u.ID, u.Username, u.Role)
	if err != nil {
		return "", nil, err
	}

	return token, u, nil
}

// ChangePassword 修改用户密码
func (m *Manager) ChangePassword(userID, oldPassword, newPassword string) error {
	// 获取用户
	u, err := m.userManager.GetUserByID(userID)
	if err != nil {
		return err
	}

	// 验证旧密码
	if !CheckPassword(oldPassword, u.PasswordHash) {
		return ErrInvalidPassword
	}

	// 哈希新密码
	newHash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	// 更新密码
	return m.userManager.UpdatePassword(userID, newHash)
}

// ResetPassword 重置用户密码（管理员操作，无需验证旧密码）
func (m *Manager) ResetPassword(userID, newPassword string) error {
	newHash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	return m.userManager.UpdatePassword(userID, newHash)
}

// GetSecretKey 获取 JWT 密钥
func (m *Manager) GetSecretKey() string {
	return m.secretKey
}

// CreateDefaultAdmin 创建默认管理员账号
func (m *Manager) CreateDefaultAdmin() (username, password string, err error) {
	// 读取环境变量
	username = os.Getenv("ADMIN_USERNAME")
	if username == "" {
		username = "admin"
	}

	password = os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		// 生成随机密码
		password, err = GenerateRandomPassword(12)
		if err != nil {
			return "", "", err
		}
	}

	// 哈希密码
	passwordHash, err := HashPassword(password)
	if err != nil {
		return "", "", err
	}

	// 创建管理员用户
	_, err = m.userManager.CreateUser(username, passwordHash, "admin", "系统管理员", "")
	if err != nil {
		return "", "", err
	}

	return username, password, nil
}
