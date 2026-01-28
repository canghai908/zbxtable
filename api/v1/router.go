package v1

import (
	"strings"
	"zbxtable/internal/handler"
	"zbxtable/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// InitRouter 初始化 Gin 路由
func InitRouter() *gin.Engine {
	// 设置运行模式
	//gin.SetMode(gin.ReleaseMode)
	gin.SetMode(gin.DebugMode)
	r := gin.New()

	// 中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// CORS 中间件
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "X-Token", "Authorization", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"}
	config.ExposeHeaders = []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"}
	config.AllowCredentials = true
	r.Use(cors.New(config))

	// 静态文件
	r.Static("/download", "./download")

	// WebSocket（必须在 NoRoute 之前）
	r.GET("/ws/:id", handler.WebSocketHandlerGin)

	// 安装引导 API（无需认证，必须在 NoRoute 之前）
	installGroup := r.Group("/install")
	{
		installGroup.GET("/status", handler.GetInstallStatus)
		installGroup.POST("/check-db", handler.CheckDatabase)
		installGroup.POST("/check-redis", handler.CheckRedis)
		installGroup.POST("/install", handler.DoInstall)
	}

	// 前端静态文件（SPA 应用）
	// 注意：在生产环境中，前端文件应该已经打包到 web 目录
	r.StaticFile("/", "./web/index.html")
	r.Static("/static", "./web/static")
	r.Static("/assets", "./web/assets")

	// SPA 路由支持：所有非 API 路由都返回 index.html（最后注册，作为兜底）
	r.NoRoute(func(c *gin.Context) {
		// 如果是 API 请求，返回 404
		if strings.HasPrefix(c.Request.URL.Path, "/v1") ||
			strings.HasPrefix(c.Request.URL.Path, "/install") ||
			strings.HasPrefix(c.Request.URL.Path, "/ws") {
			c.JSON(404, gin.H{"code": 404, "message": "Not Found"})
			return
		}
		// 否则返回前端页面（SPA 路由）
		c.File("./web/index.html")
	})

	// 安装状态检查中间件（除了安装路由）
	r.Use(middleware.CheckInstallStatusMiddleware())

	// API v1
	v1 := r.Group("/v1")
	{
		// 认证相关（无需 token）
		v1.POST("/login", handler.LoginGin)
		v1.POST("/logout", handler.LogoutGin)
		v1.POST("/receive", handler.ReceiveGin)
		v1.POST("/webhook", handler.WebhookGin)

		// 需要认证的路由
		api := v1.Group("")
		api.Use(middleware.JWTAuthMiddleware())
		{
			// Zabbix 实例管理（多 Zabbix）
			zabbixGroup := api.Group("/zabbix")
			{
				zabbixGroup.GET("/instances", handler.ListZabbixInstancesGin)
				zabbixGroup.GET("/instances/:id", handler.GetZabbixInstanceGin)
				zabbixGroup.POST("/instances", handler.CreateZabbixInstanceGin)
				zabbixGroup.POST("/instances/test", handler.TestZabbixInstanceConfigGin)
				zabbixGroup.POST("/instances/:id/test", handler.TestZabbixInstanceGin)
				zabbixGroup.PUT("/instances/:id", handler.UpdateZabbixInstanceGin)
				zabbixGroup.DELETE("/instances/:id", handler.DeleteZabbixInstanceGin)
				zabbixGroup.PUT("/instances/:id/enabled", handler.EnableZabbixInstanceGin)
				zabbixGroup.PUT("/instances/:id/activate", handler.ActivateZabbixInstanceGin)
				zabbixGroup.GET("/active", handler.GetActiveZabbixInstanceGin)

				// 租户绑定（tenant -> zabbix_instance + token）
				zabbixGroup.GET("/tenants", handler.ListZabbixTenantBindingsGin)
				zabbixGroup.POST("/tenants", handler.UpsertZabbixTenantBindingGin)
				zabbixGroup.DELETE("/tenants/:id", handler.DeleteZabbixTenantBindingGin)
			}

			// 首页/仪表板
			indexGroup := api.Group("/index")
			{
				indexGroup.GET("/routers", handler.GetRouters)
				indexGroup.GET("/baseinfo", handler.GetBaseInfo)
				indexGroup.GET("/restop", handler.GetResourceTop)
				indexGroup.GET("/inventory", handler.GetInventory)
				indexGroup.GET("/overview", handler.GetOverview)
				indexGroup.GET("/egress", handler.GetEgressData)
				indexGroup.GET("/version", handler.GetVersion)
				indexGroup.GET("/session", handler.GetZbxSession)
			}

			// 告警管理
			alarmGroup := api.Group("/alarm")
			{
				alarmGroup.GET("", handler.GetAllAlarm)
				alarmGroup.GET("/:id", handler.GetAlarmByID)
				alarmGroup.GET("/tenant", handler.GetAlarmTenant)
				alarmGroup.POST("/analysis", handler.AnalysisAlarm)
				alarmGroup.POST("/export", handler.ExportAlarm)
				alarmGroup.POST("", handler.CreateAlarm)
				alarmGroup.PUT("/:id", handler.UpdateAlarm)
				alarmGroup.DELETE("/:id", handler.DeleteAlarm)
			}

			// 问题管理
			problemGroup := api.Group("/problem")
			{
				problemGroup.GET("", handler.GetAllProblem)
			}

			// 触发器管理
			triggerGroup := api.Group("/trigger")
			{
				triggerGroup.GET("", handler.GetAllTrigger)
				triggerGroup.GET("/list", handler.GetTriggerList)
			}

			// 导出功能
			exportGroup := api.Group("/export")
			{
				exportGroup.POST("/trend", handler.ExportTrend)
				exportGroup.POST("/history", handler.ExportHistory)
				exportGroup.POST("/inspect", handler.ExportInspect)
				exportGroup.POST("/hosts", handler.ExportHosts)
				exportGroup.POST("/inventory", handler.ExportInventory)
			}

			// 主机管理
			hostGroup := api.Group("/host")
			{
				hostGroup.GET("", handler.GetAllHost)
				hostGroup.GET("/:hostid", handler.GetHostByID)
				hostGroup.POST("", handler.UpdateHost)
				hostGroup.GET("/search", handler.SearchHost)
				hostGroup.GET("/monitem/:hostid", handler.GetMonItem)
				hostGroup.GET("/interface/:hostid", handler.GetMonInterface)
				hostGroup.POST("/interface/data", handler.GetOneInterface)
				hostGroup.GET("/winmon/:hostid", handler.GetMonWinFileSystem)
				hostGroup.GET("/linmon/:hostid", handler.GetMonLinFileSystem)
				hostGroup.POST("/graph/:hostid", handler.GetHostGraph)
			}

			// 主机组管理
			hostGroupGroup := api.Group("/host_group")
			{
				hostGroupGroup.GET("", handler.GetAllHostGroup)
				hostGroupGroup.GET("/list", handler.GetAllHostGroupsList)
				hostGroupGroup.GET("/all", handler.GetAllGroupsList)
				hostGroupGroup.GET("/list/:id", handler.GetHostsByGroupID)
			}

			// 模板管理
			templateGroup := api.Group("/template")
			{
				templateGroup.GET("", handler.GetAllTemplate)
				templateGroup.GET("/all", handler.GetAllTemplateAll)
				templateGroup.GET("/list", handler.GetAllTemplateList)
				templateGroup.GET("/item/:templateid", handler.GetItemByTemplateID)
			}

			// 监控项管理
			itemGroup := api.Group("/item")
			{
				itemGroup.GET("", handler.GetItemByKey)
				itemGroup.GET("/list", handler.GetAllItemByHostID)
				itemGroup.GET("/traffic", handler.GetAllTrafficItem)
				itemGroup.GET("/topotraffic", handler.GetAllReceiveTrafficItem)
			}

			// 历史数据
			historyGroup := api.Group("/history")
			{
				historyGroup.POST("", handler.GetHistoryByItemID)
			}

			// 趋势数据
			trendGroup := api.Group("/trend")
			{
				trendGroup.GET("", handler.GetTrendByItemID)
			}

			// 图表管理
			graphGroup := api.Group("/graph")
			{
				graphGroup.POST("", handler.GetGraphByHostID)
				graphGroup.POST("/exp", handler.ExportGraph)
			}

			// 图片处理
			imagesGroup := api.Group("/images")
			{
				imagesGroup.GET("/:id", handler.GetImage)
			}

			// 拓扑管理
			topologyGroup := api.Group("/topology")
			{
				topologyGroup.GET("", handler.GetAllTopology)
				topologyGroup.GET("/:id", handler.GetTopologyByID)
				topologyGroup.POST("", handler.CreateTopology)
				topologyGroup.PUT("/:id", handler.UpdateTopology)
				topologyGroup.DELETE("/:id", handler.DeleteTopology)
				topologyGroup.POST("/status", handler.UpdateTopologyStatus)
			}

			// 拓扑数据
			topodataGroup := api.Group("/topodata")
			{
				topodataGroup.GET("/:id", handler.GetTopoDataByID)
				topodataGroup.POST("", handler.CreateTopoData)
			}

			// 系统配置
			systemGroup := api.Group("/system")
			{
				systemGroup.GET("", handler.GetAllSystem)
				systemGroup.GET("/:id", handler.GetSystemByID)
				systemGroup.PUT("/:id", handler.UpdateSystem)
				systemGroup.POST("/init/:id", handler.SystemInit)
				systemGroup.GET("/egress", handler.GetEgress)
				systemGroup.PUT("/egress", handler.UpdateEgress)
				systemGroup.GET("/config", handler.GetAllConfig)
				systemGroup.PUT("/config/:id", handler.UpdateConfig)
			}

			// 报表管理
			reportGroup := api.Group("/report")
			{
				reportGroup.GET("", handler.GetReportGin)
				reportGroup.GET("/:id", handler.GetReportOneGin)
				reportGroup.POST("", handler.CreateReportGin)
				reportGroup.PUT("/:id", handler.UpdateReportGin)
				reportGroup.DELETE("/:id", handler.DeleteReportGin)
				reportGroup.POST("/checknow", handler.CheckNowGin)
				reportGroup.POST("/status", handler.UpdateReportStatusGin)
				reportGroup.GET("/hosts", handler.GetReportHostsGin)
				reportGroup.GET("/items", handler.GetReportItemsGin)
			}

			// 任务日志
			taskLogGroup := api.Group("/task_log")
			{
				taskLogGroup.GET("", handler.GetTaskLogByReportID)
				taskLogGroup.DELETE("/:id", handler.DeleteTaskLog)
			}

			// 事件日志
			eventLogGroup := api.Group("/event_log")
			{
				eventLogGroup.GET("/:id", handler.GetEventLogByAlarmID)
			}

			// 规则管理
			ruleGroup := api.Group("/rule")
			{
				ruleGroup.GET("", handler.GetAllRule)
				ruleGroup.GET("/:id", handler.GetRuleByID)
				ruleGroup.POST("", handler.CreateRule)
				ruleGroup.PUT("/:id", handler.UpdateRule)
				ruleGroup.PUT("/status/:id", handler.UpdateRuleStatus)
				ruleGroup.DELETE("/:id", handler.DeleteRule)
			}

			// 用户管理
			userGroup := api.Group("/user")
			{
				userGroup.GET("", handler.GetUserGin)
				userGroup.POST("", handler.CreateUserGin)
				userGroup.PUT("/:id", handler.UpdateUserGin)
				userGroup.DELETE("/:id", handler.DeleteUserGin)
			}

			// 用户组管理
			groupGroup := api.Group("/group")
			{
				groupGroup.GET("", handler.GetAllGroup)
				groupGroup.POST("", handler.CreateGroup)
				groupGroup.PUT("/:id", handler.UpdateGroup)
				groupGroup.PUT("/member/:id", handler.UpdateGroupMember)
				groupGroup.DELETE("/:id", handler.DeleteGroup)
			}

			// AI 聊天
			aiGroup := api.Group("/ai")
			{
				aiGroup.POST("/chat", handler.AIChat)
			}
		}
	}

	return r
}
