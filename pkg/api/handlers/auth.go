package handlers

import (
	"ecloud_computer_auto_boot/pkg/api/types"
	"ecloud_computer_auto_boot/pkg/service/auth"
	"ecloud_computer_auto_boot/pkg/service/logger"
	"net/http"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authManager *auth.Manager
	logManager  *logger.Manager
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authManager *auth.Manager, logManager *logger.Manager) *AuthHandler {
	return &AuthHandler{
		authManager: authManager,
		logManager:  logManager,
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
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

	// 验证必填字段
	if req.Username == "" || req.Password == "" {
		types.RespondJSON(w, http.StatusBadRequest, types.Response{
			Success: false,
			Message: "用户名和密码不能为空",
		})
		return
	}

	// 验证用户名密码
	token, userInfo, err := h.authManager.Login(req.Username, req.Password)
	if err != nil {
		// 记录失败日志
		h.logManager.Log(logger.Event{
			Type:    logger.EventLoginFailed,
			Message: "用户登录失败: " + req.Username,
			Status:  "failed",
			Details: map[string]interface{}{
				"username": req.Username,
				"error":    err.Error(),
			},
		})

		types.RespondJSON(w, http.StatusUnauthorized, types.Response{
			Success: false,
			Message: "用户名或密码错误",
		})
		return
	}

	// 记录成功日志
	h.logManager.Log(logger.Event{
		Type:    logger.EventLogin,
		UserID:  userInfo.ID,
		Message: "用户登录成功: " + req.Username,
		Status:  "success",
		Details: map[string]interface{}{
			"username": req.Username,
		},
	})

	types.RespondJSON(w, http.StatusOK, types.Response{
		Success: true,
		Data: LoginResponse{
			Token: token,
			User:  userInfo.ToSafeUser(),
		},
	})
}

// VerifyToken 验证 Token
func (h *AuthHandler) VerifyToken(w http.ResponseWriter, r *http.Request) {
	// 如果能到达这里，说明中间件已验证通过
	userID := types.GetUserIDFromContext(r)
	username := types.GetUsernameFromContext(r)
	role := types.GetUserRoleFromContext(r)

	types.RespondJSON(w, http.StatusOK, types.Response{
		Success: true,
		Data: map[string]interface{}{
			"user_id":  userID,
			"username": username,
			"role":     role,
		},
		Message: "Token 有效",
	})
}
