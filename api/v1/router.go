package v1

import (
	"io/fs"
	"net/http"
	"strings"
	"zbxtable/api/v1/web"
	"zbxtable/internal/handler"
	"zbxtable/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

// InitRouter 初始化 Gin 路由
func InitRouter() *gin.Engine {
	// 设置运行模式
	gin.SetMode(gin.ReleaseMode)
	//gin.SetMode(gin.DebugMode)
	r := gin.New()

	// 中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Gzip 压缩中间件（用于压缩静态资源和 API 响应）
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	// CORS 中间件
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "X-Token", "Authorization", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"}
	config.ExposeHeaders = []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"}
	config.AllowCredentials = true
	r.Use(cors.New(config))

	// 下载文件目录（使用文件系统）
	r.Static("/download", "./download")

	// 上传文件目录（静态文件服务）
	r.Static("/upload", "./upload")

	// 获取嵌入的前端文件系统
	distFS := web.GetDistFSRoot()
	subFS, err := fs.Sub(distFS, "web")
	if err != nil {
		panic("Failed to get embedded frontend filesystem: " + err.Error())
	}

	// 前端静态文件服务（使用 go:embed，不需要安装检查）。
	// Vite 默认输出到 /assets，保留 /static 和 /css 兼容旧 Vue CLI 产物。
	// 注意：这些路由必须在安装检查中间件之前注册
	mountEmbeddedDir := func(urlPath, dir string) {
		if _, err := fs.Stat(subFS, dir); err != nil {
			return
		}
		dirFS, err := fs.Sub(subFS, dir)
		if err != nil {
			panic("Failed to get embedded frontend filesystem: " + err.Error())
		}
		r.StaticFS(urlPath, http.FS(dirFS))
	}
	mountEmbeddedDir("/assets", "assets")
	mountEmbeddedDir("/static", "static")
	mountEmbeddedDir("/css", "css")

	serveFrontendFile := func(path string, contentType string) gin.HandlerFunc {
		return func(c *gin.Context) {
			data, err := fs.ReadFile(subFS, path)
			if err != nil {
				c.Status(404)
				return
			}
			c.Data(200, contentType, data)
		}
	}

	// 根级静态资源
	r.GET("/favicon.ico", serveFrontendFile("favicon.ico", "image/x-icon"))
	r.GET("/logo.png", serveFrontendFile("logo.png", "image/png"))
	r.GET("/vite.svg", serveFrontendFile("vite.svg", "image/svg+xml"))

	// 根路径返回 index.html
	r.GET("/", func(c *gin.Context) {
		data, err := fs.ReadFile(subFS, "index.html")
		if err != nil {
			c.String(500, "Failed to load frontend: %v", err)
			return
		}
		c.Data(200, "text/html; charset=utf-8", data)
	})

	// 公开的拓扑预览接口（无需认证）
	r.GET("/public/topology/:id", handler.GetPublicTopologyByID)

	// WebSocket 路由（统一使用 /ws 前缀）
	r.GET("/ws/auth/:id", handler.WebSocketHandlerGin)      // 需要认证的 WebSocket
	r.GET("/ws/pub/:id", handler.PublicWebSocketHandlerGin) // 公开的 WebSocket（共享）

	// 安装引导 API（无需认证，必须在 NoRoute 之前）
	installGroup := r.Group("/install")
	{
		installGroup.GET("/status", handler.GetInstallStatus)
		installGroup.POST("/check-db", handler.CheckDatabase)
		installGroup.POST("/install", handler.DoInstall)
	}

	// 安装状态检查中间件（只应用于业务 API，不影响静态文件和安装路由）
	apiGroup := r.Group("")
	apiGroup.Use(middleware.CheckInstallStatusMiddleware())

	// API v1
	v1 := apiGroup.Group("/v1")
	{
		// 认证相关（无需 token）
		v1.POST("/login", handler.LoginGin)
		v1.POST("/logout", handler.LogoutGin)
		v1.POST("/receive", handler.ReceiveGin)
		v1.POST("/webhook", handler.WebhookGin)
		v1.GET("/info", handler.GetPublicSystemInfo)

		// 需要认证的路由
		api := v1.Group("")
		api.Use(middleware.JWTAuthMiddleware())
		api.Use(middleware.DemoModeMiddleware())
		{
			// Zabbix 租户管理（合并后的统一接口）
			zabbixGroup := api.Group("/zabbix")
			{
				// 租户管理
				zabbixGroup.GET("/instance", handler.ListZabbixInstanceGin)
				zabbixGroup.GET("/instance/:id", handler.GetZabbixInstanceGin)
				zabbixGroup.POST("/instance", handler.CreateZabbixInstanceGin)
				zabbixGroup.POST("/instance/test", handler.TestZabbixInstanceConfigGin)
				zabbixGroup.POST("/instance/:id/test", handler.TestZabbixInstanceGin)
				zabbixGroup.PUT("/instance/:id", handler.UpdateZabbixInstanceGin)
				zabbixGroup.DELETE("/instance/:id", handler.DeleteZabbixInstanceGin)
				zabbixGroup.PUT("/instance/:id/enabled", handler.EnableZabbixInstanceGin)
				// Webhook 安装相关
				zabbixGroup.POST("/instance/:id/install-webhook", handler.InstallWebhookGin)
				zabbixGroup.GET("/instance/:id/webhook-info", handler.GetWebhookInfoGin)
				zabbixGroup.DELETE("/instance/:id/uninstall-webhook", handler.UninstallWebhookGin)
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
				hostGroup.POST("", handler.UpdateHost)
				hostGroup.GET("/search", handler.SearchHost)
				hostGroup.GET("/filter-by-tag", handler.FilterHostsByTag)
				hostGroup.GET("/:hostid", handler.GetHostByID)
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
				hostGroupGroup.GET("", handler.GetAllGroupsList)
				//	hostGroupGroup.GET("/list", handler.GetAllGroupsList)
				hostGroupGroup.GET("/tree", handler.GetAllHostGroupsTree)
				hostGroupGroup.GET("/:id/hosts", handler.GetHostsByGroupID)
				hostGroupGroup.GET("/list/:id/hosts", handler.GetHostsByGroupID)
				hostGroupGroup.GET("/list/:id", handler.GetHostsByGroupID)
			}

			// 模板管理
			templateGroup := api.Group("/template")
			{
				templateGroup.GET("", handler.GetTemplateList)
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
				// 背景图片上传
				topologyGroup.POST("/upload-background", handler.UploadBackgroundImage)
				topologyGroup.DELETE("/delete-background", handler.DeleteBackgroundImage)
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
				systemGroup.POST("", handler.CreateSystem)
				systemGroup.GET("/:id", handler.GetSystemByID)
				systemGroup.PUT("/:id", handler.UpdateSystem)
				systemGroup.DELETE("/:id", handler.DeleteSystem)
				systemGroup.POST("/init/:id", handler.SystemInit)
				systemGroup.GET("/:id/history", handler.GetSystemHistory)
				systemGroup.GET("/config", handler.GetAllConfig)
				systemGroup.PUT("/config/:id", handler.UpdateConfig)
				systemGroup.POST("/upload-logo", handler.UploadLogo)

				// 初始配置状态
				systemGroup.GET("/setup-status", handler.GetInitialSetupStatus)
				systemGroup.POST("/complete-setup", handler.CompleteInitialSetup)

				// 配置测试
				systemGroup.POST("/test-email", handler.TestEmailConfig)
				systemGroup.POST("/test-wechat", handler.TestWechatConfig)

				// 系统更新
				systemGroup.GET("/version", handler.GetCurrentVersion)
				systemGroup.GET("/check-update", handler.CheckUpdate)
				systemGroup.POST("/update", handler.DoUpdate)
			}

			// 设备分组管理（一级菜单分类）
			assetGroupGroup := api.Group("/asset-group")
			{
				assetGroupGroup.GET("", handler.GetAllAssetGroups)
				assetGroupGroup.POST("", handler.CreateAssetGroup)
				assetGroupGroup.GET("/:id", handler.GetAssetGroupByID)
				assetGroupGroup.PUT("/:id", handler.UpdateAssetGroup)
				assetGroupGroup.DELETE("/:id", handler.DeleteAssetGroup)
			}

			// 资产类型管理
			assetTypeGroup := api.Group("/asset-type")
			{
				assetTypeGroup.GET("", handler.GetAllAssetTypes)
				assetTypeGroup.POST("", handler.CreateAssetType)
				assetTypeGroup.GET("/:id", handler.GetAssetTypeByID)
				assetTypeGroup.PUT("/:id", handler.UpdateAssetType)
				assetTypeGroup.DELETE("/:id", handler.DeleteAssetType)
				assetTypeGroup.GET("/:id/fields", handler.GetAssetTypeFields)
				assetTypeGroup.PUT("/:id/fields", handler.UpdateAssetTypeFields)
			}

			// 出口配置管理（新）
			egressConfigGroup := api.Group("/egress")
			{
				egressConfigGroup.GET("/configs", handler.GetAllEgressConfigs)
				egressConfigGroup.GET("/configs/:id", handler.GetEgressConfigByID)
				egressConfigGroup.POST("/configs", handler.AddEgressConfig)
				egressConfigGroup.PUT("/configs/:id", handler.UpdateEgressConfigHandler)
				egressConfigGroup.DELETE("/configs/:id", handler.DeleteEgressConfigHandler)
			}

			// 指标映射配置
			metricMappingGroup := api.Group("/metric_mapping")
			{
				metricMappingGroup.GET("", handler.GetMetricMappings)
				metricMappingGroup.GET("/:id", handler.GetMetricMappingByID)
				metricMappingGroup.POST("", handler.CreateOrUpdateMetricMapping)
				metricMappingGroup.PUT("/:id", handler.CreateOrUpdateMetricMapping)
				metricMappingGroup.DELETE("/:id", handler.DeleteMetricMapping)
				metricMappingGroup.POST("/:id/execute", handler.ExecuteMetricMappingManual)
				metricMappingGroup.GET("/history", handler.GetMappingHistory)

				// 指标匹配规则
				metricMappingGroup.GET("/rules", handler.GetMappingRules)
				metricMappingGroup.POST("/rules", handler.CreateOrUpdateMappingRule)
				metricMappingGroup.PUT("/rules/:id", handler.CreateOrUpdateMappingRule)
				metricMappingGroup.DELETE("/rules/:id", handler.DeleteMappingRule)

				// 按模板同步 inventory_link
				metricMappingGroup.POST("/sync/templates", handler.SyncMetricMappingOnTemplates)

				// Debug：回查模板 items（包含 inventory_link）
				metricMappingGroup.GET("/debug/template_items", handler.GetTemplateItemsDebug)
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

			managerGroup := api.Group("/manager")
			{
				managerGroup.POST("/chpwd", handler.ChangePasswordGin)
			}

			// Token 刷新（无感续期）
			api.GET("/token/refresh", handler.RefreshTokenGin)

			// AI 聊天
			aiGroup := api.Group("/ai")
			{
				aiGroup.POST("/chat", handler.AIChat)
			}

			// 菜单管理
			menuGroup := api.Group("/menu")
			{
				menuGroup.GET("", handler.GetAllMenus)
				menuGroup.GET("/parents", handler.GetParentMenus)
				menuGroup.GET("/children/:id", handler.GetMenusByParentID)
				menuGroup.GET("/:id", handler.GetMenuByID)
				menuGroup.POST("", handler.CreateMenu)
				menuGroup.PUT("/:id", handler.UpdateMenu)
				menuGroup.DELETE("/:id", handler.DeleteMenu)
			}
		}
	}

	// SPA 路由支持：所有非 API 路由都返回 index.html（最后注册，作为兜底）
	// 这个必须在最后注册，用于处理前端路由
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 如果是 API 请求或 WebSocket 请求，返回 404
		if strings.HasPrefix(path, "/v1") ||
			strings.HasPrefix(path, "/install") ||
			strings.HasPrefix(path, "/ws") ||
			strings.HasPrefix(path, "/download") {
			c.JSON(404, gin.H{"code": 404, "message": "Not Found"})
			return
		}

		// 对于根路径和其他前端路由，返回 index.html
		data, err := fs.ReadFile(subFS, "index.html")
		if err != nil {
			// 添加详细的错误信息
			c.String(500, "Failed to load frontend: %v", err)
			return
		}
		c.Data(200, "text/html; charset=utf-8", data)
	})

	return r
}
