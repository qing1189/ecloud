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

// GetLogs 获取日志列表（管理员看所有，普通用户仅看自己的）
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
	var events []logger.Event
	var total int
	var err error

	if types.IsAdmin(r) {
		// 管理员：查看所有日志
		events, total, err = h.logManager.GetLogs(page, limit, accountID, eventType)
	} else {
		// 普通用户：仅查看自己的日志
		userID := types.GetUserIDFromContext(r)
		events, total, err = h.logManager.GetLogsByUser(page, limit, userID, accountID, eventType)
	}

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
