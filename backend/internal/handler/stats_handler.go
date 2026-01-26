package handler

import (
	"gost-panel/internal/service"
	"gost-panel/pkg/response"

	"github.com/gin-gonic/gin"
)

// StatsHandler 统计控制器
type StatsHandler struct {
	statsService *service.StatsService
}

// NewStatsHandler 创建统计控制器
func NewStatsHandler(statsService *service.StatsService) *StatsHandler {
	return &StatsHandler{statsService: statsService}
}

// GetDashboard 获取仪表盘数据
func (h *StatsHandler) GetDashboard(c *gin.Context) {
	stats, err := h.statsService.GetDashboardStats()
	if err != nil {
		response.InternalError(c, "获取统计数据失败")
		return
	}

	response.Success(c, stats)
}
