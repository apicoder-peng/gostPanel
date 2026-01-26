package utils

import (
	"time"

	"gost-panel/internal/model"
	"gost-panel/pkg/gost"
)

// GetGostClient 根据节点配置创建 Gost 客户端
func GetGostClient(node *model.GostNode) *gost.Client {
	return gost.NewClient(&gost.Config{
		APIURL:   node.APIURL,
		Username: node.Username,
		Password: node.Password,
		Timeout:  5 * time.Second, // 统一设置超时时间
	})
}
