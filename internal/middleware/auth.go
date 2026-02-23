package middleware

import (
	"net/http"
	"os"
	"strings"

	"zbxtable/internal/model"

	jwtbeego "github.com/canghai908/jwt-beego"
	"github.com/gin-gonic/gin"
	"gopkg.in/ini.v1"
)

// JWTAuthMiddleware JWT 认证中间件
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("X-Token")
		if tokenString == "" {
			// 尝试从 Authorization header 获取
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    50014,
				"message": "token过期或非法的token",
				"data": gin.H{
					"items": "",
					"total": 0,
				},
			})
			c.Abort()
			return
		}

		et := jwtbeego.EasyToken{}
		valid, username, _ := et.ValidateToken(tokenString)
		if !valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    50014,
				"message": "token过期或非法的token",
				"data": gin.H{
					"items": "",
					"total": 0,
				},
			})
			c.Abort()
			return
		}

		// 将用户名存储到上下文中
		c.Set("username", username)
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

		// 已安装，继续
		c.Next()
	}
}
