package service

import (
	"gost-panel/internal/config"
	"gost-panel/internal/repository"

	"gorm.io/gorm"
)

// StatsService 统计服务
type StatsService struct {
	nodeRepo    *repository.NodeRepository
	forwardRepo *repository.ForwardRepository
	tunnelRepo  *repository.TunnelRepository
	logRepo     *repository.OperationLogRepository
}

// NewStatsService 创建统计服务
func NewStatsService(db *gorm.DB) *StatsService {
	return &StatsService{
		nodeRepo:    repository.NewNodeRepository(db),
		forwardRepo: repository.NewForwardRepository(db),
		tunnelRepo:  repository.NewTunnelRepository(db),
		logRepo:     repository.NewOperationLogRepository(db),
	}
}

// DashboardStats 仪表盘统计
type DashboardStats struct {
	Nodes    NodeStats    `json:"nodes"`
	Forwards ForwardStats `json:"forwards"`
	Tunnels  TunnelStats  `json:"tunnels"`
	Version  string       `json:"version"`
}

// NodeStats 节点统计
type NodeStats struct {
	Total   int64 `json:"total"`
	Online  int64 `json:"online"`
	Offline int64 `json:"offline"`
}

// ForwardStats 转发统计
type ForwardStats struct {
	Total   int64 `json:"total"`
	Running int64 `json:"running"`
	Stopped int64 `json:"stopped"`
}

// TunnelStats 隧道统计
type TunnelStats struct {
	Total int64 `json:"total"`
}

// GetDashboardStats 获取仪表盘统计
func (s *StatsService) GetDashboardStats() (*DashboardStats, error) {
	stats := &DashboardStats{}

	// 节点统计
	nodeTotal, err := s.nodeRepo.CountAll()
	if err != nil {
		return nil, err
	}
	nodeOnline, err := s.nodeRepo.CountByStatus("online")
	if err != nil {
		return nil, err
	}
	stats.Nodes = NodeStats{
		Total:   nodeTotal,
		Online:  nodeOnline,
		Offline: nodeTotal - nodeOnline,
	}

	// 转发统计
	forwardTotal, err := s.forwardRepo.CountAll()
	if err != nil {
		return nil, err
	}
	forwardRunning, err := s.forwardRepo.CountByStatus("running")
	if err != nil {
		return nil, err
	}
	stats.Forwards = ForwardStats{
		Total:   forwardTotal,
		Running: forwardRunning,
		Stopped: forwardTotal - forwardRunning,
	}

	// 隧道统计
	tunnelTotal, err := s.tunnelRepo.CountAll()
	if err != nil {
		return nil, err
	}
	stats.Tunnels = TunnelStats{
		Total: tunnelTotal,
	}

	stats.Version = config.Version

	return stats, nil
}
