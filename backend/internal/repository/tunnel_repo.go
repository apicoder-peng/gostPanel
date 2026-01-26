package repository

import (
	"gost-panel/internal/model"

	"gorm.io/gorm"
)

// TunnelRepository 隧道仓库
type TunnelRepository struct {
	*BaseRepository
}

// NewTunnelRepository 创建隧道仓库
func NewTunnelRepository(db *gorm.DB) *TunnelRepository {
	return &TunnelRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建隧道
func (r *TunnelRepository) Create(tunnel *model.GostTunnel) error {
	return r.DB.Create(tunnel).Error
}

// Update 更新隧道
func (r *TunnelRepository) Update(tunnel *model.GostTunnel) error {
	return r.DB.Save(tunnel).Error
}

// Delete 删除隧道
func (r *TunnelRepository) Delete(id uint) error {
	return r.DB.Delete(&model.GostTunnel{}, id).Error
}

// FindByID 根据 ID 查询隧道（包含关联节点）
func (r *TunnelRepository) FindByID(id uint) (*model.GostTunnel, error) {
	var tunnel model.GostTunnel
	err := r.DB.Preload("EntryNode").Preload("ExitNode").First(&tunnel, id).Error
	if err != nil {
		return nil, err
	}
	return &tunnel, nil
}

// List 查询隧道列表
func (r *TunnelRepository) List(opt *QueryOption) ([]model.GostTunnel, int64, error) {
	var tunnels []model.GostTunnel
	var total int64

	db := r.DB.Model(&model.GostTunnel{})

	// 应用条件过滤
	db = ApplyConditions(db, opt)

	// 统计总数（包含过滤条件）
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 预加载节点
	db = db.Preload("EntryNode").Preload("ExitNode")

	// 默认按创建时间倒序
	if opt == nil || len(opt.Orders) == 0 {
		db = db.Order("created_at DESC")
	}

	// 应用分页
	db = ApplyPagination(db, opt)

	if err := db.Find(&tunnels).Error; err != nil {
		return nil, 0, err
	}

	return tunnels, total, nil
}

// ExistsByPort 检查入口节点端口是否已被使用
func (r *TunnelRepository) ExistsByPort(entryNodeID uint, port int, excludeID ...uint) (bool, error) {
	var count int64
	db := r.DB.Model(&model.GostTunnel{}).
		Where("entry_node_id = ? AND listen_port = ?", entryNodeID, port)
	if len(excludeID) > 0 {
		db = db.Where("id != ?", excludeID[0])
	}
	err := db.Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateStatus 更新隧道状态
func (r *TunnelRepository) UpdateStatus(id uint, status model.TunnelStatus) error {
	return r.UpdateField(&model.GostTunnel{}, id, "status", status)
}

// UpdateServiceInfo 更新服务信息
func (r *TunnelRepository) UpdateServiceInfo(id uint, serviceID, chainID string) error {
	return r.UpdateFields(&model.GostTunnel{}, id, map[string]interface{}{
		"service_id": serviceID,
		"chain_id":   chainID,
	})
}

// UpdatePing 更新 Ping 值
func (r *TunnelRepository) UpdatePing(id uint, entryPing, exitPing int64) error {
	return r.DB.Model(&model.GostTunnel{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"entry_ping": entryPing,
			"exit_ping":  exitPing,
		}).Error
}

// UpdateObserverID 更新观察器 ID
func (r *TunnelRepository) UpdateObserverID(id uint, observerID string) error {
	return r.UpdateField(&model.GostTunnel{}, id, "observer_id", observerID)
}

// CountAll 统计总数
func (r *TunnelRepository) CountAll() (int64, error) {
	var count int64
	err := r.DB.Model(&model.GostTunnel{}).Count(&count).Error
	return count, err
}

// FindByNodeID 查找节点相关的隧道
func (r *TunnelRepository) FindByNodeID(nodeID uint) ([]model.GostTunnel, error) {
	var tunnels []model.GostTunnel
	err := r.DB.Where("entry_node_id = ? OR exit_node_id = ?", nodeID, nodeID).Find(&tunnels).Error
	return tunnels, err
}

// StopByNodeID 停止与该节点相关的所有隧道
func (r *TunnelRepository) StopByNodeID(nodeID uint) error {
	return r.DB.Model(&model.GostTunnel{}).
		Where("(entry_node_id = ? OR exit_node_id = ?) AND status = ?", nodeID, nodeID, model.TunnelStatusRunning).
		Update("status", model.TunnelStatusStopped).Error
}

// UpdateStats 更新流量统计
func (r *TunnelRepository) UpdateStats(id uint, inputBytes, outputBytes, totalRequests int64) error {
	return r.UpdateFields(&model.GostTunnel{}, id, map[string]interface{}{
		"input_bytes":    inputBytes,
		"output_bytes":   outputBytes,
		"total_bytes":    inputBytes + outputBytes,
		"total_requests": totalRequests,
	})
}
