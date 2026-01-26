package router

import (
	"gost-panel/internal/config"
	"gost-panel/internal/handler"
	"gost-panel/internal/middleware"
	"gost-panel/internal/repository"
	"gost-panel/internal/service"
	"gost-panel/pkg/jwt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Router 路由配置
type Router struct {
	db     *gorm.DB
	jwtCfg *jwt.Config
}

// NewRouter 创建路由实例
func NewRouter(db *gorm.DB, jwtCfg *jwt.Config) *Router {
	return &Router{
		db:     db,
		jwtCfg: jwtCfg,
	}
}

// Setup 配置路由
func (r *Router) Setup(engine *gin.Engine) {
	// 创建 JWT 实例
	jwtInstance := jwt.New(r.jwtCfg)

	// 全局中间件
	engine.Use(middleware.CORS())
	engine.Use(middleware.Logger())
	engine.Use(middleware.Recovery())
	engine.Use(middleware.ErrorHandler())

	// 健康检查
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1 路由组
	apiV1 := engine.Group("/api/v1")

	// API v1 健康检查
	apiV1.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "version": config.Version})
	})

	// 初始化服务
	authService := service.NewAuthService(r.db, r.jwtCfg)
	nodeService := service.NewNodeService(r.db)
	forwardService := service.NewForwardService(r.db)
	tunnelService := service.NewTunnelService(r.db)
	statsService := service.NewStatsService(r.db)
	logService := service.NewLogService(r.db)
	observerService := service.NewObserverService(r.db)

	// 初始化系统配置
	systemConfigRepo := repository.NewSystemConfigRepository(r.db)
	systemConfigService := service.NewSystemConfigService(systemConfigRepo)
	backupService := service.NewBackupService(r.db)

	// 初始化控制器
	authHandler := handler.NewAuthHandler(authService)
	nodeHandler := handler.NewNodeHandler(nodeService)
	forwardHandler := handler.NewForwardHandler(forwardService)
	tunnelHandler := handler.NewTunnelHandler(tunnelService)
	statsHandler := handler.NewStatsHandler(statsService)
	logHandler := handler.NewLogHandler(logService)
	observerHandler := handler.NewObserverHandler(observerService)
	systemConfigHandler := handler.NewSystemConfigHandler(systemConfigService, backupService)

	// 公开路由（无需认证）
	{
		apiV1.POST("/auth/login", authHandler.Login)
		// 流量上报接口
		apiV1.POST("/observer/report", observerHandler.Report)
		// 公开系统配置
		apiV1.GET("/system/public-config", systemConfigHandler.GetPublicConfig)
	}

	// 需要认证的路由
	authRoutes := apiV1.Group("")
	authRoutes.Use(middleware.Auth(jwtInstance))
	{
		// 认证相关
		authRoutes.GET("/auth/info", authHandler.GetUserInfo)
		authRoutes.PUT("/auth/password", authHandler.ChangePassword)
		authRoutes.POST("/auth/refresh", authHandler.RefreshToken)

		// 仪表盘统计
		authRoutes.GET("/dashboard/stats", statsHandler.GetDashboard)

		// 节点管理
		authRoutes.GET("/nodes", nodeHandler.List)
		authRoutes.GET("/nodes/:id", nodeHandler.GetByID)
		authRoutes.POST("/nodes", nodeHandler.Create)
		authRoutes.PUT("/nodes/:id", nodeHandler.Update)
		authRoutes.DELETE("/nodes/:id", nodeHandler.Delete)
		authRoutes.GET("/nodes/:id/config", nodeHandler.GetConfig)
		authRoutes.GET("/nodes/stats", nodeHandler.GetStats)

		// 转发规则管理
		authRoutes.GET("/forwards", forwardHandler.List)
		authRoutes.GET("/forwards/:id", forwardHandler.GetByID)
		authRoutes.POST("/forwards", forwardHandler.Create)
		authRoutes.PUT("/forwards/:id", forwardHandler.Update)
		authRoutes.DELETE("/forwards/:id", forwardHandler.Delete)
		authRoutes.POST("/forwards/:id/start", forwardHandler.Start)
		authRoutes.POST("/forwards/:id/stop", forwardHandler.Stop)
		authRoutes.GET("/forwards/stats", forwardHandler.GetStats)

		// 隧道管理
		authRoutes.GET("/tunnels", tunnelHandler.List)
		authRoutes.GET("/tunnels/:id", tunnelHandler.GetByID)
		authRoutes.POST("/tunnels", tunnelHandler.Create)
		authRoutes.PUT("/tunnels/:id", tunnelHandler.Update)
		authRoutes.DELETE("/tunnels/:id", tunnelHandler.Delete)
		authRoutes.POST("/tunnels/:id/start", tunnelHandler.Start)
		authRoutes.POST("/tunnels/:id/stop", tunnelHandler.Stop)
		authRoutes.GET("/tunnels/stats", tunnelHandler.GetStats)

		// 操作日志
		authRoutes.GET("/logs", logHandler.List)

		// 系统设置
		authRoutes.GET("/system/config", systemConfigHandler.GetConfig)
		authRoutes.PUT("/system/config", systemConfigHandler.UpdateConfig)
		authRoutes.POST("/system/email/test", systemConfigHandler.TestEmail)
		authRoutes.POST("/system/backup", systemConfigHandler.Backup)
	}

	// 静态文件
	r.setupStatic(engine)
}
