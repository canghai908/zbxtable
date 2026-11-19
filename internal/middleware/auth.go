package middleware

import (
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"zbxtable/internal/auth"
	"zbxtable/internal/model"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
	"gopkg.in/ini.v1"
)

// installConfirmed 一旦确认已安装即置为 true。安装是进程内单向状态
// （完成安装需重启进程才会连接数据库），因此确认后即可跳过每请求的
// os.Stat + ini.Load 磁盘读取。
var installConfirmed atomic.Bool

// extractRequestToken 从 header / Bearer / Cookie 提取登录令牌。
// 静态资源（如 <a href="/download/...">、<img src="/upload/...">）只能依赖 Cookie。
func extractRequestToken(c *gin.Context) string {
	if token := strings.TrimSpace(c.GetHeader("X-Token")); token != "" {
		return token
	}
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	}
	if cookie, err := c.Cookie("X-Token"); err == nil {
		return strings.TrimSpace(cookie)
	}
	return ""
}

func abortUnauthorizedToken(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"code":    50014,
		"message": "token过期或非法的token",
		"data": gin.H{
			"items": "",
			"total": 0,
		},
	})
	c.Abort()
}

// JWTAuthMiddleware JWT 认证中间件
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractRequestToken(c)
		if tokenString == "" {
			abortUnauthorizedToken(c)
			return
		}

		valid, username, _ := auth.ValidateToken(tokenString)
		if !valid {
			abortUnauthorizedToken(c)
			return
		}

		c.Set("username", username)
		c.Next()
	}
}

// JWTAuthUnless 对指定前缀放行，其余路径要求登录。
// 用于 /upload：logo/background 需给登录页与公开分享页匿名访问。
func JWTAuthUnless(publicPrefixes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		for _, prefix := range publicPrefixes {
			if path == prefix || strings.HasPrefix(path, prefix+"/") {
				c.Next()
				return
			}
		}
		JWTAuthMiddleware()(c)
	}
}

// RequireAdmin 要求当前登录用户角色为 admin。
// 用于用户管理、系统配置等高危写接口，防止任意有效令牌提权。
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		usernameValue, ok := c.Get("username")
		if !ok {
			response.Unauthorized(c, "登录状态失效")
			c.Abort()
			return
		}
		username, _ := usernameValue.(string)
		if username == "" {
			response.Unauthorized(c, "登录状态失效")
			c.Abort()
			return
		}

		var user model.User
		if err := model.GetDB().Where("username = ?", username).First(&user).Error; err != nil {
			response.Forbidden(c, "无权限执行此操作")
			c.Abort()
			return
		}
		if user.Role != "admin" {
			response.Forbidden(c, "需要管理员权限")
			c.Abort()
			return
		}
		c.Set("role", user.Role)
		c.Next()
	}
}

// DemoModeMiddleware 演示模式中间件
// 需求：演示模式配置在配置文件中，不额外创建演示账号；仍使用 admin 登录
// 行为：开启 demo_mode=true 时，拦截所有“可能修改数据”的请求（非 GET），从而保护演示环境数据
func DemoModeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 优先从 app.conf / .env 读取（model.GetConfKey 会优先读取 .env，其次 app.conf）
		demoMode := model.GetConfKey("demo_mode")
		if demoMode == "" {
			// 如果配置文件没配置，则回退到数据库配置（可选）
			demoMode = model.GetConfigValueByKey("demo_mode", "false")
		}

		// 统一转换成 bool 语义（仅识别 "true"）
		isDemo := strings.EqualFold(strings.TrimSpace(demoMode), "true")
		if !isDemo {
			c.Next()
			return
		}

		// 允许 GET 请求（只读）
		if c.Request.Method == http.MethodGet {
			c.Next()
			return
		}

		// 允许演示环境的必要写入：登录/注销、告警接收/webhook
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/v1/login") ||
			strings.HasPrefix(path, "/v1/logout") ||
			strings.HasPrefix(path, "/v1/receive") ||
			strings.HasPrefix(path, "/v1/webhook") {
			c.Next()
			return
		}

		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "演示模式已开启：禁止修改数据",
		})
		c.Abort()
	}
}

// CheckInstallStatusMiddleware 检查安装状态中间件
// 如果未安装，重定向到安装页面（除了安装相关的路由）
// 注意：只在配置文件存在但配置不完整时才跳转，避免后台服务未启动时误判
func CheckInstallStatusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 已确认安装：跳过磁盘检查
		if installConfirmed.Load() {
			c.Next()
			return
		}

		// 安装相关的路由不需要检查
		if strings.HasPrefix(c.Request.URL.Path, "/install") {
			c.Next()
			return
		}

		// 检查安装状态
		confPath := "./config/app.conf"
		_, err := os.Stat(confPath)
		if err != nil {
			// 配置文件不存在：可能是首次安装或后台服务未启动
			// 只对 API 请求返回错误，不跳转（避免后台未启动时误跳转）
			if c.Request.Header.Get("Accept") != "" && strings.Contains(c.Request.Header.Get("Accept"), "application/json") {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"code":    503,
					"message": "服务暂时不可用，请检查后台服务是否启动",
				})
				c.Abort()
				return
			}
			// 对于页面请求，让前端自行处理（不强制跳转）
			c.Next()
			return
		}

		cfg, err := ini.Load(confPath)
		if err != nil {
			// 配置文件存在但无法解析：配置文件损坏
			if c.Request.Header.Get("Accept") != "" && strings.Contains(c.Request.Header.Get("Accept"), "application/json") {
				c.JSON(http.StatusOK, gin.H{
					"code":     500,
					"message":  "配置文件损坏，请重新安装",
					"redirect": "/install",
				})
			} else {
				c.Redirect(http.StatusFound, "/install")
			}
			c.Abort()
			return
		}

		dbtype := cfg.Section("").Key("dbtype").String()
		dbname := cfg.Section("").Key("dbname").String()

		// 检查数据库类型和数据库名
		if dbtype == "" || dbname == "" {
			// 配置文件存在但配置不完整：需要完成安装
			if c.Request.Header.Get("Accept") != "" && strings.Contains(c.Request.Header.Get("Accept"), "application/json") {
				c.JSON(http.StatusOK, gin.H{
					"code":     500,
					"message":  "系统未安装，请先完成安装",
					"redirect": "/install",
				})
			} else {
				c.Redirect(http.StatusFound, "/install")
			}
			c.Abort()
			return
		}

		// 对于非 SQLite 数据库，需要检查 dbhost
		if dbtype != "sqlite" {
			dbhost := cfg.Section("").Key("dbhost").String()
			if dbhost == "" {
				if c.Request.Header.Get("Accept") != "" && strings.Contains(c.Request.Header.Get("Accept"), "application/json") {
					c.JSON(http.StatusOK, gin.H{
						"code":     500,
						"message":  "系统未安装，请先完成安装",
						"redirect": "/install",
					})
				} else {
					c.Redirect(http.StatusFound, "/install")
				}
				c.Abort()
				return
			}
		}

		// 已安装，继续（记录状态，后续请求跳过磁盘检查）
		installConfirmed.Store(true)
		c.Next()
	}
}
