// Package service 提供业务逻辑层服务
package service

import (
	stderrors "errors"
	"fmt"

	"gost-panel/internal/dto"
	"gost-panel/internal/errors"
	"gost-panel/internal/model"
	"gost-panel/internal/repository"
	"gost-panel/internal/utils"
	"gost-panel/pkg/gost"
	"gost-panel/pkg/logger"

	"gorm.io/gorm"
)

// ForwardService 转发服务
// 负责端口转发规则的 CRUD 操作及启停控制
type ForwardService struct {
	forwardRepo *repository.ForwardRepository
	nodeRepo    *repository.NodeRepository
	sysRepo     *repository.SystemConfigRepository
	logService  *LogService
}

// NewForwardService 创建转发服务
func NewForwardService(db *gorm.DB) *ForwardService {
	return &ForwardService{
		forwardRepo: repository.NewForwardRepository(db),
		nodeRepo:    repository.NewNodeRepository(db),
		sysRepo:     repository.NewSystemConfigRepository(db),
		logService:  NewLogService(db),
	}
}

// Create 创建转发规则
// 创建前会检查节点是否存在以及端口是否被占用
func (s *ForwardService) Create(req *dto.CreateForwardReq, userID uint, username string, ip, userAgent string) (*model.GostForward, error) {
	// 检查节点是否存在
	_, err := s.nodeRepo.FindByID(req.NodeID)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrNodeNotFound
		}
		return nil, err
	}

	// 检查端口是否已被使用
	exists, err := s.forwardRepo.ExistsByPort(req.NodeID, req.ListenPort)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.ErrForwardPortExists
	}

	// 创建转发规则
	forward := &model.GostForward{
		NodeID:     req.NodeID,
		Name:       req.Name,
		Protocol:   model.ForwardProtocol(req.Protocol),
		ListenPort: req.ListenPort,

		Targets:   req.Targets,
		Strategy:  req.Strategy,
		EnableTLS: req.EnableTLS,
		Remark:    req.Remark,
		Status:    model.ForwardStatusStopped,
	}

	if err = s.forwardRepo.Create(forward); err != nil {
		return nil, err
	}

	// 记录操作日志
	s.logService.Record(
		userID,
		username,
		model.ActionCreate,
		model.ResourceTypeForward,
		forward.ID,
		fmt.Sprintf("创建转发规则: %s", forward.Name),
		ip,
		userAgent)

	logger.Infof("创建转发规则成功: %s (:%d)", forward.Name, forward.ListenPort)
	return forward, nil
}

// Update 更新转发规则
// 更新前会检查规则是否存在、是否在运行中以及新端口是否被占用
func (s *ForwardService) Update(id uint, req *dto.UpdateForwardReq, userID uint, username string, ip, userAgent string) (*model.GostForward, error) {
	// 查询转发规则
	forward, err := s.forwardRepo.FindByID(id)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrForwardNotFound
		}
		return nil, err
	}

	// 检查是否在运行中
	if forward.Status == model.ForwardStatusRunning {
		return nil, errors.ErrForwardRunning
	}

	// 检查端口是否已被使用（排除自身）
	exists, err := s.forwardRepo.ExistsByPort(forward.NodeID, req.ListenPort, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.ErrForwardPortExists
	}

	// 更新转发规则
	forward.Name = req.Name
	forward.Protocol = model.ForwardProtocol(req.Protocol)
	forward.ListenPort = req.ListenPort

	forward.Targets = req.Targets
	forward.Strategy = req.Strategy
	forward.EnableTLS = req.EnableTLS
	forward.Remark = req.Remark

	if err = s.forwardRepo.Update(forward); err != nil {
		return nil, err
	}

	// 记录操作日志
	s.logService.Record(
		userID,
		username,
		model.ActionUpdate,
		model.ResourceTypeForward,
		forward.ID,
		fmt.Sprintf("更新转发规则: %s", forward.Name),
		ip,
		userAgent)

	return forward, nil
}

// Delete 删除转发规则
// 如果规则正在运行，会先尝试停止
func (s *ForwardService) Delete(id uint, userID uint, username string, ip, userAgent string) error {
	// 查询转发规则
	forward, err := s.forwardRepo.FindByID(id)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrForwardNotFound
		}
		return err
	}

	// 如果正在运行，先停止
	if forward.Status == model.ForwardStatusRunning {
		if err = s.Stop(id, userID, username, ip, userAgent); err != nil {
			logger.Warnf("停止转发规则失败: %v", err)
		}
	}

	// 删除转发规则
	if err = s.forwardRepo.Delete(id); err != nil {
		return err
	}

	// 记录操作日志
	s.logService.Record(
		userID,
		username,
		model.ActionDelete,
		model.ResourceTypeForward,
		id,
		fmt.Sprintf("删除转发规则: %s", forward.Name),
		ip,
		userAgent)

	logger.Infof("删除转发规则成功: %s", forward.Name)
	return nil
}

// GetByID 获取转发规则详情
func (s *ForwardService) GetByID(id uint) (*model.GostForward, error) {
	forward, err := s.forwardRepo.FindByID(id)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrForwardNotFound
		}
		return nil, err
	}
	return forward, nil
}

// List 获取转发规则列表
func (s *ForwardService) List(req *dto.ForwardListReq) ([]model.GostForward, int64, error) {
	// 设置默认值
	req.SetDefaults()

	opt := &repository.QueryOption{
		Pagination: &repository.Pagination{
			Page:     req.Page,
			PageSize: req.PageSize,
		},
		Conditions: make(map[string]any),
	}

	if req.NodeID > 0 {
		opt.Conditions["node_id = ?"] = req.NodeID
	}
	if req.Status != "" {
		opt.Conditions["status = ?"] = req.Status
	}
	if req.Keyword != "" {
		opt.Conditions["name LIKE ?"] = []interface{}{
			"%" + req.Keyword + "%",
		}
	}

	return s.forwardRepo.List(opt)
}

// Start 启动转发
// 通过 Gost API 在目标节点上创建转发服务
func (s *ForwardService) Start(id uint, userID uint, username string, ip, userAgent string) error {
	forward, err := s.forwardRepo.FindByID(id)
	if err != nil {
		return err
	}

	// 已在运行中则跳过
	if forward.Status == model.ForwardStatusRunning {
		return nil
	}

	// 获取节点并创建客户端
	node, err := s.nodeRepo.FindByID(forward.NodeID)
	if err != nil {
		return err
	}
	if node.Status == model.NodeStatusOffline {
		return errors.ErrNodeOffline
	}

	client := utils.GetGostClient(node)

	// 构建服务配置
	serviceName := fmt.Sprintf("forward-%d", forward.ID)
	// 构建目标列表
	targets := forward.Targets

	// 默认策略
	strategy := forward.Strategy
	if strategy == "" || len(targets) == 1 {
		strategy = "round"
	}

	var svc *gost.ServiceConfig
	if forward.Protocol == model.ForwardProtocolTCP {
		svc = gost.BuildTCPForwardService(serviceName, forward.ListenPort, targets, strategy)
	} else {
		svc = gost.BuildUDPForwardService(serviceName, forward.ListenPort, targets, strategy)
	}

	// 创建观察器 (使用 helper)
	observerName, err := CreateObserver(client, s.sysRepo, node.Name, forward.ID)
	if err != nil {
		return err
	}
	_ = s.forwardRepo.UpdateObserverID(id, observerName)

	if observerName != "" {
		svc.Observer = observerName
		// 启用统计功能
		if svc.Metadata == nil {
			svc.Metadata = make(map[string]any)
		}
		svc.Metadata["enableStats"] = true
		svc.Metadata["observer.period"] = "5s"
		svc.Metadata["observer.resetTraffic"] = true
	}

	// 创建服务
	if err = client.CreateService(svc); err != nil {
		_ = s.forwardRepo.UpdateStatus(id, model.ForwardStatusError)
		return errors.ErrForwardStartFailed
	}

	// 保存配置
	_ = client.SaveConfig()

	// 更新状态
	_ = s.forwardRepo.UpdateStatus(id, model.ForwardStatusRunning)
	_ = s.forwardRepo.UpdateServiceID(id, serviceName)

	client.SaveConfig()

	// 记录操作日志
	s.logService.Record(
		userID,
		username,
		model.ActionStart,
		model.ResourceTypeForward,
		id,
		fmt.Sprintf("启动转发规则: %s", forward.Name),
		ip,
		userAgent)

	logger.Infof("启动转发成功: %s", forward.Name)
	return nil
}

// Stop 停止转发
// 通过 Gost API 在目标节点上删除转发服务
func (s *ForwardService) Stop(id uint, userID uint, username string, ip, userAgent string) error {
	forward, err := s.forwardRepo.FindByID(id)
	if err != nil {
		return err
	}

	// 未运行则跳过
	if forward.Status != model.ForwardStatusRunning {
		return nil
	}

	// 获取节点并创建客户端
	node, err := s.nodeRepo.FindByID(forward.NodeID)
	if err != nil {
		return err
	}

	if node.Status == model.NodeStatusOffline {
		return errors.ErrNodeOffline
	}

	client := utils.GetGostClient(node)

	// 删除服务
	if forward.ServiceID != "" {
		if err = client.DeleteService(forward.ServiceID); err != nil {
			logger.Warnf("删除 Gost 服务失败: %v", err)
		}
	}

	// 更新状态
	_ = s.forwardRepo.UpdateStatus(id, model.ForwardStatusStopped)

	// 保存配置
	_ = client.SaveConfig()

	// 记录操作日志
	s.logService.Record(
		userID,
		username,
		model.ActionStop,
		model.ResourceTypeForward,
		id,
		fmt.Sprintf("停止转发规则: %s", forward.Name),
		ip,
		userAgent)

	logger.Infof("停止转发成功: %s", forward.Name)
	return nil
}

// GetStats 获取转发统计
func (s *ForwardService) GetStats() (map[string]int64, error) {
	total, err := s.forwardRepo.CountAll()
	if err != nil {
		return nil, err
	}

	running, err := s.forwardRepo.CountByStatus(model.ForwardStatusRunning)
	if err != nil {
		return nil, err
	}

	return map[string]int64{
		"total":   total,
		"running": running,
		"stopped": total - running,
	}, nil
}
