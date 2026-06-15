package monitor

import (
	"context"
	"ecloud_computer_auto_boot/pkg/ecloud"
	"ecloud_computer_auto_boot/pkg/service/account"
	"ecloud_computer_auto_boot/pkg/service/logger"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"gitlab.ecloud.com/ecloud/ecloudsdkcomputer"
	"gitlab.ecloud.com/ecloud/ecloudsdkcore/config"
)

// TaskStatus 任务状态
type TaskStatus struct {
	AccountID   string    `json:"account_id"`
	AccountName string    `json:"account_name"`
	Status      string    `json:"status"` // running, stopped, error
	LastCheck   time.Time `json:"last_check"`
	LastEvent   string    `json:"last_event,omitempty"`
	MachineCount int      `json:"machine_count"`
}

// MonitorTask 单个监控任务
type MonitorTask struct {
	AccountID     string
	AccountName   string
	AccountType   string
	PublicClient  *ecloud.Client
	APIClient     *ecloudsdkcomputer.Client
	Cron          *cron.Cron
	MachineIDs    []string
	Interval      int
	Status        string
	LastCheck     time.Time
	LastEvent     string
	ctx           context.Context
	cancel        context.CancelFunc
	logManager    *logger.Manager
}

// Manager 监控管理器
type Manager struct {
	tasks      map[string]*MonitorTask
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	logManager *logger.Manager
	accManager *account.Manager
}

// NewManager 创建监控管理器
func NewManager(logManager *logger.Manager, accManager *account.Manager) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		tasks:      make(map[string]*MonitorTask),
		ctx:        ctx,
		cancel:     cancel,
		logManager: logManager,
		accManager: accManager,
	}
}

// StartTask 启动监控任务
func (m *Manager) StartTask(acc account.Account) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 如果任务已存在，先停止
	if task, exists := m.tasks[acc.ID]; exists {
		task.Stop()
		delete(m.tasks, acc.ID)
	}

	// 如果未启用监控，直接返回
	if !acc.MonitorConfig.Enabled {
		return nil
	}

	// 创建任务上下文
	ctx, cancel := context.WithCancel(m.ctx)

	task := &MonitorTask{
		AccountID:   acc.ID,
		AccountName: acc.Name,
		AccountType: acc.Type,
		MachineIDs:  acc.MonitorConfig.Machines,
		Interval:    acc.MonitorConfig.Interval,
		Status:      "starting",
		ctx:         ctx,
		cancel:      cancel,
		logManager:  m.logManager,
	}

	// 初始化客户端
	if err := task.initClient(acc); err != nil {
		m.logManager.Log(logger.Event{
			Type:      logger.EventTaskStart,
			AccountID: acc.ID,
			Message:   fmt.Sprintf("任务启动失败: %s", err.Error()),
			Status:    "failed",
		})
		return err
	}

	// 启动 cron
	if err := task.startCron(); err != nil {
		m.logManager.Log(logger.Event{
			Type:      logger.EventTaskStart,
			AccountID: acc.ID,
			Message:   fmt.Sprintf("定时任务启动失败: %s", err.Error()),
			Status:    "failed",
		})
		return err
	}

	task.Status = "running"
	m.tasks[acc.ID] = task

	m.logManager.Log(logger.Event{
		Type:      logger.EventTaskStart,
		AccountID: acc.ID,
		Message:   fmt.Sprintf("监控任务已启动，间隔 %d 秒", acc.MonitorConfig.Interval),
		Status:    "success",
	})

	return nil
}

// StopTask 停止监控任务
func (m *Manager) StopTask(accountID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[accountID]
	if !exists {
		return fmt.Errorf("任务不存在")
	}

	task.Stop()
	delete(m.tasks, accountID)

	m.logManager.Log(logger.Event{
		Type:      logger.EventTaskStop,
		AccountID: accountID,
		Message:   "监控任务已停止",
		Status:    "success",
	})

	return nil
}

// ReloadTask 重新加载任务
func (m *Manager) ReloadTask(accountID string) error {
	acc, err := m.accManager.GetAccount(accountID)
	if err != nil {
		return err
	}

	// 停止旧任务
	_ = m.StopTask(accountID)

	// 启动新任务
	return m.StartTask(*acc)
}

// GetStatus 获取单个任务状态
func (m *Manager) GetStatus(accountID string) (*TaskStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, exists := m.tasks[accountID]
	if !exists {
		return nil, fmt.Errorf("任务不存在")
	}

	return &TaskStatus{
		AccountID:    task.AccountID,
		AccountName:  task.AccountName,
		Status:       task.Status,
		LastCheck:    task.LastCheck,
		LastEvent:    task.LastEvent,
		MachineCount: len(task.MachineIDs),
	}, nil
}

// GetAllStatus 获取所有任务状态
func (m *Manager) GetAllStatus() []TaskStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statuses := make([]TaskStatus, 0, len(m.tasks))
	for _, task := range m.tasks {
		statuses = append(statuses, TaskStatus{
			AccountID:    task.AccountID,
			AccountName:  task.AccountName,
			Status:       task.Status,
			LastCheck:    task.LastCheck,
			LastEvent:    task.LastEvent,
			MachineCount: len(task.MachineIDs),
		})
	}

	return statuses
}

// ReloadAll 重新加载所有任务
func (m *Manager) ReloadAll() error {
	accounts, err := m.accManager.LoadAccounts()
	if err != nil {
		return err
	}

	// 停止所有任务
	m.mu.Lock()
	for _, task := range m.tasks {
		task.Stop()
	}
	m.tasks = make(map[string]*MonitorTask)
	m.mu.Unlock()

	// 启动启用的任务
	for _, acc := range accounts {
		if acc.MonitorConfig.Enabled {
			_ = m.StartTask(acc)
		}
	}

	return nil
}

// Shutdown 关闭管理器
func (m *Manager) Shutdown() {
	m.cancel()

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, task := range m.tasks {
		task.Stop()
	}
}

// initClient 初始化客户端
func (t *MonitorTask) initClient(acc account.Account) error {
	if acc.Type == "public" {
		// 解密凭证
		username, password, _, _, err := (&account.Manager{}).GetAccountCredentials(acc.ID)
		if err != nil {
			return err
		}

		client, err := ecloud.NewClient(username, password)
		if err != nil {
			return err
		}

		// 登录
		if _, err := client.Login(); err != nil {
			return fmt.Errorf("登录失败: %w", err)
		}

		if !client.HasTrustDeviceRecord() {
			return fmt.Errorf("设备未受信任")
		}

		if _, err := client.VerifyAccessTicket(); err != nil {
			return fmt.Errorf("验证票据失败: %w", err)
		}

		if _, err := client.RecordDeviceInfo(); err != nil {
			return fmt.Errorf("记录设备信息失败: %w", err)
		}

		t.PublicClient = client
	} else {
		// 政企版
		_, _, accessKey, secretKey, err := (&account.Manager{}).GetAccountCredentials(acc.ID)
		if err != nil {
			return err
		}

		client := ecloudsdkcomputer.NewClient(&config.Config{
			AccessKey: &accessKey,
			SecretKey: &secretKey,
			PoolId:    &acc.PoolID,
		})

		t.APIClient = client
	}

	return nil
}

// startCron 启动定时任务
func (t *MonitorTask) startCron() error {
	c := cron.New(cron.WithSeconds())
	spec := fmt.Sprintf("@every %ds", t.Interval)

	_, err := c.AddFunc(spec, func() {
		t.check()
	})
	if err != nil {
		return err
	}

	c.Start()
	t.Cron = c

	// 立即执行一次
	go t.check()

	return nil
}

// Stop 停止任务
func (t *MonitorTask) Stop() {
	if t.Cron != nil {
		t.Cron.Stop()
	}
	if t.cancel != nil {
		t.cancel()
	}
}

// check 执行检查（将在后续实现具体逻辑）
func (t *MonitorTask) check() {
	t.LastCheck = time.Now()
	// TODO: 实现具体的检查逻辑（下一步补充）
}
