package model

import (
	"time"

	"gorm.io/gorm"
)

// SystemConfig 系统配置模型
type SystemConfig struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 面板地址
	PanelURL string `gorm:"size:255" json:"panel_url"`

	// 邮箱配置
	SMTPHost     string `gorm:"size:255" json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUsername string `gorm:"size:100" json:"smtp_username"`
	SMTPPassword string `gorm:"size:255" json:"smtp_password"`
	SMTPFrom     string `gorm:"size:255" json:"smtp_from"` // 发件人邮箱

	// 日志策略
	LogRetentionDays int    `gorm:"default:7" json:"log_retention_days"`
	LogLevel         string `gorm:"size:20;default:info" json:"log_level"`

	// 备份与恢复
	AutoBackup           bool `gorm:"default:false" json:"auto_backup"`
	BackupRetentionCount int  `gorm:"default:7" json:"backup_retention_count"`

	// 面板配置
	SiteTitle string `gorm:"size:100;default:Gost Panel" json:"site_title"`
	LogoURL   string `gorm:"size:255" json:"logo_url"`
	Copyright string `gorm:"size:255" json:"copyright"`
}

// TableName 指定表名
func (SystemConfig) TableName() string {
	return "system_configs"
}
