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

// TunnelService 隧道服务
// 负责隧道的 CRUD 操作及启停控制
// 启动隧道时：在出口节点创建 Relay 服务，在入口节点创建 Chain 连接到出口节点
type TunnelService struct {
	tunnelRepo *repository.TunnelRepository
	nodeRepo   *repository.NodeRepository
	logService *LogService
}

// NewTunnelService 创建隧道服务
func NewTunnelService(db *gorm.DB) *TunnelService {
	return &TunnelService{
		tunnelRepo: repository.NewTunnelRepository(db),
		nodeRepo:   repository.NewNodeRepository(db),
		logService: NewLogService(db),
	}
}

// Create 创建隧道
func (s *TunnelService) Create(req *dto.CreateTunnelReq, userID uint, username string, ip, userAgent string) (*model.GostTunnel, error) {
	// 检查入口和出口是否相同
	if req.EntryNodeID == req.ExitNodeID {
		return nil, errors.ErrTunnelNodeSame
	}

	// 检查入口节点是否存在
	entryNode, err := s.nodeRepo.FindByID(req.EntryNodeID)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrEntryNodeNotFound
		}
		return nil, err
	}

	// 检查出口节点是否存在
	exitNode, err := s.nodeRepo.FindByID(req.ExitNodeID)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrExitNodeNotFound
		}
		return nil, err
	}

	// 创建隧道
	tunnel := &model.GostTunnel{
		Name:        req.Name,
		EntryNodeID: req.EntryNodeID,
		ExitNodeID:  req.ExitNodeID,
		Protocol:    req.Protocol,
		RelayPort:   req.RelayPort,
		Remark:      req.Remark,
		Status:      model.TunnelStatusStopped,
	}

	if err = s.tunnelRepo.Create(tunnel); err != nil {
		return nil, err
	}

	// 记录操作日志
	s.logService.Record(
		userID,
		username,
		model.ActionCreate,
		model.ResourceTypeTunnel,
		tunnel.ID,
		fmt.Sprintf("创建隧道: %s (%s -> %s)", tunnel.Name, entryNode.Name, exitNode.Name),
		ip,
		userAgent)

	logger.Infof("创建隧道成功: %s", tunnel.Name)
	return tunnel, nil
}

// Update 更新隧道（仅支持更新非运行中的隧道，且不能修改入口/出口节点）
func (s *TunnelService) Update(id uint, req *dto.UpdateTunnelReq, userID uint, username string, ip, userAgent string) (*model.GostTunnel, error) {
	tunnel, err := s.tunnelRepo.FindByID(id)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrTunnelNotFound
		}
		return nil, err
	}

	// 检查是否在运行中
	if tunnel.Status == model.TunnelStatusRunning {
		return nil, errors.ErrTunnelRunning
	}

	// 更新隧道（不能修改入口/出口节点）
	tunnel.Name = req.Name
	tunnel.Protocol = req.Protocol
	tunnel.RelayPort = req.RelayPort
	tunnel.Remark = req.Remark

	if err = s.tunnelRepo.Update(tunnel); err != nil {
		return nil, err
	}

	s.logService.Record(
		userID,
		username,
		model.ActionUpdate,
		model.ResourceTypeTunnel,
		tunnel.ID,
		fmt.Sprintf("更新隧道: %s", tunnel.Name),
		ip,
		userAgent)

	return tunnel, nil
}

// Delete 删除隧道
// 如果有规则正在使用此隧道，不允许删除
func (s *TunnelService) Delete(id uint, userID uint, username string, ip, userAgent string) error {
	tunnel, err := s.tunnelRepo.FindByID(id)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrTunnelNotFound
		}
		return err
	}

	// 检查是否有规则正在使用此隧道
	hasRules, err := s.tunnelRepo.HasRules(id)
	if err != nil {
		return err
	}
	if hasRules {
		return errors.ErrTunnelHasRules
	}

	// 如果隧道正在运行，先停止
	if tunnel.Status == model.TunnelStatusRunning {
		if err = s.Stop(id, userID, username, ip, userAgent); err != nil {
			logger.Warnf("停止隧道失败: %v", err)
		}
	}

	// 删除隧道
	if err = s.tunnelRepo.Delete(id); err != nil {
		return err
	}

	s.logService.Record(
		userID,
		username,
		model.ActionDelete,
		model.ResourceTypeTunnel,
		id,
		fmt.Sprintf("删除隧道: %s", tunnel.Name),
		ip,
		userAgent)

	logger.Infof("删除隧道成功: %s", tunnel.Name)
	return nil
}

// GetByID 获取隧道详情
func (s *TunnelService) GetByID(id uint) (*model.GostTunnel, error) {
	tunnel, err := s.tunnelRepo.FindByID(id)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrTunnelNotFound
		}
		return nil, err
	}
	return tunnel, nil
}

// List 获取隧道列表
func (s *TunnelService) List(req *dto.TunnelListReq) ([]model.GostTunnel, int64, error) {
	req.SetDefaults()

	opt := &repository.QueryOption{
		Pagination: &repository.Pagination{
			Page:     req.Page,
			PageSize: req.PageSize,
		},
		Conditions: make(map[string]any),
	}

	if req.NodeID > 0 {
		opt.Conditions["entry_node_id = ? OR exit_node_id = ?"] = []interface{}{req.NodeID, req.NodeID}
	}
	if req.Status != "" {
		opt.Conditions["status = ?"] = req.Status
	}
	if req.Keyword != "" {
		opt.Conditions["name LIKE ?"] = []interface{}{
			"%" + req.Keyword + "%",
		}
	}

	return s.tunnelRepo.List(opt)
}

// Start 启动隧道
// 在出口节点创建 Relay 服务，在入口节点创建 Chain 连接到出口节点
func (s *TunnelService) Start(id uint, userID uint, username string, ip, userAgent string) error {
	tunnel, err := s.tunnelRepo.FindByID(id)
	if err != nil {
		return err
	}

	// 已在运行中则跳过
	if tunnel.Status == model.TunnelStatusRunning {
		return nil
	}

	// 获取入口和出口节点
	entryNode, err := s.nodeRepo.FindByID(tunnel.EntryNodeID)
	if err != nil {
		return errors.ErrEntryNodeNotFound
	}
	exitNode, err := s.nodeRepo.FindByID(tunnel.ExitNodeID)
	if err != nil {
		return errors.ErrExitNodeNotFound
	}

	// 检查节点状态
	if entryNode.Status == model.NodeStatusOffline {
		return errors.ErrEntryNodeOffline
	}
	if exitNode.Status == model.NodeStatusOffline {
		return errors.ErrExitNodeOffline
	}

	// 步骤1：在出口节点创建 Relay 服务
	exitClient := utils.GetGostClient(exitNode)
	relayServiceName := fmt.Sprintf("relay-tunnel-%d", tunnel.ID)

	relaySvc := &gost.ServiceConfig{
		Name: relayServiceName,
		Addr: fmt.Sprintf(":%d", tunnel.RelayPort),
		Handler: &gost.HandlerConfig{
			Type: "relay",
		},
		Listener: &gost.ListenerConfig{
			Type: tunnel.Protocol,
		},
	}

	if err = exitClient.CreateService(relaySvc); err != nil {
		_ = s.tunnelRepo.UpdateStatus(id, model.TunnelStatusError)
		return errors.ErrTunnelRelayCreateFailed
	}

	// 保存出口节点配置
	_ = exitClient.SaveConfig()

	// 步骤2：在入口节点创建 Chain 连接到出口节点的 Relay 服务
	entryClient := utils.GetGostClient(entryNode)

	// 从出口节点配置中获取主机 IP
	exitHost := exitNode.Address
	if exitHost == "" {
		// 回滚：删除出口节点的 Relay 服务
		_ = exitClient.DeleteService(relayServiceName)
		_ = exitClient.SaveConfig()
		_ = s.tunnelRepo.UpdateStatus(id, model.TunnelStatusError)
		return errors.ErrExtractHostFailed
	}

	chainName := fmt.Sprintf("tunnel-%d-chain", tunnel.ID)
	relayAddr := fmt.Sprintf("%s:%d", exitHost, tunnel.RelayPort)

	chain := &gost.ChainConfig{
		Name: chainName,
		Hops: []*gost.HopConfig{
			{
				Name: "hop-0",
				Nodes: []*gost.NodeConfig{
					{
						Name: "exit-relay",
						Addr: relayAddr,
						Connector: &gost.ConnectorConfig{
							Type: "relay",
						},
						Dialer: &gost.DialerConfig{
							Type: tunnel.Protocol,
						},
					},
				},
			},
		},
	}

	if err = entryClient.CreateChain(chain); err != nil {
		// 回滚：删除出口节点的 Relay 服务
		_ = exitClient.DeleteService(relayServiceName)
		_ = exitClient.SaveConfig()
		_ = s.tunnelRepo.UpdateStatus(id, model.TunnelStatusError)
		return errors.ErrTunnelChainCreateFailed
	}

	// 保存入口节点配置
	_ = entryClient.SaveConfig()

	// 更新隧道状态和服务 ID
	_ = s.tunnelRepo.UpdateServiceInfo(id, relayServiceName, chainName)
	_ = s.tunnelRepo.UpdateStatus(id, model.TunnelStatusRunning)

	s.logService.Record(
		userID,
		username,
		model.ActionStart,
		model.ResourceTypeTunnel,
		id,
		fmt.Sprintf("启动隧道: %s", tunnel.Name),
		ip,
		userAgent)

	logger.Infof("启动隧道成功: %s (Relay: %s -> Chain: %s)", tunnel.Name, relayServiceName, chainName)
	return nil
}

// Stop 停止隧道
// 删除入口节点的 Chain 和出口节点的 Relay 服务
func (s *TunnelService) Stop(id uint, userID uint, username string, ip, userAgent string) error {
	tunnel, err := s.tunnelRepo.FindByID(id)
	if err != nil {
		return err
	}

	// 未运行则跳过
	if tunnel.Status != model.TunnelStatusRunning {
		return nil
	}

	// 获取入口和出口节点
	entryNode, _ := s.nodeRepo.FindByID(tunnel.EntryNodeID)
	exitNode, _ := s.nodeRepo.FindByID(tunnel.ExitNodeID)

	// 步骤1：删除入口节点的 Chain
	if entryNode != nil && entryNode.Status == model.NodeStatusOnline && tunnel.ChainID != "" {
		entryClient := utils.GetGostClient(entryNode)
		if err = entryClient.DeleteChain(tunnel.ChainID); err != nil {
			logger.Warnf("删除隧道 Chain 失败: %v", err)
		}
		_ = entryClient.SaveConfig()
	}

	// 步骤2：删除出口节点的 Relay 服务
	if exitNode != nil && exitNode.Status == model.NodeStatusOnline && tunnel.ServiceID != "" {
		exitClient := utils.GetGostClient(exitNode)
		if err = exitClient.DeleteService(tunnel.ServiceID); err != nil {
			logger.Warnf("删除隧道 Relay 服务失败: %v", err)
		}
		_ = exitClient.SaveConfig()
	}

	// 更新状态
	_ = s.tunnelRepo.UpdateStatus(id, model.TunnelStatusStopped)

	s.logService.Record(
		userID,
		username,
		model.ActionStop,
		model.ResourceTypeTunnel,
		id,
		fmt.Sprintf("停止隧道: %s", tunnel.Name),
		ip,
		userAgent)

	logger.Infof("停止隧道成功: %s", tunnel.Name)
	return nil
}

// GetChainID 获取隧道的 Chain ID（供规则服务使用）
func (s *TunnelService) GetChainID(tunnelID uint) (string, error) {
	tunnel, err := s.tunnelRepo.FindByID(tunnelID)
	if err != nil {
		return "", err
	}
	return tunnel.ChainID, nil
}

// GetEntryNodeID 获取隧道的入口节点 ID（供规则服务使用）
func (s *TunnelService) GetEntryNodeID(tunnelID uint) (uint, error) {
	tunnel, err := s.tunnelRepo.FindByID(tunnelID)
	if err != nil {
		return 0, err
	}
	return tunnel.EntryNodeID, nil
}
