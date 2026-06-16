package api

import (
	"ecloud_computer_auto_boot/pkg/api/types"
	"ecloud_computer_auto_boot/pkg/service/auth"
	"ecloud_computer_auto_boot/pkg/util"
	"net/http"
	"strings"
	"time"
)

// AuthMiddleware JWT 认证中间件
func AuthMiddleware(authManager *auth.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 从 Header 获取 Token
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				types.RespondJSON(w, http.StatusUnauthorized, types.Response{
					Success: false,
					Message: "未提供认证令牌",
				})
				return
			}

			// 提取 Token (Bearer token)
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				types.RespondJSON(w, http.StatusUnauthorized, types.Response{
					Success: false,
					Message: "无效的认证令牌格式",
				})
				return
			}

			token := parts[1]

			// 验证 Token
			claims, err := authManager.ValidateToken(token)
			if err != nil {
				types.RespondJSON(w, http.StatusUnauthorized, types.Response{
					Success: false,
					Message: "认证令牌无效或已过期",
				})
				return
			}

			// 将用户信息存入 Context
			ctx := types.WithUserContext(r.Context(), claims.UserID, claims.Username, claims.Role)

			// 继续处理请求
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// CORSMiddleware CORS 中间件
func CORSMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// 处理预检请求
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// LoggingMiddleware 请求日志中间件
func LoggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// 记录请求
			util.Log().Debug("[API] %s %s", r.Method, r.URL.Path)

			// 包装 ResponseWriter 以捕获状态码
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			// 记录响应
			duration := time.Since(start)
			util.Log().Debug("[API] %s %s - %d (%s)", r.Method, r.URL.Path, wrapped.statusCode, duration)
		})
	}
}

// responseWriter 包装器，用于捕获状态码
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
