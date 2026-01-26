package model

import (
	"time"

	"gorm.io/gorm"
)

// ForwardStatus 转发状态
type ForwardStatus string

const (
	ForwardStatusRunning ForwardStatus = "running" // 运行中
	ForwardStatusStopped ForwardStatus = "stopped" // 已停止
	ForwardStatusError   ForwardStatus = "error"   // 错误
)

// ForwardProtocol 转发协议
type ForwardProtocol string

const (
	ForwardProtocolTCP ForwardProtocol = "tcp" // TCP 协议
	ForwardProtocolUDP ForwardProtocol = "udp" // UDP 协议
)

// GostForward 端口转发模型
type GostForward struct {
	ID         uint            `gorm:"primaryKey" json:"id"`
	NodeID     uint            `gorm:"not null;index" json:"node_id"`                // 节点ID
	Name       string          `gorm:"size:100;not null" json:"name"`                // 规则名称
	Protocol   ForwardProtocol `gorm:"size:10;not null;default:tcp" json:"protocol"` // 协议
	ListenPort int             `gorm:"not null" json:"listen_port"`                  // 监听端口

	Targets   []string      `gorm:"type:json;serializer:json" json:"targets"` // 多目标列表 (host:port)
	Strategy  string        `gorm:"size:20;default:round" json:"strategy"`    // 负载均衡策略 (round, random, fifo)
	EnableTLS bool          `gorm:"default:false" json:"enable_tls"`          // 是否启用 TLS
	Status    ForwardStatus `gorm:"size:20;default:stopped" json:"status"`    // 状态
	ServiceID string        `gorm:"size:100" json:"service_id"`               // Gost 服务 ID

	// 流量监控配置
	ObserverID string `gorm:"size:100" json:"observer_id"` // 观察器 ID

	// 流量统计 (由观察器更新)
	InputBytes    int64 `gorm:"default:0" json:"input_bytes"`    // 入站总流量 (bytes)
	OutputBytes   int64 `gorm:"default:0" json:"output_bytes"`   // 出站总流量 (bytes)
	TotalBytes    int64 `gorm:"default:0" json:"total_bytes"`    // 总流量 (Input + Output)
	TotalRequests int64 `gorm:"default:0" json:"total_requests"` // 总请求数

	Remark    string         `gorm:"type:text" json:"remark"` // 备注
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	Node *GostNode `gorm:"foreignKey:NodeID" json:"node,omitempty"`
}

// TableName 指定表名
func (GostForward) TableName() string {
	return "gost_forwards"
}
