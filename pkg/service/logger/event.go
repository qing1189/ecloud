package logger

import (
	"ecloud_computer_auto_boot/pkg/store"
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
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Status    string                 `json:"status"` // success, failed, info
}

// LogStore 日志存储
type LogStore struct {
	Events []Event `json:"events"`
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

// saveEvent 保存事件到文件
func (m *Manager) saveEvent(event Event) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var logStore LogStore
	_ = store.ReadJSON(store.LogsFile, &logStore)

	logStore.Events = append(logStore.Events, event)

	// 保留最近 1000 条
	if len(logStore.Events) > 1000 {
		logStore.Events = logStore.Events[len(logStore.Events)-1000:]
	}

	_ = store.WriteJSON(store.LogsFile, logStore)
}

// GetLogs 获取日志列表
func (m *Manager) GetLogs(page, limit int, accountID string, eventType EventType) ([]Event, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var logStore LogStore
	if err := store.ReadJSON(store.LogsFile, &logStore); err != nil {
		return nil, 0, err
	}

	// 筛选
	filtered := make([]Event, 0)
	for i := len(logStore.Events) - 1; i >= 0; i-- {
		event := logStore.Events[i]

		if accountID != "" && event.AccountID != accountID {
			continue
		}

		if eventType != "" && event.Type != eventType {
			continue
		}

		filtered = append(filtered, event)
	}

	total := len(filtered)

	// 分页
	if limit <= 0 {
		limit = 50
	}
	if page <= 0 {
		page = 1
	}

	start := (page - 1) * limit
	end := start + limit

	if start >= total {
		return []Event{}, total, nil
	}

	if end > total {
		end = total
	}

	return filtered[start:end], total, nil
}

// generateEventID 生成事件ID
func generateEventID() string {
	return time.Now().Format("20060102150405") + randomString(6)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
