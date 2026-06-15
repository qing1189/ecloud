package handlers

import (
	"ecloud_computer_auto_boot/pkg/api/types"
	"ecloud_computer_auto_boot/pkg/service/logger"
	"net/http"
	"strconv"
)

// LogHandler 日志处理器
type LogHandler struct {
	logManager *logger.Manager
}

// NewLogHandler 创建日志处理器
func NewLogHandler(logManager *logger.Manager) *LogHandler {
	return &LogHandler{
		logManager: logManager,
	}
}

// GetLogs 获取日志列表
func (h *LogHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	// 解析查询参数
	query := r.URL.Query()

	page, _ := strconv.Atoi(query.Get("page"))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit <= 0 {
		limit = 50
	}

	accountID := query.Get("account_id")
	eventType := logger.EventType(query.Get("type"))

	// 获取日志
	events, total, err := h.logManager.GetLogs(page, limit, accountID, eventType)
	if err != nil {
		types.RespondJSON(w, http.StatusInternalServerError, types.Response{
			Success: false,
			Message: "获取日志失败",
			Error:   err.Error(),
		})
		return
	}

	types.RespondJSON(w, http.StatusOK, types.Response{
		Success: true,
		Data: map[string]interface{}{
			"events": events,
			"total":  total,
			"page":   page,
			"limit":  limit,
		},
	})
}
