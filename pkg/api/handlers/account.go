package handlers

import (
	"ecloud_computer_auto_boot/pkg/api"
	"ecloud_computer_auto_boot/pkg/ecloud"
	"ecloud_computer_auto_boot/pkg/service/account"
	"ecloud_computer_auto_boot/pkg/service/logger"
	"ecloud_computer_auto_boot/pkg/service/monitor"
	"net/http"
	"strings"
)

// AccountHandler 账号处理器
type AccountHandler struct {
	accManager     *account.Manager
	monitorManager *monitor.Manager
	logManager     *logger.Manager
}

// NewAccountHandler 创建账号处理器
func NewAccountHandler(accManager *account.Manager, monitorManager *monitor.Manager, logManager *logger.Manager) *AccountHandler {
	return &AccountHandler{
		accManager:     accManager,
		monitorManager: monitorManager,
		logManager:     logManager,
	}
}

// ListAccounts 列出所有账号
func (h *AccountHandler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.accManager.ListAccounts()
	if err != nil {
		api.respondJSON(w, http.StatusInternalServerError, api.Response{
			Success: false,
			Message: "获取账号列表失败",
			Error:   err.Error(),
		})
		return
	}

	api.respondJSON(w, http.StatusOK, api.Response{
		Success: true,
		Data:    accounts,
	})
}

// AddAccountRequest 添加账号请求
type AddAccountRequest struct {
	Name      string                  `json:"name"`
	Type      string                  `json:"type"` // public 或 business
	Username  string                  `json:"username,omitempty"`
	Password  string                  `json:"password,omitempty"`
	AccessKey string                  `json:"access_key,omitempty"`
	SecretKey string                  `json:"secret_key,omitempty"`
	PoolID    string                  `json:"pool_id,omitempty"`
	Config    account.MonitorConfig   `json:"monitor_config"`
}

// AddAccount 添加账号（支持设备信任验证）
func (h *AccountHandler) AddAccount(w http.ResponseWriter, r *http.Request) {
	var req AddAccountRequest
	if err := api.parseJSON(r, &req); err != nil {
		api.respondJSON(w, http.StatusBadRequest, api.Response{
			Success: false,
			Message: "请求格式错误",
		})
		return
	}

	// 验证必填字段
	if req.Name == "" {
		api.respondJSON(w, http.StatusBadRequest, api.Response{
			Success: false,
			Message: "账号名称不能为空",
		})
		return
	}

	if req.Type != "public" && req.Type != "business" {
		api.respondJSON(w, http.StatusBadRequest, api.Response{
			Success: false,
			Message: "账号类型必须为 public 或 business",
		})
		return
	}

	// 公众版需要用户名密码
	if req.Type == "public" && (req.Username == "" || req.Password == "") {
		api.respondJSON(w, http.StatusBadRequest, api.Response{
			Success: false,
			Message: "公众版账号需要提供用户名和密码",
		})
		return
	}

	// 政企版需要密钥
	if req.Type == "business" && (req.AccessKey == "" || req.SecretKey == "") {
		api.respondJSON(w, http.StatusBadRequest, api.Response{
			Success: false,
			Message: "政企版账号需要提供 AccessKey 和 SecretKey",
		})
		return
	}

	// 创建账号对象
	acc := account.Account{
		Name:          req.Name,
		Type:          req.Type,
		Username:      req.Username,
		Password:      req.Password,
		AccessKey:     req.AccessKey,
		SecretKey:     req.SecretKey,
		PoolID:        req.PoolID,
		MonitorConfig: req.Config,
	}

	// 公众版需要检查设备信任
	if req.Type == "public" {
		// 创建客户端
		client, err := ecloud.NewClient(req.Username, req.Password)
		if err != nil {
			api.respondJSON(w, http.StatusBadRequest, api.Response{
				Success: false,
				Message: "创建客户端失败",
				Error:   err.Error(),
			})
			return
		}

		// 尝试登录
		if _, err := client.Login(); err != nil {
			api.respondJSON(w, http.StatusBadRequest, api.Response{
				Success: false,
				Message: "登录失败，请检查用户名和密码",
				Error:   err.Error(),
			})
			return
		}

		// 检查是否需要设备信任
		if !client.HasTrustDeviceRecord() {
			// 发送验证码
			resp, err := client.SendTrustDeviceVerifySms()
			if err != nil {
				api.respondJSON(w, http.StatusInternalServerError, api.Response{
					Success: false,
					Message: "发送验证码失败",
					Error:   err.Error(),
				})
				return
			}

			// 保存临时账号
			tempID := account.SaveTempAccount(client, &acc)

			// 返回需要验证的响应（特殊状态码 10001）
			api.respondJSON(w, http.StatusOK, api.Response{
				Success: false,
				Message: "需要设备验证",
				Data: map[string]interface{}{
					"needVerify": true,
					"tempId":     tempID,
					"mobile":     maskMobile(client.GetSession().Mobile),
					"expireTime": int(resp.Body.(map[string]interface{})["expireTime"].(float64)),
				},
			})
			return
		}
	}

	// 不需要验证或政企版，直接保存
	if err := h.accManager.AddAccount(acc); err != nil {
		api.respondJSON(w, http.StatusInternalServerError, api.Response{
			Success: false,
			Message: "添加账号失败",
			Error:   err.Error(),
		})
		return
	}

	// 记录日志
	h.logManager.Log(logger.Event{
		Type:      logger.EventAccountAdd,
		AccountID: acc.ID,
		Message:   "添加账号: " + acc.Name,
		Status:    "success",
	})

	// 如果启用监控，启动任务
	if acc.MonitorConfig.Enabled {
		_ = h.monitorManager.StartTask(acc)
	}

	api.respondJSON(w, http.StatusOK, api.Response{
		Success: true,
		Message: "账号添加成功",
		Data:    acc.ToSafeAccount(),
	})
}

// GetAccount 获取账号详情
func (h *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	// 从 URL 中提取 ID
	path := r.URL.Path
	id := strings.TrimPrefix(path, "/api/accounts/")

	acc, err := h.accManager.GetAccount(id)
	if err != nil {
		api.respondJSON(w, http.StatusNotFound, api.Response{
			Success: false,
			Message: "账号不存在",
		})
		return
	}

	api.respondJSON(w, http.StatusOK, api.Response{
		Success: true,
		Data:    acc.ToSafeAccount(),
	})
}

// UpdateAccountRequest 更新账号请求
type UpdateAccountRequest struct {
	Name      *string                `json:"name,omitempty"`
	Password  *string                `json:"password,omitempty"`
	SecretKey *string                `json:"secret_key,omitempty"`
	Config    *account.MonitorConfig `json:"monitor_config,omitempty"`
}

// UpdateAccount 更新账号
func (h *AccountHandler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	// 从 URL 中提取 ID
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		api.respondJSON(w, http.StatusBadRequest, api.Response{
			Success: false,
			Message: "无效的请求路径",
		})
		return
	}
	id := parts[3]

	var req UpdateAccountRequest
	if err := api.parseJSON(r, &req); err != nil {
		api.respondJSON(w, http.StatusBadRequest, api.Response{
			Success: false,
			Message: "请求格式错误",
		})
		return
	}

	// 更新账号
	err := h.accManager.UpdateAccount(id, func(acc *account.Account) error {
		if req.Name != nil {
			acc.Name = *req.Name
		}
		if req.Password != nil && *req.Password != "" {
			acc.Password = *req.Password
		}
		if req.SecretKey != nil && *req.SecretKey != "" {
			acc.SecretKey = *req.SecretKey
		}
		if req.Config != nil {
			acc.MonitorConfig = *req.Config
		}
		return nil
	})

	if err != nil {
		api.respondJSON(w, http.StatusInternalServerError, api.Response{
			Success: false,
			Message: "更新账号失败",
			Error:   err.Error(),
		})
		return
	}

	// 记录日志
	h.logManager.Log(logger.Event{
		Type:      logger.EventAccountUpdate,
		AccountID: id,
		Message:   "更新账号配置",
		Status:    "success",
	})

	// 重新加载监控任务
	_ = h.monitorManager.ReloadTask(id)

	api.respondJSON(w, http.StatusOK, api.Response{
		Success: true,
		Message: "账号更新成功",
	})
}

// DeleteAccount 删除账号
func (h *AccountHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	// 从 URL 中提取 ID
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		api.respondJSON(w, http.StatusBadRequest, api.Response{
			Success: false,
			Message: "无效的请求路径",
		})
		return
	}
	id := parts[3]

	// 停止监控任务
	_ = h.monitorManager.StopTask(id)

	// 删除账号
	if err := h.accManager.DeleteAccount(id); err != nil {
		api.respondJSON(w, http.StatusInternalServerError, api.Response{
			Success: false,
			Message: "删除账号失败",
			Error:   err.Error(),
		})
		return
	}

	// 记录日志
	h.logManager.Log(logger.Event{
		Type:      logger.EventAccountDelete,
		AccountID: id,
		Message:   "删除账号",
		Status:    "success",
	})

	api.respondJSON(w, http.StatusOK, api.Response{
		Success: true,
		Message: "账号删除成功",
	})
}

// ToggleMonitor 启用/停用监控
func (h *AccountHandler) ToggleMonitor(w http.ResponseWriter, r *http.Request) {
	// 从 URL 中提取 ID
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		api.respondJSON(w, http.StatusBadRequest, api.Response{
			Success: false,
			Message: "无效的请求路径",
		})
		return
	}
	id := parts[3]

	acc, err := h.accManager.GetAccount(id)
	if err != nil {
		api.respondJSON(w, http.StatusNotFound, api.Response{
			Success: false,
			Message: "账号不存在",
		})
		return
	}

	// 切换状态
	newEnabled := !acc.MonitorConfig.Enabled

	err = h.accManager.UpdateAccount(id, func(acc *account.Account) error {
		acc.MonitorConfig.Enabled = newEnabled
		return nil
	})

	if err != nil {
		api.respondJSON(w, http.StatusInternalServerError, api.Response{
			Success: false,
			Message: "切换监控状态失败",
			Error:   err.Error(),
		})
		return
	}

	// 启动或停止任务
	if newEnabled {
		acc.MonitorConfig.Enabled = true
		_ = h.monitorManager.StartTask(*acc)
	} else {
		_ = h.monitorManager.StopTask(id)
	}

	api.respondJSON(w, http.StatusOK, api.Response{
		Success: true,
		Message: "监控状态已更新",
		Data: map[string]bool{
			"enabled": newEnabled,
		},
	})
}

// VerifyDeviceRequest 验证设备请求
type VerifyDeviceRequest struct {
	TempID string `json:"tempId"`
	Code   string `json:"code"`
}

// VerifyDevice 提交验证码，完成设备信任
func (h *AccountHandler) VerifyDevice(w http.ResponseWriter, r *http.Request) {
	var req VerifyDeviceRequest
	if err := api.parseJSON(r, &req); err != nil {
		api.respondJSON(w, http.StatusBadRequest, api.Response{
			Success: false,
			Message: "请求格式错误",
		})
		return
	}

	if req.TempID == "" || req.Code == "" {
		api.respondJSON(w, http.StatusBadRequest, api.Response{
			Success: false,
			Message: "临时ID和验证码不能为空",
		})
		return
	}

	// 获取临时账号
	temp, err := account.GetTempAccount(req.TempID)
	if err != nil {
		api.respondJSON(w, http.StatusBadRequest, api.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// 提交验证码
	resp, err := temp.Client.TrustDevice(req.Code)
	if err != nil {
		api.respondJSON(w, http.StatusInternalServerError, api.Response{
			Success: false,
			Message: "验证失败",
			Error:   err.Error(),
		})
		return
	}

	if !resp.Success() {
		api.respondJSON(w, http.StatusBadRequest, api.Response{
			Success: false,
			Message: "验证码错误或已过期",
			Error:   resp.ErrorMessage,
		})
		return
	}

	// 验证成功，保存账号
	if err := h.accManager.AddAccount(*temp.Account); err != nil {
		api.respondJSON(w, http.StatusInternalServerError, api.Response{
			Success: false,
			Message: "保存账号失败",
			Error:   err.Error(),
		})
		return
	}

	// 清理临时账号
	account.DeleteTempAccount(req.TempID)

	// 记录日志
	h.logManager.Log(logger.Event{
		Type:      logger.EventAccountAdd,
		AccountID: temp.Account.ID,
		Message:   "添加账号（已验证设备）: " + temp.Account.Name,
		Status:    "success",
	})

	// 如果启用监控，启动任务
	if temp.Account.MonitorConfig.Enabled {
		_ = h.monitorManager.StartTask(*temp.Account)
	}

	api.respondJSON(w, http.StatusOK, api.Response{
		Success: true,
		Message: "验证成功，账号添加完成",
		Data:    temp.Account.ToSafeAccount(),
	})
}

// ResendCodeRequest 重新发送验证码请求
type ResendCodeRequest struct {
	TempID string `json:"tempId"`
}

// ResendCode 重新发送验证码
func (h *AccountHandler) ResendCode(w http.ResponseWriter, r *http.Request) {
	var req ResendCodeRequest
	if err := api.parseJSON(r, &req); err != nil {
		api.respondJSON(w, http.StatusBadRequest, api.Response{
			Success: false,
			Message: "请求格式错误",
		})
		return
	}

	if req.TempID == "" {
		api.respondJSON(w, http.StatusBadRequest, api.Response{
			Success: false,
			Message: "临时ID不能为空",
		})
		return
	}

	// 获取临时账号
	temp, err := account.GetTempAccount(req.TempID)
	if err != nil {
		api.respondJSON(w, http.StatusBadRequest, api.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// 重新发送验证码
	resp, err := temp.Client.SendTrustDeviceVerifySms()
	if err != nil {
		api.respondJSON(w, http.StatusInternalServerError, api.Response{
			Success: false,
			Message: "发送验证码失败",
			Error:   err.Error(),
		})
		return
	}

	api.respondJSON(w, http.StatusOK, api.Response{
		Success: true,
		Message: "验证码已重新发送",
		Data: map[string]interface{}{
			"expireTime": int(resp.Body.(map[string]interface{})["expireTime"].(float64)),
		},
	})
}

// maskMobile 手机号脱敏
func maskMobile(mobile string) string {
	if len(mobile) != 11 {
		return mobile
	}
	return mobile[:3] + "****" + mobile[7:]
}
