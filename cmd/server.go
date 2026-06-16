package cmd

import (
	"ecloud_computer_auto_boot/bootstrap"
	"ecloud_computer_auto_boot/pkg/api"
	"ecloud_computer_auto_boot/pkg/service/account"
	"ecloud_computer_auto_boot/pkg/service/auth"
	"ecloud_computer_auto_boot/pkg/service/logger"
	"ecloud_computer_auto_boot/pkg/service/monitor"
	"ecloud_computer_auto_boot/pkg/service/user"
	"ecloud_computer_auto_boot/pkg/store"
	"ecloud_computer_auto_boot/pkg/util"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the web server",
	Long:  `Start the web management server with RESTful API and dashboard.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 初始化应用
		bootstrap.InitApplication()

		// 初始化存储
		if err := store.Init("."); err != nil {
			util.Log().Error("存储初始化失败: %s", err)
			os.Exit(1)
		}

		// 初始化用户管理器
		userManager := user.NewManager()

		// 初始化认证管理器
		authManager := auth.NewManager(userManager)
		needCreateAdmin, err := authManager.Init()
		if err != nil {
			util.Log().Error("认证初始化失败: %s", err)
			os.Exit(1)
		}

		// 如果需要创建默认管理员
		if needCreateAdmin {
			username, password, err := authManager.CreateDefaultAdmin()
			if err != nil {
				util.Log().Error("创建默认管理员失败: %s", err)
				os.Exit(1)
			}

			util.Log().Info("=================================================")
			util.Log().Info("首次启动检测到，已创建默认管理员账号:")
			util.Log().Info("")
			util.Log().Info("    用户名: %s", username)
			util.Log().Info("    密码: %s", password)
			util.Log().Info("")
			util.Log().Info("请妥善保存登录信息，登录后可修改密码")
			util.Log().Info("=================================================")
		}

		// 初始化日志管理器
		logManager := logger.NewManager()
		logManager.Start()
		defer logManager.Stop()

		// 初始化账号管理器
		accManager := account.NewManager()

		// 初始化监控管理器
		monitorManager := monitor.NewManager(logManager, accManager)
		defer monitorManager.Shutdown()

		// 加载所有启用的账号并启动监控
		accounts, err := accManager.LoadAccounts()
		if err != nil {
			util.Log().Warning("加载账号列表失败: %s", err)
		} else {
			for _, acc := range accounts {
				if acc.MonitorConfig.Enabled {
					if err := monitorManager.StartTask(acc); err != nil {
						util.Log().Error("启动账号 %s 的监控任务失败: %s", acc.Name, err)
					}
				}
			}
		}

		// 创建路由
		router := api.NewRouter(authManager, userManager, accManager, monitorManager, logManager)
		mux := http.NewServeMux()
		router.Setup(mux)

		// 健康检查
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		// 静态文件服务（前端）
		fs := http.FileServer(http.Dir("static"))
		mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// API 请求不走静态文件
			if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
				http.NotFound(w, r)
				return
			}

			// 检查文件是否存在
			if _, err := os.Stat("static" + r.URL.Path); os.IsNotExist(err) {
				// 文件不存在，返回 index.html（SPA 路由）
				http.ServeFile(w, r, "static/index.html")
				return
			}

			// 文件存在，直接服务
			fs.ServeHTTP(w, r)
		}))

		// 启动服务器
		port := 8088
		addr := fmt.Sprintf(":%d", port)

		util.Log().Info("=================================================")
		util.Log().Info("Web 服务已启动")
		util.Log().Info("访问地址: http://0.0.0.0%s", addr)
		util.Log().Info("API 接口: http://0.0.0.0%s/api/", addr)
		util.Log().Info("=================================================")

		// 启动 HTTP 服务器
		server := &http.Server{
			Addr:    addr,
			Handler: mux,
		}

		// 启动服务器（非阻塞）
		go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				util.Log().Error("服务器启动失败: %s", err)
				os.Exit(1)
			}
		}()

		// 等待中断信号
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
		sig := <-sigChan
		util.Log().Info("收到信号 %s, 开始关闭服务", sig)

		// 优雅关闭
		if err := server.Close(); err != nil {
			util.Log().Error("关闭服务器失败: %s", err)
		}

		util.Log().Info("服务已关闭")
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
