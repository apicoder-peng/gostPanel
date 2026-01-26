package repository

import (
	"gost-panel/internal/model"

	"gorm.io/gorm"
)

// OperationLogRepository 操作日志仓库
type OperationLogRepository struct {
	*BaseRepository
}

// NewOperationLogRepository 创建操作日志仓库
func NewOperationLogRepository(db *gorm.DB) *OperationLogRepository {
	return &OperationLogRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建操作日志
func (r *OperationLogRepository) Create(log *model.OperationLog) error {
	return r.DB.Create(log).Error
}

// List 查询操作日志列表
func (r *OperationLogRepository) List(opt *QueryOption) ([]model.OperationLog, int64, error) {
	var logs []model.OperationLog
	var total int64

	db := r.DB.Model(&model.OperationLog{})

	// 应用条件过滤
	db = ApplyConditions(db, opt)

	// 统计总数（包含过滤条件）
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 默认按时间倒序
	db = db.Order("created_at DESC")

	// 应用分页
	db = ApplyPagination(db, opt)

	if err := db.Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// FindByUserID 根据用户 ID 查询操作日志
func (r *OperationLogRepository) FindByUserID(userID uint, opt *QueryOption) ([]model.OperationLog, int64, error) {
	if opt == nil {
		opt = &QueryOption{}
	}
	if opt.Conditions == nil {
		opt.Conditions = make(map[string]any)
	}
	opt.Conditions["user_id = ?"] = userID

	return r.List(opt)
}
