package api

import (
	"ecloud_computer_auto_boot/pkg/api/handlers"
	"ecloud_computer_auto_boot/pkg/api/types"
	"ecloud_computer_auto_boot/pkg/service/account"
	"ecloud_computer_auto_boot/pkg/service/auth"
	"ecloud_computer_auto_boot/pkg/service/logger"
	"ecloud_computer_auto_boot/pkg/service/monitor"
	"net/http"
)

// Router 路由配置
type Router struct {
	authHandler    *handlers.AuthHandler
	accountHandler *handlers.AccountHandler
	monitorHandler *handlers.MonitorHandler
	logHandler     *handlers.LogHandler
	authManager    *auth.Manager
}

// NewRouter 创建路由器
func NewRouter(
	authManager *auth.Manager,
	accManager *account.Manager,
	monitorManager *monitor.Manager,
	logManager *logger.Manager,
) *Router {
	return &Router{
		authHandler:    handlers.NewAuthHandler(authManager),
		accountHandler: handlers.NewAccountHandler(accManager, monitorManager, logManager),
		monitorHandler: handlers.NewMonitorHandler(monitorManager),
		logHandler:     handlers.NewLogHandler(logManager),
		authManager:    authManager,
	}
}

// Setup 设置路由
func (router *Router) Setup(mux *http.ServeMux) {
	// 应用全局中间件
	corsMiddleware := CORSMiddleware()
	loggingMiddleware := LoggingMiddleware()
	authMiddleware := AuthMiddleware(router.authManager)

	// 认证接口（无需认证）
	mux.Handle("/api/auth/login", chain(
		http.HandlerFunc(router.authHandler.Login),
		corsMiddleware,
		loggingMiddleware,
	))

	// 需要认证的接口
	mux.Handle("/api/auth/verify", chain(
		http.HandlerFunc(router.authHandler.VerifyToken),
		corsMiddleware,
		loggingMiddleware,
		authMiddleware,
	))

	mux.Handle("/api/auth/change-password", chain(
		http.HandlerFunc(router.authHandler.ChangePassword),
		corsMiddleware,
		loggingMiddleware,
		authMiddleware,
	))

	// 账号管理接口
	mux.Handle("/api/accounts", chain(
		http.HandlerFunc(router.handleAccounts),
		corsMiddleware,
		loggingMiddleware,
		authMiddleware,
	))

	mux.Handle("/api/accounts/", chain(
		http.HandlerFunc(router.handleAccountByID),
		corsMiddleware,
		loggingMiddleware,
		authMiddleware,
	))

	// 设备验证接口（添加账号时的验证）
	mux.Handle("/api/accounts/verify", chain(
		http.HandlerFunc(router.accountHandler.VerifyDevice),
		corsMiddleware,
		loggingMiddleware,
		authMiddleware,
	))

	mux.Handle("/api/accounts/resend-code", chain(
		http.HandlerFunc(router.accountHandler.ResendCode),
		corsMiddleware,
		loggingMiddleware,
		authMiddleware,
	))

	// 监控状态接口
	mux.Handle("/api/monitor/status", chain(
		http.HandlerFunc(router.monitorHandler.GetAllStatus),
		corsMiddleware,
		loggingMiddleware,
		authMiddleware,
	))

	mux.Handle("/api/monitor/status/", chain(
		http.HandlerFunc(router.monitorHandler.GetStatus),
		corsMiddleware,
		loggingMiddleware,
		authMiddleware,
	))

	mux.Handle("/api/monitor/reload", chain(
		http.HandlerFunc(router.monitorHandler.ReloadAll),
		corsMiddleware,
		loggingMiddleware,
		authMiddleware,
	))

	// 日志接口
	mux.Handle("/api/logs", chain(
		http.HandlerFunc(router.logHandler.GetLogs),
		corsMiddleware,
		loggingMiddleware,
		authMiddleware,
	))
}

// handleAccounts 处理 /api/accounts 路由
func (router *Router) handleAccounts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		router.accountHandler.ListAccounts(w, r)
	case http.MethodPost:
		router.accountHandler.AddAccount(w, r)
	default:
		types.RespondJSON(w, http.StatusMethodNotAllowed, types.Response{
			Success: false,
			Message: "方法不允许",
		})
	}
}

// handleAccountByID 处理 /api/accounts/:id 路由
func (router *Router) handleAccountByID(w http.ResponseWriter, r *http.Request) {
	// 检查是否是 toggle 操作
	if len(r.URL.Path) > 14 && r.URL.Path[len(r.URL.Path)-7:] == "/toggle" {
		if r.Method == http.MethodPost {
			router.accountHandler.ToggleMonitor(w, r)
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		router.accountHandler.GetAccount(w, r)
	case http.MethodPut:
		router.accountHandler.UpdateAccount(w, r)
	case http.MethodDelete:
		router.accountHandler.DeleteAccount(w, r)
	default:
		types.RespondJSON(w, http.StatusMethodNotAllowed, types.Response{
			Success: false,
			Message: "方法不允许",
		})
	}
}

// chain 中间件链
func chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
