package logger

import (
	"encoding/json"
	"ecloud_computer_auto_boot/pkg/store"
	"math/rand"
	"sync"
	"time"
)

// EventType 事件类型
type EventType string

const (
	EventBoot          EventType = "boot"           // 开机
	EventBootSuccess   EventType = "boot_success"   // 开机成功
	EventBootFailed    EventType = "boot_failed"    // 开机失败
	EventLogin         EventType = "login"          // 登录
	EventLoginFailed   EventType = "login_failed"   // 登录失败
	EventConfigChange  EventType = "config_change"  // 配置变更
	EventAccountAdd    EventType = "account_add"    // 添加账号
	EventAccountUpdate EventType = "account_update" // 更新账号
	EventAccountDelete EventType = "account_delete" // 删除账号
	EventTaskStart     EventType = "task_start"     // 任务启动
	EventTaskStop      EventType = "task_stop"      // 任务停止
)

// Event 操作日志事件
type Event struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Type      EventType              `json:"type"`
	AccountID string                 `json:"account_id,omitempty"`
	UserID    string                 `json:"user_id,omitempty"` // 操作者用户 ID
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Status    string                 `json:"status"` // success, failed, info
}

// Manager 日志管理器
type Manager struct {
	mu        sync.RWMutex
	eventChan chan Event
	stopChan  chan struct{}
}

// NewManager 创建日志管理器
func NewManager() *Manager {
	return &Manager{
		eventChan: make(chan Event, 100),
		stopChan:  make(chan struct{}),
	}
}

// Start 启动日志服务
func (m *Manager) Start() {
	go m.processEvents()
}

// Stop 停止日志服务
func (m *Manager) Stop() {
	close(m.stopChan)
}

// Log 记录日志
func (m *Manager) Log(event Event) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.ID == "" {
		event.ID = generateEventID()
	}

	select {
	case m.eventChan <- event:
	default:
		// 缓冲区满，丢弃旧事件
	}
}

// processEvents 处理事件队列
func (m *Manager) processEvents() {
	for {
		select {
		case event := <-m.eventChan:
			m.saveEvent(event)
		case <-m.stopChan:
			return
		}
	}
}

// saveEvent 保存事件到 SQLite
func (m *Manager) saveEvent(event Event) {
	m.mu.Lock()
	defer m.mu.Unlock()

	details := store.JSONString(event.Details)
	ts := event.Timestamp.Format("2006-01-02 15:04:05")

	_, _ = store.DB().Exec(`INSERT INTO logs (id, timestamp, type, account_id, user_id, message, details, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		event.ID, ts, string(event.Type), event.AccountID, event.UserID, event.Message, details, event.Status)
}

// GetLogs 获取日志列表
func (m *Manager) GetLogs(page, limit int, accountID string, eventType EventType) ([]Event, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 构建 WHERE 条件
	where := "1=1"
	args := []interface{}{}
	if accountID != "" {
		where += " AND account_id = ?"
		args = append(args, accountID)
	}
	if eventType != "" {
		where += " AND type = ?"
		args = append(args, string(eventType))
	}

	// 查询总数
	var total int
	countQuery := "SELECT COUNT(*) FROM logs WHERE " + where
	err := store.DB().QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	if limit <= 0 {
		limit = 50
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	query := `SELECT id, timestamp, type, account_id, user_id, message, details, status
		FROM logs WHERE ` + where + ` ORDER BY timestamp DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := store.DB().Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	events := make([]Event, 0)
	for rows.Next() {
		var event Event
		var ts, eventTypeStr, detailsStr string
		err := rows.Scan(&event.ID, &ts, &eventTypeStr, &event.AccountID, &event.UserID, &event.Message, &detailsStr, &event.Status)
		if err != nil {
			return nil, 0, err
		}
		event.Type = EventType(eventTypeStr)
		event.Timestamp = parseTime(ts)
		parseDetails(detailsStr, &event.Details)
		events = append(events, event)
	}

	if events == nil {
		events = []Event{}
	}

	return events, total, rows.Err()
}

// GetLogsByUser 获取指定用户的日志列表
func (m *Manager) GetLogsByUser(page, limit int, userID string, accountID string, eventType EventType) ([]Event, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 构建 WHERE 条件
	where := "user_id = ?"
	args := []interface{}{userID}

	if accountID != "" {
		where += " AND account_id = ?"
		args = append(args, accountID)
	}
	if eventType != "" {
		where += " AND type = ?"
		args = append(args, string(eventType))
	}

	// 查询总数
	var total int
	countQuery := "SELECT COUNT(*) FROM logs WHERE " + where
	err := store.DB().QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	if limit <= 0 {
		limit = 50
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	query := `SELECT id, timestamp, type, account_id, user_id, message, details, status
		FROM logs WHERE ` + where + ` ORDER BY timestamp DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := store.DB().Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	events := make([]Event, 0)
	for rows.Next() {
		var event Event
		var ts, eventTypeStr, detailsStr string
		err := rows.Scan(&event.ID, &ts, &eventTypeStr, &event.AccountID, &event.UserID, &event.Message, &detailsStr, &event.Status)
		if err != nil {
			return nil, 0, err
		}
		event.Type = EventType(eventTypeStr)
		event.Timestamp = parseTime(ts)
		parseDetails(detailsStr, &event.Details)
		events = append(events, event)
	}

	if events == nil {
		events = []Event{}
	}

	return events, total, rows.Err()
}

// parseDetails 解析 JSON details 字段
func parseDetails(s string, v *map[string]interface{}) {
	if s == "" || s == "{}" {
		*v = map[string]interface{}{}
		return
	}
	if err := json.Unmarshal([]byte(s), v); err != nil {
		*v = map[string]interface{}{}
	}
}

// generateEventID 生成事件ID
func generateEventID() string {
	return time.Now().Format("20060102150405") + randomString(6)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// parseTime 解析 SQLite 时间字符串
func parseTime(s string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		return time.Now()
	}
	return t
}
