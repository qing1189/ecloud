package handlers

import (
	"ecloud_computer_auto_boot/pkg/api/types"
	"ecloud_computer_auto_boot/pkg/service/auth"
	"ecloud_computer_auto_boot/pkg/service/logger"
	"ecloud_computer_auto_boot/pkg/service/user"
	"encoding/json"
	"net/http"
	"strings"
)

// UserHandler 用户管理 Handler
type UserHandler struct {
	userManager *user.Manager
	authManager *auth.Manager
	logManager  *logger.Manager
}

// NewUserHandler 创建用户 Handler
func NewUserHandler(userManager *user.Manager, authManager *auth.Manager, logManager *logger.Manager) *UserHandler {
	return &UserHandler{
		userManager: userManager,
		authManager: authManager,
		logManager:  logManager,
	}
}

// ListUsers 列出所有用户（仅管理员）
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// 检查权限
	if !types.IsAdmin(r) {
		types.RespondJSON(w, http.StatusForbidden, types.Response{
			Success: false,
			Message: "权限不足",
		})
		return
	}

	users, err := h.userManager.ListUsers()
	if err != nil {
		types.RespondJSON(w, http.StatusInternalServerError, types.Response{
			Success: false,
			Message: "获取用户列表失败: " + err.Error(),
		})
		return
	}

	types.RespondJSON(w, http.StatusOK, types.Response{
		Success: true,
		Data:    users,
	})
}

// GetCurrentUser 获取当前用户信息
func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID := types.GetUserIDFromContext(r)
	if userID == "" {
		types.RespondJSON(w, http.StatusUnauthorized, types.Response{
			Success: false,
			Message: "未登录",
		})
		return
	}

	u, err := h.userManager.GetUserByID(userID)
	if err != nil {
		types.RespondJSON(w, http.StatusInternalServerError, types.Response{
			Success: false,
			Message: "获取用户信息失败: " + err.Error(),
		})
		return
	}

	types.RespondJSON(w, http.StatusOK, types.Response{
		Success: true,
		Data:    u.ToSafeUser(),
	})
}

// GetUser 获取指定用户信息
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	// 从 URL 路径提取用户 ID
	userID := strings.TrimPrefix(r.URL.Path, "/api/users/")
	if userID == "" {
		types.RespondJSON(w, http.StatusBadRequest, types.Response{
			Success: false,
			Message: "用户 ID 不能为空",
		})
		return
	}

	// 权限检查：管理员或本人
	currentUserID := types.GetUserIDFromContext(r)
	if !types.IsAdmin(r) && currentUserID != userID {
		types.RespondJSON(w, http.StatusForbidden, types.Response{
			Success: false,
			Message: "权限不足",
		})
		return
	}

	u, err := h.userManager.GetUserByID(userID)
	if err != nil {
		if err == user.ErrUserNotFound {
			types.RespondJSON(w, http.StatusNotFound, types.Response{
				Success: false,
				Message: "用户不存在",
			})
			return
		}
		types.RespondJSON(w, http.StatusInternalServerError, types.Response{
			Success: false,
			Message: "获取用户信息失败: " + err.Error(),
		})
		return
	}

	types.RespondJSON(w, http.StatusOK, types.Response{
		Success: true,
		Data:    u.ToSafeUser(),
	})
}

// CreateUser 创建用户（仅管理员）
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	// 检查权限
	if !types.IsAdmin(r) {
		types.RespondJSON(w, http.StatusForbidden, types.Response{
			Success: false,
			Message: "权限不足",
		})
		return
	}

	var req struct {
		Username    string `json:"username"`
		Password    string `json:"password"`
		Role        string `json:"role"`
		DisplayName string `json:"display_name"`
		Email       string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		types.RespondJSON(w, http.StatusBadRequest, types.Response{
			Success: false,
			Message: "请求参数错误",
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

	// 验证密码长度
	if len(req.Password) < 8 {
		types.RespondJSON(w, http.StatusBadRequest, types.Response{
			Success: false,
			Message: "密码至少8位",
		})
		return
	}

	// 默认角色为 user
	if req.Role == "" {
		req.Role = "user"
	}

	// 哈希密码
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		types.RespondJSON(w, http.StatusInternalServerError, types.Response{
			Success: false,
			Message: "密码加密失败: " + err.Error(),
		})
		return
	}

	// 创建用户
	u, err := h.userManager.CreateUser(req.Username, passwordHash, req.Role, req.DisplayName, req.Email)
	if err != nil {
		if err == user.ErrUserExists {
			types.RespondJSON(w, http.StatusConflict, types.Response{
				Success: false,
				Message: "用户名已存在",
			})
			return
		}
		if err == user.ErrInvalidUsername {
			types.RespondJSON(w, http.StatusBadRequest, types.Response{
				Success: false,
				Message: "用户名格式不正确（3-32位字母数字下划线）",
			})
			return
		}
		if err == user.ErrInvalidRole {
			types.RespondJSON(w, http.StatusBadRequest, types.Response{
				Success: false,
				Message: "角色必须为 admin 或 user",
			})
			return
		}
		types.RespondJSON(w, http.StatusInternalServerError, types.Response{
			Success: false,
			Message: "创建用户失败: " + err.Error(),
		})
		return
	}

	// 记录日志
	h.logManager.Log(logger.Event{
		Type:    "user_create",
		UserID:  types.GetUserIDFromContext(r),
		Message: "创建用户: " + req.Username,
		Status:  "success",
		Details: map[string]interface{}{
			"username": req.Username,
			"role":     req.Role,
		},
	})

	types.RespondJSON(w, http.StatusCreated, types.Response{
		Success: true,
		Data:    u.ToSafeUser(),
		Message: "创建用户成功",
	})
}

// UpdateUser 更新用户信息
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// 从 URL 路径提取用户 ID
	userID := strings.TrimPrefix(r.URL.Path, "/api/users/")
	if userID == "" {
		types.RespondJSON(w, http.StatusBadRequest, types.Response{
			Success: false,
			Message: "用户 ID 不能为空",
		})
		return
	}

	// 权限检查：管理员或本人
	currentUserID := types.GetUserIDFromContext(r)
	if !types.IsAdmin(r) && currentUserID != userID {
		types.RespondJSON(w, http.StatusForbidden, types.Response{
			Success: false,
			Message: "权限不足",
		})
		return
	}

	var req struct {
		DisplayName string `json:"display_name"`
		Email       string `json:"email"`
		Role        string `json:"role"` // 仅管理员可修改
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		types.RespondJSON(w, http.StatusBadRequest, types.Response{
			Success: false,
			Message: "请求参数错误",
		})
		return
	}

	err := h.userManager.UpdateUser(userID, func(u *user.User) error {
		if req.DisplayName != "" {
			u.DisplayName = req.DisplayName
		}
		if req.Email != "" {
			u.Email = req.Email
		}
		// 仅管理员可修改角色
		if req.Role != "" && types.IsAdmin(r) {
			if req.Role != "admin" && req.Role != "user" {
				return user.ErrInvalidRole
			}
			u.Role = req.Role
		}
		return nil
	})

	if err != nil {
		if err == user.ErrUserNotFound {
			types.RespondJSON(w, http.StatusNotFound, types.Response{
				Success: false,
				Message: "用户不存在",
			})
			return
		}
		types.RespondJSON(w, http.StatusInternalServerError, types.Response{
			Success: false,
			Message: "更新用户失败: " + err.Error(),
		})
		return
	}

	// 记录日志
	h.logManager.Log(logger.Event{
		Type:    "user_update",
		UserID:  currentUserID,
		Message: "更新用户信息: " + userID,
		Status:  "success",
	})

	types.RespondJSON(w, http.StatusOK, types.Response{
		Success: true,
		Message: "更新用户成功",
	})
}

// DeleteUser 删除用户（仅管理员）
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	// 检查权限
	if !types.IsAdmin(r) {
		types.RespondJSON(w, http.StatusForbidden, types.Response{
			Success: false,
			Message: "权限不足",
		})
		return
	}

	// 从 URL 路径提取用户 ID
	userID := strings.TrimPrefix(r.URL.Path, "/api/users/")
	if userID == "" {
		types.RespondJSON(w, http.StatusBadRequest, types.Response{
			Success: false,
			Message: "用户 ID 不能为空",
		})
		return
	}

	// 不能删除自己
	currentUserID := types.GetUserIDFromContext(r)
	if currentUserID == userID {
		types.RespondJSON(w, http.StatusBadRequest, types.Response{
			Success: false,
			Message: "不能删除自己",
		})
		return
	}

	// 删除用户（会级联删除其云账号）
	err := h.userManager.DeleteUser(userID)
	if err != nil {
		if err == user.ErrUserNotFound {
			types.RespondJSON(w, http.StatusNotFound, types.Response{
				Success: false,
				Message: "用户不存在",
			})
			return
		}
		if err == user.ErrCannotDeleteAdmin {
			types.RespondJSON(w, http.StatusBadRequest, types.Response{
				Success: false,
				Message: "不能删除最后一个管理员账号",
			})
			return
		}
		types.RespondJSON(w, http.StatusInternalServerError, types.Response{
			Success: false,
			Message: "删除用户失败: " + err.Error(),
		})
		return
	}

	// 记录日志
	h.logManager.Log(logger.Event{
		Type:    "user_delete",
		UserID:  currentUserID,
		Message: "删除用户: " + userID,
		Status:  "success",
	})

	types.RespondJSON(w, http.StatusOK, types.Response{
		Success: true,
		Message: "删除用户成功",
	})
}

// ChangePassword 修改用户密码
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	// 从 URL 路径提取用户 ID
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		types.RespondJSON(w, http.StatusBadRequest, types.Response{
			Success: false,
			Message: "无效的请求路径",
		})
		return
	}
	userID := pathParts[3]

	// 权限检查：管理员或本人
	currentUserID := types.GetUserIDFromContext(r)
	isAdmin := types.IsAdmin(r)
	if !isAdmin && currentUserID != userID {
		types.RespondJSON(w, http.StatusForbidden, types.Response{
			Success: false,
			Message: "权限不足",
		})
		return
	}

	var req struct {
		OldPassword string `json:"old_password"` // 非管理员必填
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		types.RespondJSON(w, http.StatusBadRequest, types.Response{
			Success: false,
			Message: "请求参数错误",
		})
		return
	}

	// 验证新密码
	if req.NewPassword == "" || len(req.NewPassword) < 8 {
		types.RespondJSON(w, http.StatusBadRequest, types.Response{
			Success: false,
			Message: "新密码至少8位",
		})
		return
	}

	var err error
	if isAdmin && currentUserID != userID {
		// 管理员重置其他用户密码，无需验证旧密码
		err = h.authManager.ResetPassword(userID, req.NewPassword)
	} else {
		// 本人修改密码，需要验证旧密码
		if req.OldPassword == "" {
			types.RespondJSON(w, http.StatusBadRequest, types.Response{
				Success: false,
				Message: "旧密码不能为空",
			})
			return
		}
		err = h.authManager.ChangePassword(userID, req.OldPassword, req.NewPassword)
	}

	if err != nil {
		if err == auth.ErrInvalidPassword {
			types.RespondJSON(w, http.StatusBadRequest, types.Response{
				Success: false,
				Message: "旧密码错误",
			})
			return
		}
		if err == user.ErrUserNotFound {
			types.RespondJSON(w, http.StatusNotFound, types.Response{
				Success: false,
				Message: "用户不存在",
			})
			return
		}
		types.RespondJSON(w, http.StatusInternalServerError, types.Response{
			Success: false,
			Message: "修改密码失败: " + err.Error(),
		})
		return
	}

	// 记录日志
	h.logManager.Log(logger.Event{
		Type:    "password_change",
		UserID:  currentUserID,
		Message: "修改用户密码: " + userID,
		Status:  "success",
	})

	types.RespondJSON(w, http.StatusOK, types.Response{
		Success: true,
		Message: "修改密码成功",
	})
}
