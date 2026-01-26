package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gost-panel/internal/config"
	"gost-panel/internal/errors"
	"gost-panel/internal/repository"
	"gost-panel/pkg/logger"

	"gorm.io/gorm"
)

// BackupService 备份服务
type BackupService struct {
	db      *gorm.DB
	sysRepo *repository.SystemConfigRepository

	stopChan chan struct{}
}

// NewBackupService 创建备份服务
func NewBackupService(db *gorm.DB) *BackupService {
	return &BackupService{
		db:       db,
		sysRepo:  repository.NewSystemConfigRepository(db),
		stopChan: make(chan struct{}),
	}
}

// Start 启动自动备份任务 (按小时检查)
func (s *BackupService) Start() {
	go func() {
		// 启动时先检查一次
		s.processAutoBackup()

		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.processAutoBackup()
			case <-s.stopChan:
				return
			}
		}
	}()
	logger.Info("自动备份服务已启动")
}

// Stop 停止自动备份任务
func (s *BackupService) Stop() {
	close(s.stopChan)
	logger.Info("自动备份服务已停止")
}

// processAutoBackup 执行自动备份逻辑
func (s *BackupService) processAutoBackup() {
	cfg, err := s.sysRepo.Get()
	if err != nil {
		logger.Errorf("自动备份检查失败: 获取系统配置错误: %v", err)
		return
	}

	if !cfg.AutoBackup {
		return
	}

	backupDir := "backups"
	// 确保目录存在
	if err = os.MkdirAll(backupDir, 0755); err != nil {
		logger.Errorf("自动备份失败: 无法创建目录: %v", err)
		return
	}

	// 检查今天是否已经备份过
	// 简单策略: 每天生成一个文件名含日期的文件
	// 如果存在今天的文件，则跳过
	today := time.Now().Format("20060102")
	prefix := fmt.Sprintf("gost_panel_%s", today)

	files, err := os.ReadDir(backupDir)
	if err != nil {
		logger.Errorf("自动备份失败: 读取目录错误: %v", err)
		return
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), prefix) {
			// 今天已备份
			return
		}
	}

	// 执行备份
	logger.Infof("开始执行自动备份...")
	if err := s.CreateBackup(); err != nil {
		logger.Errorf("自动备份执行失败: %v", err)
	}
}

// CreateBackup 创建备份
func (s *BackupService) CreateBackup() error {
	// 获取数据库路径
	dbPath := config.Get().Database.Path
	if dbPath == "" {
		return errors.ErrDBPathNotConfigured
	}

	// 确保备份目录存在
	backupDir := "backups"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return errors.ErrBackupDirCreateFailed
	}

	// 生成备份文件名
	filename := fmt.Sprintf("gost_panel_%s.db", time.Now().Format("20060102_150405"))
	targetPath := filepath.Join(backupDir, filename)

	// 使用 VACUUM INTO 进行在线备份 (SQLite)
	err := s.db.Exec("VACUUM INTO ?", targetPath).Error
	if err != nil {
		logger.Warnf("VACUUM INTO 备份失败 (%v)，尝试直接文件复制", err)
		if err := copyFile(dbPath, targetPath); err != nil {
			return errors.ErrBackupFailed
		}
	}

	logger.Infof("数据库备份成功: %s", targetPath)

	// 清理旧备份
	go s.cleanOldBackups(backupDir)

	return nil
}

// cleanOldBackups 清理旧备份
func (s *BackupService) cleanOldBackups(backupDir string) {
	cfg, err := s.sysRepo.Get()
	if err != nil {
		logger.Errorf("获取系统配置失败: %v", err)
		return
	}

	retentionCount := cfg.BackupRetentionCount
	if retentionCount <= 0 {
		return
	}

	// 获取所有备份文件
	files, err := os.ReadDir(backupDir)
	if err != nil {
		logger.Errorf("读取备份目录失败: %v", err)
		return
	}

	var backupFiles []os.DirEntry
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "gost_panel_") && strings.HasSuffix(file.Name(), ".db") {
			backupFiles = append(backupFiles, file)
		}
	}

	// 如果备份数量未超过保留数量，直接返回
	if len(backupFiles) <= retentionCount {
		return
	}

	// 按修改时间排序（从新到旧） (按文件名排序即可，因包含时间戳)
	sort.Slice(backupFiles, func(i, j int) bool {
		return backupFiles[i].Name() > backupFiles[j].Name() // 降序，新的在前
	})

	// 删除多余的备份
	for i := retentionCount; i < len(backupFiles); i++ {
		filePath := filepath.Join(backupDir, backupFiles[i].Name())
		if err := os.Remove(filePath); err != nil {
			logger.Errorf("删除旧备份失败 %s: %v", filePath, err)
		} else {
			logger.Infof("已删除旧备份: %s", filePath)
		}
	}
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func(sourceFile *os.File) {
		_ = sourceFile.Close()
	}(sourceFile)

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func(destFile *os.File) {
		_ = destFile.Close()
	}(destFile)

	_, err = io.Copy(destFile, sourceFile)
	return err
}
