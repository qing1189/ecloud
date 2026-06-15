package handlers

import (
	"ecloud_computer_auto_boot/pkg/api"
	"ecloud_computer_auto_boot/pkg/service/monitor"
	"net/http"
	"strings"
)

// MonitorHandler 监控处理器
type MonitorHandler struct {
	monitorManager *monitor.Manager
}

// NewMonitorHandler 创建监控处理器
func NewMonitorHandler(monitorManager *monitor.Manager) *MonitorHandler {
	return &MonitorHandler{
		monitorManager: monitorManager,
	}
}

// GetAllStatus 获取所有任务状态
func (h *MonitorHandler) GetAllStatus(w http.ResponseWriter, r *http.Request) {
	statuses := h.monitorManager.GetAllStatus()

	api.respondJSON(w, http.StatusOK, api.Response{
		Success: true,
		Data:    statuses,
	})
}

// GetStatus 获取单个任务状态
func (h *MonitorHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	// 从 URL 中提取 ID
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 5 {
		api.respondJSON(w, http.StatusBadRequest, api.Response{
			Success: false,
			Message: "无效的请求路径",
		})
		return
	}
	id := parts[4]

	status, err := h.monitorManager.GetStatus(id)
	if err != nil {
		api.respondJSON(w, http.StatusNotFound, api.Response{
			Success: false,
			Message: "任务不存在",
		})
		return
	}

	api.respondJSON(w, http.StatusOK, api.Response{
		Success: true,
		Data:    status,
	})
}

// ReloadAll 重新加载所有任务
func (h *MonitorHandler) ReloadAll(w http.ResponseWriter, r *http.Request) {
	if err := h.monitorManager.ReloadAll(); err != nil {
		api.respondJSON(w, http.StatusInternalServerError, api.Response{
			Success: false,
			Message: "重新加载任务失败",
			Error:   err.Error(),
		})
		return
	}

	api.respondJSON(w, http.StatusOK, api.Response{
		Success: true,
		Message: "任务已重新加载",
	})
}
