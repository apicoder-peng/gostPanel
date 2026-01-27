package handler

import (
	"gost-panel/internal/dto"
	"gost-panel/internal/service"
	"gost-panel/pkg/response"

	"github.com/gin-gonic/gin"
)

// SystemConfigHandler 系统配置控制器
type SystemConfigHandler struct {
	systemConfigService *service.SystemConfigService
	backupService       *service.BackupService
}

// NewSystemConfigHandler 创建系统配置控制器
func NewSystemConfigHandler(sysService *service.SystemConfigService, backupService *service.BackupService) *SystemConfigHandler {
	return &SystemConfigHandler{
		systemConfigService: sysService,
		backupService:       backupService,
	}
}

// GetConfig 获取配置
func (h *SystemConfigHandler) GetConfig(c *gin.Context) {
	config, err := h.systemConfigService.GetConfig()
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, config)
}

// GetPublicConfig 获取公开配置
func (h *SystemConfigHandler) GetPublicConfig(c *gin.Context) {
	config, err := h.systemConfigService.GetPublicConfig()
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, config)
}

// UpdateConfig 更新配置
func (h *SystemConfigHandler) UpdateConfig(c *gin.Context) {
	var req dto.UpdateSystemConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if err := h.systemConfigService.UpdateConfig(&req); err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessWithMessage(c, "配置更新成功", nil)
}

// TestEmail 发送测试邮件
func (h *SystemConfigHandler) TestEmail(c *gin.Context) {
	var req dto.EmailConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if err := h.systemConfigService.SendTestEmail(&req); err != nil {
		response.Error(c, 500, 50001, "发送失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "邮件发送成功", nil)
}

// Backup 立即备份
func (h *SystemConfigHandler) Backup(c *gin.Context) {
	if err := h.backupService.CreateBackup(); err != nil {
		response.Error(c, 500, 50002, "备份失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "备份成功", nil)
}
