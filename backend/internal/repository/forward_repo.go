package repository

import (
	"gost-panel/internal/model"

	"gorm.io/gorm"
)

// ForwardRepository 转发规则仓库
type ForwardRepository struct {
	*BaseRepository
}

// NewForwardRepository 创建转发规则仓库
func NewForwardRepository(db *gorm.DB) *ForwardRepository {
	return &ForwardRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建转发规则
func (r *ForwardRepository) Create(forward *model.GostForward) error {
	return r.DB.Create(forward).Error
}

// Update 更新转发规则
func (r *ForwardRepository) Update(forward *model.GostForward) error {
	return r.DB.Save(forward).Error
}

// Delete 删除转发规则
func (r *ForwardRepository) Delete(id uint) error {
	return r.DB.Delete(&model.GostForward{}, id).Error
}

// FindByID 根据 ID 查询转发规则
func (r *ForwardRepository) FindByID(id uint) (*model.GostForward, error) {
	var forward model.GostForward
	err := r.DB.Preload("Node").First(&forward, id).Error
	if err != nil {
		return nil, err
	}
	return &forward, nil
}

// List 查询转发规则列表
func (r *ForwardRepository) List(opt *QueryOption) ([]model.GostForward, int64, error) {
	var forwards []model.GostForward
	var total int64

	db := r.DB.Model(&model.GostForward{})

	// 应用条件过滤
	db = ApplyConditions(db, opt)

	// 统计总数（包含过滤条件）
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 预加载节点
	db = db.Preload("Node")

	// 默认按创建时间倒序
	if opt == nil || len(opt.Orders) == 0 {
		db = db.Order("created_at DESC")
	}

	// 应用分页
	db = ApplyPagination(db, opt)

	if err := db.Find(&forwards).Error; err != nil {
		return nil, 0, err
	}

	return forwards, total, nil
}

// FindByNodeID 根据节点 ID 查询转发规则
func (r *ForwardRepository) FindByNodeID(nodeID uint) ([]model.GostForward, error) {
	var forwards []model.GostForward
	err := r.DB.Where("node_id = ?", nodeID).Find(&forwards).Error
	return forwards, err
}

// ExistsByPort 检查端口是否已被使用
func (r *ForwardRepository) ExistsByPort(nodeID uint, port int, excludeID ...uint) (bool, error) {
	var count int64
	db := r.DB.Model(&model.GostForward{}).
		Where("node_id = ? AND listen_port = ?", nodeID, port)
	if len(excludeID) > 0 {
		db = db.Where("id != ?", excludeID[0])
	}
	err := db.Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateStatus 更新转发状态
func (r *ForwardRepository) UpdateStatus(id uint, status model.ForwardStatus) error {
	return r.UpdateField(&model.GostForward{}, id, "status", status)
}

// UpdateServiceID 更新服务 ID
func (r *ForwardRepository) UpdateServiceID(id uint, serviceID string) error {
	return r.UpdateField(&model.GostForward{}, id, "service_id", serviceID)
}

// UpdatePing 更新 Ping 值
func (r *ForwardRepository) UpdatePing(id uint, ping int64) error {
	return r.DB.Model(&model.GostForward{}).Where("id = ?", id).
		Update("ping", ping).Error
}

// UpdateObserverID 更新观察器 ID
func (r *ForwardRepository) UpdateObserverID(id uint, observerID string) error {
	return r.UpdateField(&model.GostForward{}, id, "observer_id", observerID)
}

// CountByNodeID 按节点统计数量
func (r *ForwardRepository) CountByNodeID(nodeID uint) (int64, error) {
	var count int64
	err := r.DB.Model(&model.GostForward{}).Where("node_id = ?", nodeID).Count(&count).Error
	return count, err
}

// CountAll 统计总数
func (r *ForwardRepository) CountAll() (int64, error) {
	var count int64
	err := r.DB.Model(&model.GostForward{}).Count(&count).Error
	return count, err
}

// CountByStatus 按状态统计
func (r *ForwardRepository) CountByStatus(status model.ForwardStatus) (int64, error) {
	var count int64
	err := r.DB.Model(&model.GostForward{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

// StopByNodeID 停止指定节点的所有转发
func (r *ForwardRepository) StopByNodeID(nodeID uint) error {
	return r.DB.Model(&model.GostForward{}).
		Where("node_id = ? AND status = ?", nodeID, model.ForwardStatusRunning).
		Update("status", model.ForwardStatusStopped).Error
}

// UpdateStats 更新流量统计
func (r *ForwardRepository) UpdateStats(id uint, inputBytes, outputBytes, totalRequests int64) error {
	return r.UpdateFields(&model.GostForward{}, id, map[string]interface{}{
		"input_bytes":    inputBytes,
		"output_bytes":   outputBytes,
		"total_bytes":    inputBytes + outputBytes,
		"total_requests": totalRequests,
	})
}
