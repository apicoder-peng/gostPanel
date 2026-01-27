package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"gost-panel/internal/config"
	"gost-panel/internal/model"
	"gost-panel/internal/router"
	"gost-panel/internal/service"
	"gost-panel/pkg/jwt"
	"gost-panel/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func main() {
	// 解析命令行参数
	var configPath string
	flag.StringVar(&configPath, "c", "", "配置文件路径")
	flag.StringVar(&configPath, "config", "", "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(configPath)
	if err != nil {
		panic("加载配置失败: " + err.Error())
	}

	// 初始化日志
	if err = logger.Init(&logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
		Output: cfg.Log.Output,
	}); err != nil {
		panic("初始化日志失败: " + err.Error())
	}
	defer func() {
		_ = logger.Sync()
	}()

	logger.Info("Gost Panel 启动中...")

	// 设置 Gin 模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化数据库
	db, err := initDatabase(cfg)
	if err != nil {
		logger.Fatalf("初始化数据库失败: %v", err)
	}
	logger.Info("数据库初始化完成")

	// 自动迁移
	if err = autoMigrate(db); err != nil {
		logger.Fatalf("数据库迁移失败: %v", err)
	}
	logger.Info("数据库迁移完成")

	// 初始化默认管理员
	if err = initDefaultAdmin(db, cfg); err != nil {
		logger.Fatalf("初始化管理员失败: %v", err)
	}

	// 初始化系统配置
	if err = initSystemConfig(db); err != nil {
		logger.Fatalf("初始化系统配置失败: %v", err)
	}

	// 创建 Gin 引擎
	engine := gin.New()

	// 配置路由
	jwtCfg := &jwt.Config{
		Secret: cfg.JWT.Secret,
		Expire: cfg.JWT.Expire,
	}
	r := router.NewRouter(db, jwtCfg)
	r.Setup(engine)

	// 启动服务器
	go func() {
		logger.Infof("服务器启动在 %s", cfg.Server.Port)
		if err = engine.Run(cfg.Server.Port); err != nil {
			logger.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 启动节点健康检测服务
	healthService := service.NewNodeHealthService(db)
	healthService.Start()

	// 启动规则状态同步服务
	syncService := service.NewRuleSyncService(db)
	syncService.Start()

	// 启动自动备份服务
	backupService := service.NewBackupService(db)
	backupService.Start()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Gost Panel 正在关闭...")

	// 停止相关的后台服务
	healthService.Stop()
	syncService.Stop()
	backupService.Stop()
}

// initDatabase 初始化数据库
func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	// 配置 GORM 日志
	var gormLogLevel gormlogger.LogLevel
	switch cfg.Log.Level {
	case "debug":
		gormLogLevel = gormlogger.Info
	case "info":
		gormLogLevel = gormlogger.Warn
	default:
		gormLogLevel = gormlogger.Error
	}

	gormConfig := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormLogLevel),
	}

	// 连接 SQLite 数据库
	db, err := gorm.Open(sqlite.Open(cfg.Database.Path), gormConfig)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// autoMigrate 自动迁移数据库表结构
func autoMigrate(db *gorm.DB) error {
	// 1. 执行自动迁移（添加新字段）
	if err := db.AutoMigrate(
		&model.User{},
		&model.GostNode{},
		&model.GostRule{},
		&model.GostTunnel{},
		&model.OperationLog{},
		&model.SystemConfig{},
	); err != nil {
		return err
	}

	return nil
}

// initDefaultAdmin 初始化默认管理员
func initDefaultAdmin(db *gorm.DB, cfg *config.Config) error {
	jwtCfg := &jwt.Config{
		Secret: cfg.JWT.Secret,
		Expire: cfg.JWT.Expire,
	}
	authService := service.NewAuthService(db, jwtCfg)
	return authService.InitDefaultAdmin("admin", "admin123")
}

// initSystemConfig 初始化系统配置
func initSystemConfig(db *gorm.DB) error {
	var count int64
	if err := db.Model(&model.SystemConfig{}).Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		// 设置默认值
		sysConfig := &model.SystemConfig{
			SiteTitle:            "Gost Panel",
			LogoURL:              "https://gost.run/images/gost.png",
			Copyright:            "https://github.com/code-gopher/gostPanel",
			LogRetentionDays:     7,
			LogLevel:             "info",
			AutoBackup:           false,
			BackupRetentionCount: 7,
		}
		if err := db.Create(sysConfig).Error; err != nil {
			return err
		}
		logger.Info("初始化默认系统配置完成")
	}
	return nil
}
