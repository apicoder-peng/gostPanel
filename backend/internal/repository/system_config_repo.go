package repository

import (
	"errors"
	"gost-panel/internal/model"

	"gorm.io/gorm"
)

// SystemConfigRepository 系统配置仓库
type SystemConfigRepository struct {
	db *gorm.DB
}

// NewSystemConfigRepository 创建系统配置仓库
func NewSystemConfigRepository(db *gorm.DB) *SystemConfigRepository {
	return &SystemConfigRepository{db: db}
}

// Get 获取系统配置 (单例)
func (r *SystemConfigRepository) Get() (*model.SystemConfig, error) {
	var config model.SystemConfig
	// 尝试查找 ID 为 1 的记录
	result := r.db.First(&config, 1)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// 如果不存在，创建默认配置
			config = model.SystemConfig{
				ID:               1,
				SiteTitle:        "Gost Panel",
				LogLevel:         "info",
				LogRetentionDays: 7,
			}
			if err := r.db.Create(&config).Error; err != nil {
				return nil, err
			}
			return &config, nil
		}
		return nil, result.Error
	}
	return &config, nil
}

// Update 更新系统配置
func (r *SystemConfigRepository) Update(config *model.SystemConfig) error {
	// 确保 ID 为 1
	config.ID = 1
	return r.db.Save(config).Error
}
