package model

import (
	"time"
)

// OperationLog 操作日志模型
type OperationLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"index" json:"user_id"`           // 用户ID
	Username     string    `gorm:"size:50" json:"username"`        // 用户名（冗余存储）
	Action       string    `gorm:"size:50;not null" json:"action"` // 操作类型
	ResourceType string    `gorm:"size:50" json:"resource_type"`   // 资源类型
	ResourceID   uint      `json:"resource_id"`                    // 资源ID
	Details      string    `gorm:"type:text" json:"details"`       // 详细信息 (JSON)
	IPAddress    string    `gorm:"size:50" json:"ip_address"`      // IP 地址
	UserAgent    string    `gorm:"size:255" json:"user_agent"`     // User-Agent
	CreatedAt    time.Time `gorm:"index" json:"created_at"`

	// 关联
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (OperationLog) TableName() string {
	return "operation_logs"
}

// 操作类型常量
const (
	ActionLogin          = "login"           // 登录
	ActionLogout         = "logout"          // 登出
	ActionChangePassword = "change_password" // 修改密码
	ActionCreate         = "create"          // 创建
	ActionUpdate         = "update"          // 更新
	ActionDelete         = "delete"          // 删除
	ActionStart          = "start"           // 启动
	ActionStop           = "stop"            // 停止
)

// 资源类型常量
const (
	ResourceTypeNode    = "node"    // 节点
	ResourceTypeForward = "forward" // 转发规则
	ResourceTypeTunnel  = "tunnel"  // 隧道
)
