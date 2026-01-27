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

// NodeService 节点服务
// 负责节点的 CRUD 操作和业务逻辑处理
type NodeService struct {
	nodeRepo   *repository.NodeRepository
	logService *LogService
}

// NewNodeService 创建节点服务
func NewNodeService(db *gorm.DB) *NodeService {
	return &NodeService{
		nodeRepo:   repository.NewNodeRepository(db),
		logService: NewLogService(db),
	}
}

// Create 创建节点
// 创建前会检查节点名称是否已存在
func (s *NodeService) Create(req *dto.CreateNodeReq, userID uint, username string, ip, userAgent string) (*model.GostNode, error) {
	// 检查名称是否存在
	exists, err := s.nodeRepo.ExistsByName(req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.ErrNodeNameExists
	}

	// 创建节点
	node := &model.GostNode{
		Name:     req.Name,
		Address:  req.Address,
		Port:     req.Port,
		Username: req.Username,
		Password: req.Password,
		Remark:   req.Remark,
		Status:   model.NodeStatusOffline,
	}

	if err = s.nodeRepo.Create(node); err != nil {
		return nil, err
	}

	// 记录操作日志
	s.logService.Record(
		userID,
		username,
		model.ActionCreate,
		model.ResourceTypeNode,
		node.ID,
		fmt.Sprintf("创建节点: %s", node.Name),
		ip,
		userAgent)

	logger.Infof("创建节点成功: %s (%s:%d)", node.Name, node.Address, node.Port)
	return node, nil
}

// Update 更新节点
// 更新前会检查节点是否存在以及新名称是否与其他节点冲突
func (s *NodeService) Update(id uint, req *dto.UpdateNodeReq, userID uint, username string, ip, userAgent string) (*model.GostNode, error) {
	// 查询节点
	node, err := s.nodeRepo.FindByID(id)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrNodeNotFound
		}
		return nil, err
	}

	// 检查名称是否存在（排除自身）
	exists, err := s.nodeRepo.ExistsByName(req.Name, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.ErrNodeNameExists
	}

	// 更新节点
	node.Name = req.Name
	node.Address = req.Address
	node.Port = req.Port
	node.Username = req.Username
	node.Password = req.Password
	node.Remark = req.Remark

	if err = s.nodeRepo.Update(node); err != nil {
		return nil, err
	}

	// 记录操作日志
	s.logService.Record(
		userID,
		username,
		model.ActionUpdate,
		model.ResourceTypeNode,
		node.ID,
		fmt.Sprintf("更新节点: %s", node.Name),
		ip,
		userAgent)

	return node, nil
}

// Delete 删除节点
// 删除前会检查节点是否存在以及是否有关联的转发规则
func (s *NodeService) Delete(id uint, userID uint, username string, ip, userAgent string) error {
	// 查询节点（包含关联）
	node, err := s.nodeRepo.FindByIDWithRelations(id)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrNodeNotFound
		}
		return err
	}

	// 检查是否有关联的规则
	if len(node.Rules) > 0 {
		return errors.ErrNodeHasRules
	}

	// 删除节点前，用户需要手动删除相关隧道
	if len(node.EntryTunnels) > 0 || len(node.ExitTunnels) > 0 {
		return errors.ErrNodeHasTunnels
	}

	// 删除节点
	if err = s.nodeRepo.Delete(id); err != nil {
		return err
	}

	// 记录操作日志
	s.logService.Record(
		userID,
		username,
		model.ActionDelete,
		model.ResourceTypeNode,
		id,
		fmt.Sprintf("删除节点: %s", node.Name),
		ip,
		userAgent)

	logger.Infof("删除节点成功: %s", node.Name)
	return nil
}

// GetByID 获取节点详情
func (s *NodeService) GetByID(id uint) (*model.GostNode, error) {
	node, err := s.nodeRepo.FindByID(id)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrNodeNotFound
		}
		return nil, err
	}
	return node, nil
}

// List 获取节点列表
func (s *NodeService) List(req *dto.NodeListReq) ([]model.GostNode, int64, error) {
	// 设置默认值
	req.SetDefaults()

	opt := &repository.QueryOption{
		Pagination: &repository.Pagination{
			Page:     req.Page,
			PageSize: req.PageSize,
		},
		Conditions: make(map[string]any),
	}

	// 状态筛选
	if req.Status != "" {
		opt.Conditions["status = ?"] = req.Status
	}

	// 关键词搜索
	if req.Keyword != "" {
		opt.Conditions["name LIKE ? OR address LIKE ?"] = []interface{}{
			"%" + req.Keyword + "%",
			"%" + req.Keyword + "%",
		}
	}

	return s.nodeRepo.List(opt)
}

// GetStats 获取节点统计
func (s *NodeService) GetStats() (map[string]int64, error) {
	total, err := s.nodeRepo.CountAll()
	if err != nil {
		return nil, err
	}

	online, err := s.nodeRepo.CountByStatus(model.NodeStatusOnline)
	if err != nil {
		return nil, err
	}

	return map[string]int64{
		"total":   total,
		"online":  online,
		"offline": total - online,
	}, nil
}

// CreateGostClient 创建节点的 Gost 客户端
func (s *NodeService) CreateGostClient(id uint) (*gost.Client, error) {
	node, err := s.nodeRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	return utils.GetGostClient(node), nil
}

// GetConfig 获取节点配置
func (s *NodeService) GetConfig(id uint) (*gost.GostConfig, error) {
	node, err := s.nodeRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	client := utils.GetGostClient(node)

	config, err := client.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("获取节点配置失败: %v", err)
	}

	return config, nil
}
