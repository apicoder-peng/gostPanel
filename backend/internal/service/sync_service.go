package service

import (
	"fmt"
	"sync"
	"time"

	"gost-panel/internal/model"
	"gost-panel/internal/repository"
	"gost-panel/internal/utils"
	"gost-panel/pkg/logger"

	"gorm.io/gorm"
)

// RuleSyncService 规则状态同步服务
// 定时从 Gost 节点同步转发和隧道规则的真实运行状态
type RuleSyncService struct {
	nodeRepo    *repository.NodeRepository
	forwardRepo *repository.ForwardRepository
	tunnelRepo  *repository.TunnelRepository
	ticker      *time.Ticker
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

// NewRuleSyncService 创建规则状态同步服务
func NewRuleSyncService(db *gorm.DB) *RuleSyncService {
	return &RuleSyncService{
		nodeRepo:    repository.NewNodeRepository(db),
		forwardRepo: repository.NewForwardRepository(db),
		tunnelRepo:  repository.NewTunnelRepository(db),
		stopChan:    make(chan struct{}),
	}
}

// Start 启动定时同步任务（每 5 秒）
func (s *RuleSyncService) Start() {
	s.ticker = time.NewTicker(5 * time.Second)
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()
		logger.Info("规则状态同步服务已启动 (5s 间隔)")

		// 立即执行一次
		s.syncAll()

		for {
			select {
			case <-s.ticker.C:
				s.syncAll()
			case <-s.stopChan:
				logger.Info("规则状态同步服务已停止")
				return
			}
		}
	}()
}

// Stop 停止同步服务
func (s *RuleSyncService) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	close(s.stopChan)
	s.wg.Wait()
}

// syncAll 同步所有节点的规则状态
func (s *RuleSyncService) syncAll() {
	nodes, _, err := s.nodeRepo.List(nil)
	if err != nil {
		logger.Errorf("[Sync] 获取节点列表失败: %v", err)
		return
	}

	for _, node := range nodes {
		// 并发同步每个节点
		go s.syncNodeRules(node)
	}
}

// syncNodeRules 同步单个节点的规则
func (s *RuleSyncService) syncNodeRules(node model.GostNode) {
	// 如果节点离线，跳过规则同步（由 NodeHealthService 处理离线逻辑）
	if node.Status == model.NodeStatusOffline {
		return
	}

	client := utils.GetGostClient(&node)

	// 获取节点真实运行配置
	gostCfg, err := client.GetConfig()
	if err != nil {
		logger.Debugf("[Sync] 获取节点 %d (%s) 配置失败: %v", node.ID, node.Name, err)
		return
	}

	// 提取节点上的 Service 状态
	serviceStates := make(map[string]string)
	for _, svc := range gostCfg.Services {
		state := "stopped"
		if svc.Status != nil {
			state = svc.Status.State
		}
		serviceStates[svc.Name] = state
	}

	runningChains := make(map[string]bool)
	for _, chain := range gostCfg.Chains {
		runningChains[chain.Name] = true
	}

	// 1. 同步转发规则 (Forwarding)
	forwards, err := s.forwardRepo.FindByNodeID(node.ID)
	if err != nil {
		logger.Errorf("[Sync] 获取节点 %d 转发规则失败: %v", node.ID, err)
	} else {
		for _, f := range forwards {
			s.syncForwardStatus(f, serviceStates)
		}
	}

	// 2. 同步隧道规则 (Tunnel)
	// 注意：隧道运行在入口节点上，所以仅在入口节点同步其状态
	tunnels, err := s.tunnelRepo.FindByNodeID(node.ID)
	if err != nil {
		logger.Errorf("[Sync] 获取节点 %d 隧道规则失败: %v", node.ID, err)
	} else {
		for _, t := range tunnels {
			// 仅在当前节点是该隧道的入口节点时同步状态
			if t.EntryNodeID == node.ID {
				s.syncTunnelStatus(t, serviceStates, runningChains)
			}
		}
	}
}

// syncForwardStatus 同步转发规则状态
func (s *RuleSyncService) syncForwardStatus(f model.GostForward, serviceStates map[string]string) {
	serviceID := f.ServiceID
	if serviceID == "" {
		serviceID = fmt.Sprintf("forward-%d", f.ID)
	}

	state := serviceStates[serviceID]
	newStatus := utils.GostStateToForwardStatus(state)

	// 如果状态不一致
	if f.Status != newStatus {
		logger.Infof("[Sync] 转发规则 %d (%s) 状态变更: %s -> %s (Gost State: %s)", f.ID, f.Name, f.Status, newStatus, state)
		_ = s.forwardRepo.UpdateStatus(f.ID, newStatus)
	}
}

// syncTunnelStatus 同步隧道规则状态
func (s *RuleSyncService) syncTunnelStatus(t model.GostTunnel, serviceStates map[string]string, runningChains map[string]bool) {
	serviceID := t.ServiceID
	chainID := t.ChainID

	state := serviceStates[serviceID]
	newStatus := utils.GostStateToTunnelStatus(state, runningChains[chainID])

	if t.Status != newStatus {
		logger.Infof("[Sync] 隧道规则 %d (%s) 状态变更: %s -> %s (Gost State: %s)", t.ID, t.Name, t.Status, newStatus, state)
		_ = s.tunnelRepo.UpdateStatus(t.ID, newStatus)
	}
}
