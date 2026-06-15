package handlers

import (
	"ecloud_computer_auto_boot/pkg/api/types"
	"ecloud_computer_auto_boot/pkg/service/auth"
	"net/http"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authManager *auth.Manager
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authManager *auth.Manager) *AuthHandler {
	return &AuthHandler{
		authManager: authManager,
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Password string `json:"password"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string `json:"token"`
}

// Login 登录
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := types.ParseJSON(r, &req); err != nil {
		types.RespondJSON(w, http.StatusBadRequest, types.Response{
			Success: false,
			Message: "请求格式错误",
		})
		return
	}

	// 验证密码
	if err := h.authManager.VerifyPassword(req.Password); err != nil {
		types.RespondJSON(w, http.StatusUnauthorized, types.Response{
			Success: false,
			Message: "密码错误",
		})
		return
	}

	// 生成 Token
	token, err := h.authManager.GenerateToken("admin")
	if err != nil {
		types.RespondJSON(w, http.StatusInternalServerError, types.Response{
			Success: false,
			Message: "生成令牌失败",
		})
		return
	}

	types.RespondJSON(w, http.StatusOK, types.Response{
		Success: true,
		Data: LoginResponse{
			Token: token,
		},
	})
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req ChangePasswordRequest
	if err := types.ParseJSON(r, &req); err != nil {
		types.RespondJSON(w, http.StatusBadRequest, types.Response{
			Success: false,
			Message: "请求格式错误",
		})
		return
	}

	// 验证新密码长度
	if len(req.NewPassword) < 8 {
		types.RespondJSON(w, http.StatusBadRequest, types.Response{
			Success: false,
			Message: "新密码长度至少为 8 位",
		})
		return
	}

	// 修改密码
	if err := h.authManager.ChangePassword(req.OldPassword, req.NewPassword); err != nil {
		types.RespondJSON(w, http.StatusUnauthorized, types.Response{
			Success: false,
			Message: "旧密码错误",
		})
		return
	}

	types.RespondJSON(w, http.StatusOK, types.Response{
		Success: true,
		Message: "密码修改成功",
	})
}

// VerifyToken 验证 Token
func (h *AuthHandler) VerifyToken(w http.ResponseWriter, r *http.Request) {
	// 如果能到达这里，说明中间件已验证通过
	types.RespondJSON(w, http.StatusOK, types.Response{
		Success: true,
		Message: "Token 有效",
	})
}
