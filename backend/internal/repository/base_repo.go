package repository

import (
	"gorm.io/gorm"
)

// BaseRepository 基础仓库
type BaseRepository struct {
	DB *gorm.DB
}

// NewBaseRepository 创建基础仓库
func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{DB: db}
}

// Pagination 分页参数
type Pagination struct {
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
	Total    int64 `json:"total"`
}

// GetOffset 获取偏移量
func (p *Pagination) GetOffset() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
	return (p.Page - 1) * p.PageSize
}

// QueryOption 查询选项
type QueryOption struct {
	Preloads   []string       // 预加载关联
	Orders     []string       // 排序
	Conditions map[string]any // 条件过滤
	Pagination *Pagination    // 分页
}

// ApplyConditions 应用查询条件(不包含分页)
func ApplyConditions(db *gorm.DB, opt *QueryOption) *gorm.DB {
	if opt == nil {
		return db
	}

	// 预加载关联
	for _, preload := range opt.Preloads {
		db = db.Preload(preload)
	}

	// 条件过滤
	for key, value := range opt.Conditions {
		// 如果 value 是切片类型,需要展开参数
		if slice, ok := value.([]interface{}); ok {
			db = db.Where(key, slice...)
		} else {
			db = db.Where(key, value)
		}
	}

	// 排序
	for _, order := range opt.Orders {
		db = db.Order(order)
	}

	return db
}

// ApplyPagination 应用分页
func ApplyPagination(db *gorm.DB, opt *QueryOption) *gorm.DB {
	if opt == nil || opt.Pagination == nil {
		return db
	}

	db = db.Offset(opt.Pagination.GetOffset()).Limit(opt.Pagination.PageSize)
	return db
}

// ApplyOptions 应用所有查询选项（为了兼容旧代码）
func ApplyOptions(db *gorm.DB, opt *QueryOption) *gorm.DB {
	db = ApplyConditions(db, opt)
	db = ApplyPagination(db, opt)
	return db
}

// UpdateField 更新单个字段
func (r *BaseRepository) UpdateField(model any, id uint, column string, value any) error {
	return r.DB.Model(model).Where("id = ?", id).Update(column, value).Error
}

// UpdateFields 更新多个字段
func (r *BaseRepository) UpdateFields(model any, id uint, values map[string]any) error {
	return r.DB.Model(model).Where("id = ?", id).Updates(values).Error
}
