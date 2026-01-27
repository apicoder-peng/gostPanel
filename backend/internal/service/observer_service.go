package service

import (
	"gost-panel/internal/dto"
	"gost-panel/internal/errors"
	"gost-panel/internal/model"
	"gost-panel/internal/repository"
	"gost-panel/pkg/gost"
	"gost-panel/pkg/logger"
	"strings"

	"gorm.io/gorm"
)

// ObserverService 观察器服务
type ObserverService struct {
	ruleRepo *repository.RuleRepository
	nodeRepo *repository.NodeRepository
}

// NewObserverService 创建观察器服务
func NewObserverService(db *gorm.DB) *ObserverService {
	return &ObserverService{
		ruleRepo: repository.NewRuleRepository(db),
		nodeRepo: repository.NewNodeRepository(db),
	}
}

// HandleReport 处理观察器上报的数据
func (s *ObserverService) HandleReport(req *dto.ObserverReportReq) error {
	for _, event := range req.Events {
		if err := s.processEvent(&event); err != nil {
			logger.Warnf("处理观察器事件失败: %v", err)
		}
	}
	return nil
}

// processEvent 处理单个事件
func (s *ObserverService) processEvent(event *dto.ObserverEvent) error {
	// 只处理统计类型的事件
	if event.Type != "stats" || event.Stats == nil {
		return nil
	}

	serviceName := event.Service
	if serviceName == "" {
		return nil
	}

	// 解析服务名称，格式: rule-{id} 或 forward-{id} 或 tunnel-{id}
	if strings.HasPrefix(serviceName, "rule-") {
		return s.updateRuleStats(serviceName, event.Stats, "rule-")
	} else if strings.HasPrefix(serviceName, "forward-") {
		// 保持向后兼容
		return s.updateRuleStats(serviceName, event.Stats, "forward-")
	} else if strings.HasPrefix(serviceName, "tunnel-") {
		// 保持向后兼容
		return s.updateRuleStats(serviceName, event.Stats, "tunnel-")
	}

	return nil
}

// updateRuleStats 更新规则统计
func (s *ObserverService) updateRuleStats(serviceName string, stats *dto.ObserverStats, prefix string) error {
	// 解析 ID
	var id uint
	if _, err := parseServiceID(serviceName, prefix, &id); err != nil {
		return err
	}

	// 更新规则统计数据
	if err := s.ruleRepo.UpdateStats(id, stats.InputBytes, stats.OutputBytes, stats.TotalConns); err != nil {
		return err
	}

	// 同步更新节点统计
	// 1. 查询规则获取关联节点
	rule, err := s.ruleRepo.FindByID(id)
	if err != nil {
		// 如果找不到规则，可能已被删除，忽略错误
		return nil
	}

	// 2. 确定节点 ID
	var nodeID uint
	if rule.Type == model.RuleTypeTunnel && rule.Tunnel != nil {
		nodeID = rule.Tunnel.EntryNodeID
	} else if rule.NodeID != nil {
		nodeID = *rule.NodeID
	}

	// 3. 更新节点流量
	if nodeID > 0 {
		if err := s.nodeRepo.UpdateStats(nodeID, stats.InputBytes, stats.OutputBytes); err != nil {
			logger.Warnf("更新节点流量失败: %v", err)
		}
	}

	logger.Debugf("更新规则统计: %s%d, In: %d, Out: %d, Req: %d",
		prefix, id, stats.InputBytes, stats.OutputBytes, stats.TotalConns)
	return nil
}

// parseServiceID 从服务名称解析 ID
func parseServiceID(serviceName, prefix string, id *uint) (bool, error) {
	if !strings.HasPrefix(serviceName, prefix) {
		return false, nil
	}

	idStr := strings.TrimPrefix(serviceName, prefix)
	var parsedID uint
	if _, err := parseUint(idStr, &parsedID); err != nil {
		return false, err
	}

	*id = parsedID
	return true, nil
}

// parseUint 解析无符号整数
func parseUint(s string, result *uint) (bool, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return false, nil
		}
		n = n*10 + int(c-'0')
	}
	*result = uint(n)
	return true, nil
}

// EnsureGlobalObserver 确保全局流量监控观察器存在
// 返回 observerName (如果成功) 或 error
func EnsureGlobalObserver(client *gost.Client, sysRepo *repository.SystemConfigRepository) (string, error) {
	// 获取系统配置中的面板地址
	sysConfig, err := sysRepo.Get()
	if err != nil || sysConfig.PanelURL == "" {
		return "", errors.ErrPanelURLNotFound
	}

	// 使用固定名称，确保每个节点只有一个观察器
	observerName := "observer-global"
	observer := &gost.ObserverConfig{
		Name: observerName,
		Plugin: &gost.PluginConfig{
			Type:    "http",
			Addr:    sysConfig.PanelURL + "/api/v1/observer/report",
			Timeout: "10s",
		},
	}

	if err = client.CreateObserver(observer); err != nil {
		logger.Warnf("创建/更新观察器失败: %v", err)
		return "", errors.ErrObserverCreateFailed
	}

	logger.Infof("确保观察器存在: %s (URL: %s)", observerName, sysConfig.PanelURL)
	return observerName, nil
}
