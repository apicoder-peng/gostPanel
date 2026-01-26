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
// 负责多跳隧道的 CRUD 操作及启停控制
// 流量路径: 客户端 -> 入口节点(EntryNode) -> 出口节点(ExitNode) -> 目标地址
type TunnelService struct {
	tunnelRepo *repository.TunnelRepository
	nodeRepo   *repository.NodeRepository
	sysRepo    *repository.SystemConfigRepository
	logService *LogService
}

// NewTunnelService 创建隧道服务
func NewTunnelService(db *gorm.DB) *TunnelService {
	return &TunnelService{
		tunnelRepo: repository.NewTunnelRepository(db),
		nodeRepo:   repository.NewNodeRepository(db),
		sysRepo:    repository.NewSystemConfigRepository(db),
		logService: NewLogService(db),
	}
}

// Create 创建隧道
// 创建前会检查入口/出口节点是否存在，以及端口是否被占用
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

	// 检查端口是否已被使用
	exists, err := s.tunnelRepo.ExistsByPort(req.EntryNodeID, req.EntryPort)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.ErrPortInUse
	}

	// 使用出口节点配置的 Relay 端口
	relayPort := exitNode.RelayPort
	if relayPort == 0 {
		return nil, errors.ErrExitNodeNoRelayPort
	}

	// 创建隧道
	tunnel := &model.GostTunnel{
		Name:        req.Name,
		EntryNodeID: req.EntryNodeID,
		ExitNodeID:  req.ExitNodeID,
		Protocol:    req.Protocol,
		ListenPort:  req.EntryPort,

		Targets:   req.Targets,
		Strategy:  req.Strategy,
		RelayPort: relayPort,
		Remark:    req.Remark,
		Status:    model.TunnelStatusStopped,
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

// Update 更新隧道
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

	// 检查入口和出口是否相同
	if req.EntryNodeID == req.ExitNodeID {
		return nil, errors.ErrTunnelNodeSame
	}

	// 检查端口是否已被使用（排除自身）
	exists, err := s.tunnelRepo.ExistsByPort(req.EntryNodeID, req.EntryPort, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.ErrPortInUse
	}

	// 获取出口节点以获取 Relay 端口
	exitNode, err := s.nodeRepo.FindByID(req.ExitNodeID)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrExitNodeNotFound
		}
		return nil, err
	}

	// 使用出口节点配置的 Relay 端口
	relayPort := exitNode.RelayPort
	if relayPort == 0 {
		return nil, errors.ErrExitNodeNoRelayPort
	}

	// 更新隧道
	tunnel.Name = req.Name
	tunnel.EntryNodeID = req.EntryNodeID
	tunnel.ExitNodeID = req.ExitNodeID
	tunnel.Protocol = req.Protocol
	tunnel.ListenPort = req.EntryPort

	tunnel.Targets = req.Targets
	tunnel.Strategy = req.Strategy
	tunnel.RelayPort = relayPort
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
func (s *TunnelService) Delete(id uint, userID uint, username string, ip, userAgent string) error {
	tunnel, err := s.tunnelRepo.FindByID(id)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrTunnelNotFound
		}
		return err
	}

	// 如果正在运行，先停止
	if tunnel.Status == model.TunnelStatusRunning {
		if err = s.Stop(id, userID, username, ip, userAgent); err != nil {
			logger.Warnf("停止隧道失败: %v", err)
		}
	}

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
	// 设置默认值
	req.SetDefaults()

	opt := &repository.QueryOption{
		Pagination: &repository.Pagination{
			Page:     req.Page,
			PageSize: req.PageSize,
		},
		Conditions: make(map[string]any),
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
// 在入口节点创建服务：监听端口 -> 通过链(chain)连接出口节点 -> 转发到目标
func (s *TunnelService) Start(id uint, userID uint, username string, ip, userAgent string) error {
	tunnel, err := s.tunnelRepo.FindByID(id)
	if err != nil {
		return err
	}

	if tunnel.Status == model.TunnelStatusRunning {
		return nil
	}

	// 获取入口和出口节点
	entryNode, err := s.nodeRepo.FindByID(tunnel.EntryNodeID)
	if err != nil {
		return errors.ErrOperationFailed
	}
	if entryNode.Status == model.NodeStatusOffline {
		return errors.ErrNodeOffline
	}

	exitNode, err := s.nodeRepo.FindByID(tunnel.ExitNodeID)
	if err != nil {
		return errors.ErrOperationFailed
	}
	if exitNode.Status == model.NodeStatusOffline {
		return errors.ErrNodeOffline
	}

	// 从出口节点 API URL 中提取主机 IP
	exitHost := utils.ExtractHost(exitNode.APIURL)
	if exitHost == "" {
		_ = s.tunnelRepo.UpdateStatus(id, model.TunnelStatusError)
		return errors.ErrExtractHostFailed
	}

	// 构建 relay 服务地址
	relayAddr := fmt.Sprintf("%s:%d", exitHost, tunnel.RelayPort)
	logger.Infof("隧道 relay 地址: %s", relayAddr)

	// 在入口节点创建 chain 连接到出口节点的 relay 服务
	entryClient := utils.GetGostClient(entryNode)

	chainName := fmt.Sprintf("chain-tunnel-%d", tunnel.ID)
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
							Type: "tcp",
						},
					},
				},
			},
		},
	}

	if err = entryClient.CreateChain(chain); err != nil {
		_ = s.tunnelRepo.UpdateStatus(id, model.TunnelStatusError)
		return errors.ErrTunnelChainCreateFailed
	}

	// 在入口节点创建服务，通过链转发到目标
	serviceName := fmt.Sprintf("tunnel-%d", tunnel.ID)

	// 构建目标列表
	targets := tunnel.Targets
	nodes := make([]*gost.ForwarderNode, 0)
	for i, t := range targets {
		nodes = append(nodes, &gost.ForwarderNode{
			Name: fmt.Sprintf("target-%d", i),
			Addr: t,
		})
	}

	// 默认策略
	strategy := tunnel.Strategy
	if strategy == "" || len(targets) == 1 {
		strategy = "round"
	}

	var svc *gost.ServiceConfig
	if model.ForwardProtocol(tunnel.Protocol) == model.ForwardProtocolTCP {
		svc = gost.BuildTCPForwardService(serviceName, tunnel.ListenPort, targets, strategy)
	} else {
		svc = gost.BuildUDPForwardService(serviceName, tunnel.ListenPort, targets, strategy)
	}

	logger.Infof("准备创建服务: %+v, Handler: %+v, Chain: %s", svc, svc.Handler, svc.Handler.Chain)

	// 创建观察器 (使用 helper)
	observerName, err := CreateObserver(entryClient, s.sysRepo, entryNode.Name, tunnel.ID)
	if err != nil {
		return err
	}
	_ = s.tunnelRepo.UpdateObserverID(id, observerName)

	if observerName != "" {
		svc.Observer = observerName
		if svc.Metadata == nil {
			svc.Metadata = make(map[string]any)
		}
		svc.Metadata["enableStats"] = true
		svc.Metadata["observer.period"] = "5s"
		svc.Metadata["observer.resetTraffic"] = true
	}

	if err = entryClient.CreateService(svc); err != nil {
		_ = entryClient.DeleteChain(chainName)
		_ = s.tunnelRepo.UpdateStatus(id, model.TunnelStatusError)
		return errors.ErrTunnelServiceCreateFailed
	}

	// 保存配置
	_ = entryClient.SaveConfig()

	// 更新状态
	_ = s.tunnelRepo.UpdateStatus(id, model.TunnelStatusRunning)
	_ = s.tunnelRepo.UpdateServiceInfo(id, serviceName, chainName)

	s.logService.Record(
		userID,
		username,
		model.ActionStart,
		model.ResourceTypeTunnel,
		id,
		fmt.Sprintf("启动隧道: %s", tunnel.Name),
		ip,
		userAgent)

	logger.Infof("启动隧道成功: %s (%s -> %s)", tunnel.Name, entryNode.Name, exitNode.Name)
	return nil
}

// Stop 停止隧道
func (s *TunnelService) Stop(id uint, userID uint, username string, ip, userAgent string) error {
	tunnel, err := s.tunnelRepo.FindByID(id)
	if err != nil {
		return err
	}

	if tunnel.Status != model.TunnelStatusRunning {
		return nil
	}

	// 获取入口节点
	entryNode, err := s.nodeRepo.FindByID(tunnel.EntryNodeID)
	if err != nil {
		return err
	}

	if entryNode.Status == model.NodeStatusOffline {
		return errors.ErrNodeOffline
	}

	entryClient := utils.GetGostClient(entryNode)

	// 删除入口节点上的服务
	if tunnel.ServiceID != "" {
		if err = entryClient.DeleteService(tunnel.ServiceID); err != nil {
			logger.Warnf("删除入口节点服务失败: %v", err)
		}
	}

	// 删除入口节点上的链
	if tunnel.ChainID != "" {
		if err = entryClient.DeleteChain(tunnel.ChainID); err != nil {
			logger.Warnf("删除入口节点链失败: %v", err)
		}
	}

	// 保存配置
	_ = entryClient.SaveConfig()

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

// GetStats 获取隧道统计
func (s *TunnelService) GetStats() (map[string]int64, error) {
	total, err := s.tunnelRepo.CountAll()
	if err != nil {
		return nil, err
	}

	return map[string]int64{
		"total": total,
	}, nil
}
