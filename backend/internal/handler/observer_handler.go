package handler

import (
	"gost-panel/internal/dto"
	"gost-panel/internal/service"
	"gost-panel/pkg/logger"

	"github.com/gin-gonic/gin"
)

// ObserverHandler 观察器控制器
type ObserverHandler struct {
	observerService *service.ObserverService
}

// NewObserverHandler 创建观察器控制器
func NewObserverHandler(observerService *service.ObserverService) *ObserverHandler {
	return &ObserverHandler{observerService: observerService}
}

// Report 接收 GOST 观察器上报的数据
// POST /api/v1/observer/report
func (h *ObserverHandler) Report(c *gin.Context) {
	var req dto.ObserverReportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("解析观察器上报数据失败: %v", err)
		c.JSON(400, dto.ObserverReportResp{OK: false})
		return
	}

	// 处理上报数据
	if err := h.observerService.HandleReport(&req); err != nil {
		logger.Warnf("处理观察器上报数据失败: %v", err)
		c.JSON(500, dto.ObserverReportResp{OK: false})
		return
	}

	// 返回成功响应（GOST 需要 ok: true 才认为上报成功）
	c.JSON(200, dto.ObserverReportResp{OK: true})
}
