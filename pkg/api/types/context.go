package types

import (
	"context"
	"net/http"
)

// Context key 类型
type contextKey string

const (
	// ContextKeyUserID 用户 ID
	ContextKeyUserID contextKey = "user_id"
	// ContextKeyUsername 用户名
	ContextKeyUsername contextKey = "username"
	// ContextKeyUserRole 用户角色
	ContextKeyUserRole contextKey = "user_role"
)

// GetUserIDFromContext 从 Context 获取用户 ID
func GetUserIDFromContext(r *http.Request) string {
	if userID, ok := r.Context().Value(ContextKeyUserID).(string); ok {
		return userID
	}
	return ""
}

// GetUsernameFromContext 从 Context 获取用户名
func GetUsernameFromContext(r *http.Request) string {
	if username, ok := r.Context().Value(ContextKeyUsername).(string); ok {
		return username
	}
	return ""
}

// GetUserRoleFromContext 从 Context 获取用户角色
func GetUserRoleFromContext(r *http.Request) string {
	if role, ok := r.Context().Value(ContextKeyUserRole).(string); ok {
		return role
	}
	return ""
}

// IsAdmin 检查当前用户是否为管理员
func IsAdmin(r *http.Request) bool {
	return GetUserRoleFromContext(r) == "admin"
}

// WithUserContext 将用户信息存入 Context
func WithUserContext(ctx context.Context, userID, username, role string) context.Context {
	ctx = context.WithValue(ctx, ContextKeyUserID, userID)
	ctx = context.WithValue(ctx, ContextKeyUsername, username)
	ctx = context.WithValue(ctx, ContextKeyUserRole, role)
	return ctx
}
