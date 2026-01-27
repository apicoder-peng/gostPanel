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
// 定时从 Gost 节点同步规则的真实运行状态
type RuleSyncService struct {
	nodeRepo   *repository.NodeRepository
	ruleRepo   *repository.RuleRepository
	tunnelRepo *repository.TunnelRepository
	ticker     *time.Ticker
	stopChan   chan struct{}
	wg         sync.WaitGroup
}

// NewRuleSyncService 创建规则状态同步服务
func NewRuleSyncService(db *gorm.DB) *RuleSyncService {
	return &RuleSyncService{
		nodeRepo:   repository.NewNodeRepository(db),
		ruleRepo:   repository.NewRuleRepository(db),
		tunnelRepo: repository.NewTunnelRepository(db),
		stopChan:   make(chan struct{}),
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
	// 如果节点离线，跳过规则同步
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

	// 1. 同步规则状态
	rules, err := s.ruleRepo.FindByNodeID(node.ID)
	if err != nil {
		logger.Errorf("[Sync] 获取节点 %d 规则失败: %v", node.ID, err)
	} else {
		for _, r := range rules {
			s.syncRuleStatus(r, serviceStates)
		}
	}

	// 2. 同步隧道状态
	// 仅在入口节点检查隧道 Chain 是否存在
	// 理由：入口节点的 Forward Chain 是隧道存在的核心标志。 exits 节点的 Relay 服务只是依赖。
	chainStates := make(map[string]bool)
	for _, chain := range gostCfg.Chains {
		chainStates[chain.Name] = true
	}

	tunnels, err := s.tunnelRepo.FindByNodeID(node.ID)
	if err != nil {
		logger.Errorf("[Sync] 获取节点 %d 隧道规则失败: %v", node.ID, err)
	} else {
		for _, t := range tunnels {
			// 仅对入口节点同步隧道状态
			if t.EntryNodeID == node.ID {
				s.syncTunnelStatus(t, chainStates)
			}
		}
	}
}

// syncRuleStatus 同步规则状态
func (s *RuleSyncService) syncRuleStatus(r model.GostRule, serviceStates map[string]string) {
	serviceID := r.ServiceID
	if serviceID == "" {
		serviceID = fmt.Sprintf("rule-%d", r.ID)
	}

	state := serviceStates[serviceID]
	newStatus := utils.GostStateToRuleStatus(state)

	// 如果状态不一致
	if r.Status != newStatus {
		logger.Infof("[Sync] 规则 %d (%s) 状态变更: %s -> %s (Gost State: %s)", r.ID, r.Name, r.Status, newStatus, state)
		_ = s.ruleRepo.UpdateStatus(r.ID, newStatus)
	}
}

// syncTunnelStatus 同步隧道状态
func (s *RuleSyncService) syncTunnelStatus(t model.GostTunnel, chainStates map[string]bool) {
	// 检查 Forward Chain 是否存在
	chainID := t.ChainID
	if chainID == "" {
		chainID = fmt.Sprintf("tunnel-%d-chain", t.ID)
	}

	exists := chainStates[chainID]
	var newStatus model.TunnelStatus
	if exists {
		newStatus = model.TunnelStatusRunning
	} else {
		newStatus = model.TunnelStatusStopped
	}

	if t.Status != newStatus {
		logger.Infof("[Sync] 隧道 %d (%s) 状态变更: %s -> %s (Chain Exists: %v)", t.ID, t.Name, t.Status, newStatus, exists)
		_ = s.tunnelRepo.UpdateStatus(t.ID, newStatus)
	}
}
