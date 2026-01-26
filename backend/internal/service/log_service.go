// Package service 提供业务逻辑层服务
package service

import (
	"gost-panel/internal/dto"
	"gost-panel/internal/model"
	"gost-panel/internal/repository"
	"gost-panel/pkg/logger"

	"gorm.io/gorm"
)

// LogService 日志服务
// 负责操作日志的查询
type LogService struct {
	logRepo *repository.OperationLogRepository
}

// NewLogService 创建日志服务
func NewLogService(db *gorm.DB) *LogService {
	return &LogService{
		logRepo: repository.NewOperationLogRepository(db),
	}
}

// List 获取日志列表
func (s *LogService) List(req *dto.LogListReq) ([]model.OperationLog, int64, error) {
	// 设置默认值
	req.SetDefaults()

	opt := &repository.QueryOption{
		Pagination: &repository.Pagination{
			Page:     req.Page,
			PageSize: req.PageSize,
		},
		Conditions: make(map[string]any),
	}

	if req.Action != "" {
		opt.Conditions["action = ?"] = req.Action
	}
	if req.ResourceType != "" {
		opt.Conditions["resource_type = ?"] = req.ResourceType
	}
	if req.Username != "" {
		opt.Conditions["username LIKE ?"] = "%" + req.Username + "%"
	}

	return s.logRepo.List(opt)
}

// LogAction 操作日志参数
type LogAction struct {
	UserID       uint
	Username     string
	Action       string
	ResourceType string
	ResourceID   uint
	Details      string
	IP           string
	UserAgent    string
}

// Record 记录操作日志
func (s *LogService) Record(userID uint, username, action, resourceType string, resourceID uint, details, ip, userAgent string) {
	s.RecordAction(&LogAction{
		UserID:       userID,
		Username:     username,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      details,
		IP:           ip,
		UserAgent:    userAgent,
	})
}

// RecordAction 使用结构体记录操作日志
func (s *LogService) RecordAction(action *LogAction) {
	log := &model.OperationLog{
		UserID:       action.UserID,
		Username:     action.Username,
		Action:       action.Action,
		ResourceType: action.ResourceType,
		ResourceID:   action.ResourceID,
		Details:      action.Details,
		IPAddress:    action.IP,
		UserAgent:    action.UserAgent,
	}
	if err := s.logRepo.Create(log); err != nil {
		logger.Errorf("记录操作日志失败: %v", err)
	}
}
