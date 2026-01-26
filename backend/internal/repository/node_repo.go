package repository

import (
	"gost-panel/internal/model"
	"time"

	"gorm.io/gorm"
)

// NodeRepository 节点仓库
type NodeRepository struct {
	*BaseRepository
}

// NewNodeRepository 创建节点仓库
func NewNodeRepository(db *gorm.DB) *NodeRepository {
	return &NodeRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建节点
func (r *NodeRepository) Create(node *model.GostNode) error {
	return r.DB.Create(node).Error
}

// Update 更新节点
func (r *NodeRepository) Update(node *model.GostNode) error {
	return r.DB.Save(node).Error
}

// Delete 删除节点
func (r *NodeRepository) Delete(id uint) error {
	return r.DB.Delete(&model.GostNode{}, id).Error
}

// FindByID 根据 ID 查询节点
func (r *NodeRepository) FindByID(id uint) (*model.GostNode, error) {
	var node model.GostNode
	err := r.DB.First(&node, id).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// FindByIDWithRelations 根据 ID 查询节点（包含关联）
func (r *NodeRepository) FindByIDWithRelations(id uint) (*model.GostNode, error) {
	var node model.GostNode
	err := r.DB.Preload("Forwards").Preload("EntryTunnels").Preload("ExitTunnels").First(&node, id).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// List 查询节点列表
func (r *NodeRepository) List(opt *QueryOption) ([]model.GostNode, int64, error) {
	var nodes []model.GostNode
	var total int64

	db := r.DB.Model(&model.GostNode{})

	// 应用条件过滤（统计总数前）
	db = ApplyConditions(db, opt)

	// 统计总数（包含过滤条件）
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 默认按创建时间倒序
	if opt == nil || len(opt.Orders) == 0 {
		db = db.Order("created_at DESC")
	}

	// 应用分页
	db = ApplyPagination(db, opt)

	if err := db.Find(&nodes).Error; err != nil {
		return nil, 0, err
	}

	return nodes, total, nil
}

// FindByName 根据名称查询节点
func (r *NodeRepository) FindByName(name string) (*model.GostNode, error) {
	var node model.GostNode
	err := r.DB.Where("name = ?", name).First(&node).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// ExistsByName 检查名称是否存在
func (r *NodeRepository) ExistsByName(name string, excludeID ...uint) (bool, error) {
	var count int64
	db := r.DB.Model(&model.GostNode{}).Where("name = ?", name)
	if len(excludeID) > 0 {
		db = db.Where("id != ?", excludeID[0])
	}
	err := db.Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateStatus 更新节点状态
func (r *NodeRepository) UpdateStatus(id uint, status model.NodeStatus) error {
	return r.UpdateField(&model.GostNode{}, id, "status", status)
}

// UpdateLastCheck 更新最后检查时间
func (r *NodeRepository) UpdateLastCheck(id uint) error {
	return r.UpdateField(&model.GostNode{}, id, "last_check_at", time.Now())
}

// GetAllOnline 获取所有在线节点
func (r *NodeRepository) GetAllOnline() ([]model.GostNode, error) {
	var nodes []model.GostNode
	err := r.DB.Where("status = ?", model.NodeStatusOnline).Find(&nodes).Error
	return nodes, err
}

// CountByStatus 按状态统计数量
func (r *NodeRepository) CountByStatus(status model.NodeStatus) (int64, error) {
	var count int64
	err := r.DB.Model(&model.GostNode{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

// CountAll 统计总数
func (r *NodeRepository) CountAll() (int64, error) {
	var count int64
	err := r.DB.Model(&model.GostNode{}).Count(&count).Error
	return count, err
}
