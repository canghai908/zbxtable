package middleware

import (
	"net/http"
	"os"
	"strings"

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

// CheckInstallStatusMiddleware 检查安装状态中间件
// 如果未安装，重定向到安装页面（除了安装相关的路由）
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
			// 未安装，重定向到安装页面
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

		cfg, err := ini.Load(confPath)
		if err != nil {
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

		dbtype := cfg.Section("").Key("dbtype").String()
		dbhost := cfg.Section("").Key("dbhost").String()
		dbname := cfg.Section("").Key("dbname").String()

		if dbtype == "" || dbhost == "" || dbname == "" {
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

		// 已安装，继续
		c.Next()
	}
}
