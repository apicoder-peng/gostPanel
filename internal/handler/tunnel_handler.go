package handler

import (
	"strconv"

	"gost-panel/internal/dto"
	"gost-panel/internal/service"
	"gost-panel/pkg/response"

	"github.com/gin-gonic/gin"
)

// TunnelHandler 隧道控制器
// 处理多跳隧道相关的 HTTP 请求
type TunnelHandler struct {
	tunnelService *service.TunnelService
}

// NewTunnelHandler 创建隧道控制器
func NewTunnelHandler(tunnelService *service.TunnelService) *TunnelHandler {
	return &TunnelHandler{tunnelService: tunnelService}
}

// Create 创建隧道
func (h *TunnelHandler) Create(c *gin.Context) {
	var req dto.CreateTunnelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, _ := c.Get("userID")
	username, _ := c.Get("username")

	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	tunnel, err := h.tunnelService.Create(&req, userID.(uint), username.(string), ip, ua)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, tunnel)
}

// Update 更新隧道
func (h *TunnelHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的隧道 ID")
		return
	}

	var req dto.UpdateTunnelReq
	if err = c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, _ := c.Get("userID")
	username, _ := c.Get("username")

	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	tunnel, err := h.tunnelService.Update(uint(id), &req, userID.(uint), username.(string), ip, ua)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, tunnel)
}

// Delete 删除隧道
func (h *TunnelHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的隧道 ID")
		return
	}

	userID, _ := c.Get("userID")
	username, _ := c.Get("username")

	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	if err := h.tunnelService.Delete(uint(id), userID.(uint), username.(string), ip, ua); err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// GetByID 获取隧道详情
func (h *TunnelHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的隧道 ID")
		return
	}

	tunnel, err := h.tunnelService.GetByID(uint(id))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, tunnel)
}

// List 获取隧道列表
func (h *TunnelHandler) List(c *gin.Context) {
	var req dto.TunnelListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	tunnels, total, err := h.tunnelService.List(&req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessPage(c, tunnels, total, req.Page, req.PageSize)
}

// Start 启动隧道
func (h *TunnelHandler) Start(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的隧道 ID")
		return
	}

	userID, _ := c.Get("userID")
	username, _ := c.Get("username")

	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	if err := h.tunnelService.Start(uint(id), userID.(uint), username.(string), ip, ua); err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessWithMessage(c, "启动成功", nil)
}

// Stop 停止隧道
func (h *TunnelHandler) Stop(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的隧道 ID")
		return
	}

	userID, _ := c.Get("userID")
	username, _ := c.Get("username")

	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	if err := h.tunnelService.Stop(uint(id), userID.(uint), username.(string), ip, ua); err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessWithMessage(c, "停止成功", nil)
}
