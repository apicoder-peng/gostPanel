package handler

import (
	"gost-panel/internal/dto"
	"gost-panel/internal/service"
	"gost-panel/pkg/response"

	"github.com/gin-gonic/gin"
)

// LogHandler 日志控制器
// 处理操作日志相关的 HTTP 请求
type LogHandler struct {
	logService *service.LogService
}

// NewLogHandler 创建日志控制器
func NewLogHandler(logService *service.LogService) *LogHandler {
	return &LogHandler{logService: logService}
}

// List 获取日志列表
func (h *LogHandler) List(c *gin.Context) {
	var req dto.LogListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	logs, total, err := h.logService.List(&req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessPage(c, logs, total, req.Page, req.PageSize)
}
