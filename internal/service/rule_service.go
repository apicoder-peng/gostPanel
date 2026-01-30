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

// RuleService 规则服务
// 入口选择：NodeID 或 TunnelID 二选一
// - 端口转发 (forward)：选择 NodeID，直接在该节点上创建转发服务
// - 隧道转发 (tunnel)：选择 TunnelID，在隧道的入口节点上创建转发服务，使用隧道的 Chain
type RuleService struct {
	ruleRepo      *repository.RuleRepository
	nodeRepo      *repository.NodeRepository
	tunnelRepo    *repository.TunnelRepository
	sysRepo       *repository.SystemConfigRepository
	logService    *LogService
	tunnelService *TunnelService
}

// NewRuleService 创建规则服务
func NewRuleService(db *gorm.DB) *RuleService {
	return &RuleService{
		ruleRepo:      repository.NewRuleRepository(db),
		nodeRepo:      repository.NewNodeRepository(db),
		tunnelRepo:    repository.NewTunnelRepository(db),
		sysRepo:       repository.NewSystemConfigRepository(db),
		logService:    NewLogService(db),
		tunnelService: NewTunnelService(db),
	}
}

// Create 创建规则
// 根据类型验证入口：端口转发需要 NodeID，隧道转发需要 TunnelID
func (s *RuleService) Create(req *dto.CreateRuleReq, userID uint, username string, ip, userAgent string) (*model.GostRule, error) {
	var entryNodeID uint

	// 根据规则类型验证入口
	if req.Type == string(model.RuleTypeForward) {
		// 端口转发：需要 NodeID
		if req.NodeID == nil || *req.NodeID == 0 {
			return nil, errors.ErrNodeRequired
		}
		// 检查节点是否存在
		_, err := s.nodeRepo.FindByID(*req.NodeID)
		if err != nil {
			if stderrors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.ErrNodeNotFound
			}
			return nil, err
		}
		entryNodeID = *req.NodeID
	} else if req.Type == string(model.RuleTypeTunnel) {
		// 隧道转发：需要 TunnelID
		if req.TunnelID == nil || *req.TunnelID == 0 {
			return nil, errors.ErrTunnelRequired
		}
		// 检查隧道是否存在
		tunnel, err := s.tunnelRepo.FindByID(*req.TunnelID)
		if err != nil {
			if stderrors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.ErrTunnelNotFound
			}
			return nil, err
		}
		// 使用隧道的入口节点
		entryNodeID = tunnel.EntryNodeID
	} else {
		return nil, errors.ErrRuleTypeInvalid
	}

	// 检查端口是否已被使用
	exists, err := s.ruleRepo.ExistsByPort(entryNodeID, req.ListenPort)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.ErrRulePortExists
	}

	// 创建规则
	rule := &model.GostRule{
		NodeID:     req.NodeID,
		TunnelID:   req.TunnelID,
		Name:       req.Name,
		Type:       model.RuleType(req.Type),
		Protocol:   model.RuleProtocol(req.Protocol),
		ListenPort: req.ListenPort,
		Targets:    req.Targets,
		Strategy:   req.Strategy,
		EnableTLS:  req.EnableTLS,
		Remark:     req.Remark,
		Status:     model.RuleStatusStopped,
	}

	if err = s.ruleRepo.Create(rule); err != nil {
		return nil, err
	}

	// 记录操作日志
	s.logService.Record(
		userID,
		username,
		model.ActionCreate,
		model.ResourceTypeRule,
		rule.ID,
		fmt.Sprintf("创建规则: %s (类型: %s)", rule.Name, rule.Type),
		ip,
		userAgent)

	logger.Infof("创建规则成功: %s (:%d)", rule.Name, rule.ListenPort)
	return rule, nil
}

// Update 更新规则（不能修改类型和入口）
func (s *RuleService) Update(id uint, req *dto.UpdateRuleReq, userID uint, username string, ip, userAgent string) (*model.GostRule, error) {
	rule, err := s.ruleRepo.FindByID(id)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrRuleNotFound
		}
		return nil, err
	}

	// 检查是否在运行中
	if rule.Status == model.RuleStatusRunning {
		return nil, errors.ErrRuleRunning
	}

	// 获取入口节点 ID（用于端口冲突检查）
	entryNodeID := s.getEntryNodeID(rule)

	// 检查端口是否已被使用（排除自身）
	exists, err := s.ruleRepo.ExistsByPort(entryNodeID, req.ListenPort, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.ErrRulePortExists
	}

	// 更新规则（不修改类型和入口）
	rule.Name = req.Name
	rule.Protocol = model.RuleProtocol(req.Protocol)
	rule.ListenPort = req.ListenPort
	rule.Targets = req.Targets
	rule.Strategy = req.Strategy
	rule.EnableTLS = req.EnableTLS
	rule.Remark = req.Remark

	if err = s.ruleRepo.Update(rule); err != nil {
		return nil, err
	}

	s.logService.Record(
		userID,
		username,
		model.ActionUpdate,
		model.ResourceTypeRule,
		rule.ID,
		fmt.Sprintf("更新规则: %s", rule.Name),
		ip,
		userAgent)

	return rule, nil
}

// Delete 删除规则
func (s *RuleService) Delete(id uint, userID uint, username string, ip, userAgent string) error {
	rule, err := s.ruleRepo.FindByID(id)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrRuleNotFound
		}
		return err
	}

	// 如果正在运行，先停止
	if rule.Status == model.RuleStatusRunning {
		if err = s.Stop(id, userID, username, ip, userAgent); err != nil {
			logger.Warnf("停止规则失败: %v", err)
		}
	}

	// 删除规则
	if err = s.ruleRepo.Delete(id); err != nil {
		return err
	}

	s.logService.Record(
		userID,
		username,
		model.ActionDelete,
		model.ResourceTypeRule,
		id,
		fmt.Sprintf("删除规则: %s", rule.Name),
		ip,
		userAgent)

	logger.Infof("删除规则成功: %s", rule.Name)
	return nil
}

// GetByID 获取规则详情
func (s *RuleService) GetByID(id uint) (*model.GostRule, error) {
	rule, err := s.ruleRepo.FindByID(id)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrRuleNotFound
		}
		return nil, err
	}
	return rule, nil
}

// List 获取规则列表
func (s *RuleService) List(req *dto.RuleListReq) ([]model.GostRule, int64, error) {
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
	if req.TunnelID > 0 {
		opt.Conditions["tunnel_id = ?"] = req.TunnelID
	}
	if req.Type != "" {
		opt.Conditions["type = ?"] = req.Type
	}
	if req.Status != "" {
		opt.Conditions["status = ?"] = req.Status
	}
	if req.Keyword != "" {
		opt.Conditions["name LIKE ?"] = []interface{}{
			"%" + req.Keyword + "%",
		}
	}

	return s.ruleRepo.List(opt)
}

// Start 启动规则
func (s *RuleService) Start(id uint, userID uint, username string, ip, userAgent string) error {
	rule, err := s.ruleRepo.FindByID(id)
	if err != nil {
		return err
	}

	// 已在运行中则跳过
	if rule.Status == model.RuleStatusRunning {
		return nil
	}

	// 获取入口节点
	entryNodeID := s.getEntryNodeID(rule)
	node, err := s.nodeRepo.FindByID(entryNodeID)
	if err != nil {
		return errors.ErrNodeNotFound
	}
	if node.Status == model.NodeStatusOffline {
		return errors.ErrNodeOffline
	}

	client := utils.GetGostClient(node)
	serviceName := fmt.Sprintf("rule-%d", rule.ID)

	// 根据规则类型处理
	if rule.Type == model.RuleTypeTunnel {
		if err = s.startTunnelRule(rule, client, serviceName); err != nil {
			return err
		}
	} else {
		if err = s.startForwardRule(rule, client, serviceName); err != nil {
			return err
		}
	}

	s.logService.Record(
		userID,
		username,
		model.ActionStart,
		model.ResourceTypeRule,
		id,
		fmt.Sprintf("启动规则: %s", rule.Name),
		ip,
		userAgent)

	logger.Infof("启动规则成功: %s", rule.Name)
	return nil
}

// startForwardRule 启动端口转发规则（直连目标）
func (s *RuleService) startForwardRule(rule *model.GostRule, client *gost.Client, serviceName string) error {
	// 端口转发没有 Chain ID
	return s.buildAndStartService(client, rule, serviceName, "")
}

// startTunnelRule 启动隧道转发规则（通过隧道链路）
func (s *RuleService) startTunnelRule(rule *model.GostRule, client *gost.Client, serviceName string) error {
	if rule.TunnelID == nil {
		return errors.ErrTunnelRequired
	}

	// 获取隧道信息
	tunnel, err := s.tunnelRepo.FindByID(*rule.TunnelID)
	if err != nil {
		return errors.ErrTunnelNotFound
	}

	// 确保隧道已启动
	if tunnel.Status != model.TunnelStatusRunning {
		return errors.ErrTunnelNotRunning
	}

	// 检查隧道是否有 Chain ID
	if tunnel.ChainID == "" {
		_ = s.ruleRepo.UpdateStatus(rule.ID, model.RuleStatusError)
		return errors.ErrTunnelChainNotFound
	}

	// 使用通用逻辑启动服务，传入 Chain ID
	return s.buildAndStartService(client, rule, serviceName, tunnel.ChainID)
}

// Stop 停止规则
func (s *RuleService) Stop(id uint, userID uint, username string, ip, userAgent string) error {
	rule, err := s.ruleRepo.FindByID(id)
	if err != nil {
		return err
	}

	if rule.Status != model.RuleStatusRunning {
		return nil
	}

	// 获取入口节点
	entryNodeID := s.getEntryNodeID(rule)
	node, err := s.nodeRepo.FindByID(entryNodeID)
	if err != nil {
		// 节点不存在，直接更新状态
		_ = s.ruleRepo.UpdateStatus(id, model.RuleStatusStopped)
		return nil
	}

	if node.Status == model.NodeStatusOffline {
		// 节点离线，直接更新状态
		_ = s.ruleRepo.UpdateStatus(id, model.RuleStatusStopped)
		return nil
	}

	client := utils.GetGostClient(node)

	// 删除服务
	if rule.ServiceID != "" {
		if err = client.DeleteService(rule.ServiceID); err != nil {
			logger.Warnf("删除 Gost 服务失败: %v", err)
		}
	}

	_ = s.ruleRepo.UpdateStatus(id, model.RuleStatusStopped)
	_ = client.SaveConfig()

	s.logService.Record(
		userID,
		username,
		model.ActionStop,
		model.ResourceTypeRule,
		id,
		fmt.Sprintf("停止规则: %s", rule.Name),
		ip,
		userAgent)

	logger.Infof("停止规则成功: %s", rule.Name)
	return nil
}

// getEntryNodeID 获取规则的入口节点 ID
func (s *RuleService) getEntryNodeID(rule *model.GostRule) uint {
	if rule.Type == model.RuleTypeTunnel && rule.TunnelID != nil {
		// 隧道转发：使用隧道的入口节点
		nodeID, err := s.tunnelService.GetEntryNodeID(*rule.TunnelID)
		if err == nil {
			return nodeID
		}
	}
	// 端口转发：使用规则的 NodeID
	if rule.NodeID != nil {
		return *rule.NodeID
	}
	return 0
}

// setupRuleObserver 配置规则的观察器
func (s *RuleService) setupRuleObserver(client *gost.Client, rule *model.GostRule, svc *gost.ServiceConfig) error {
	// 确保全局观察器存在
	observerName, err := EnsureGlobalObserver(client, s.sysRepo)
	if err != nil {
		return err
	}

	// 更新规则关联的 ObserverID
	_ = s.ruleRepo.UpdateObserverID(rule.ID, observerName)

	// 配置服务的观察器参数
	if observerName != "" {
		svc.Observer = observerName
		if svc.Metadata == nil {
			svc.Metadata = make(map[string]any)
		}
		svc.Metadata["enableStats"] = true
		svc.Metadata["observer.period"] = "5s"
		svc.Metadata["observer.resetTraffic"] = false
	}
	return nil
}

// buildAndStartService 构建并启动 Gost 服务 (处理通用逻辑)
func (s *RuleService) buildAndStartService(client *gost.Client, rule *model.GostRule, serviceName string, chainID string) error {
	targets := rule.Targets
	strategy := rule.Strategy
	if strategy == "" || len(targets) == 1 {
		strategy = "round"
	}

	var svc *gost.ServiceConfig
	if rule.Protocol == model.RuleProtocolTCP {
		svc = gost.BuildTCPForwardService(serviceName, rule.ListenPort, targets, strategy)
	} else {
		svc = gost.BuildUDPForwardService(serviceName, rule.ListenPort, targets, strategy)
	}

	// 如果有 Chain ID，则关联（用于隧道转发）
	if chainID != "" {
		svc.Handler.Chain = chainID
	}

	// 配置观察器
	if err := s.setupRuleObserver(client, rule, svc); err != nil {
		return err
	}

	if err := client.CreateService(svc); err != nil {
		_ = s.ruleRepo.UpdateStatus(rule.ID, model.RuleStatusError)
		return errors.ErrRuleStartFailed
	}

	_ = client.SaveConfig()
	_ = s.ruleRepo.UpdateStatus(rule.ID, model.RuleStatusRunning)
	_ = s.ruleRepo.UpdateServiceID(rule.ID, serviceName)

	return nil
}
