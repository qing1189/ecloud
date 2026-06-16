package user

import "time"

// User 用户信息
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // 不输出到 JSON
	Role         string    `json:"role"`
	DisplayName  string    `json:"display_name"`
	Email        string    `json:"email"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	LastLoginAt  time.Time `json:"last_login_at,omitempty"`
}

// SafeUser 安全的用户信息（不包含敏感字段）
type SafeUser struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	Role        string    `json:"role"`
	DisplayName string    `json:"display_name"`
	Email       string    `json:"email"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	LastLoginAt time.Time `json:"last_login_at,omitempty"`
}

// ToSafeUser 转换为安全的用户信息（脱敏）
func (u *User) ToSafeUser() SafeUser {
	return SafeUser{
		ID:          u.ID,
		Username:    u.Username,
		Role:        u.Role,
		DisplayName: u.DisplayName,
		Email:       u.Email,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
		LastLoginAt: u.LastLoginAt,
	}
}
