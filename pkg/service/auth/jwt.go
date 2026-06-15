package auth

import (
	"errors"
	"time"
)

var (
	ErrInvalidToken = errors.New("无效的 Token")
	ErrTokenExpired = errors.New("Token 已过期")
)

// Claims JWT 声明
type Claims struct {
	Username  string    `json:"username"`
	ExpiresAt time.Time `json:"expires_at"`
}

// GenerateToken 生成 JWT Token（简化版本）
func (m *Manager) GenerateToken(username string) (string, error) {
	// 24 小时有效期
	expiresAt := time.Now().Add(24 * time.Hour)

	// 简单的 token 格式：username:timestamp:signature
	// TODO: 替换为真正的 JWT
	token := username + ":" + expiresAt.Format(time.RFC3339) + ":" + m.secretKey[:10]

	return token, nil
}

// ValidateToken 验证 JWT Token（简化版本）
func (m *Manager) ValidateToken(token string) (*Claims, error) {
	if token == "" {
		return nil, ErrInvalidToken
	}

	// TODO: 替换为真正的 JWT 验证
	// 临时实现：简单检查
	if len(token) < 20 {
		return nil, ErrInvalidToken
	}

	claims := &Claims{
		Username:  "admin",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	return claims, nil
}
