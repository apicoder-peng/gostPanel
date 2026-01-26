// Package handler 提供 HTTP 请求处理器
package handler

import (
	"strconv"

	"gost-panel/internal/dto"
	"gost-panel/internal/service"
	"gost-panel/pkg/response"

	"github.com/gin-gonic/gin"
)

// ForwardHandler 转发控制器
// 处理端口转发相关的 HTTP 请求
type ForwardHandler struct {
	forwardService *service.ForwardService
}

// NewForwardHandler 创建转发控制器
func NewForwardHandler(forwardService *service.ForwardService) *ForwardHandler {
	return &ForwardHandler{forwardService: forwardService}
}

// Create 创建转发规则
func (h *ForwardHandler) Create(c *gin.Context) {
	var req dto.CreateForwardReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, _ := c.Get("userID")
	username, _ := c.Get("username")

	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	forward, err := h.forwardService.Create(&req, userID.(uint), username.(string), ip, ua)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, forward)
}

// Update 更新转发规则
func (h *ForwardHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的转发 ID")
		return
	}

	var req dto.UpdateForwardReq
	if err = c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, _ := c.Get("userID")
	username, _ := c.Get("username")

	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	forward, err := h.forwardService.Update(uint(id), &req, userID.(uint), username.(string), ip, ua)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, forward)
}

// Delete 删除转发规则
func (h *ForwardHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的转发 ID")
		return
	}

	userID, _ := c.Get("userID")
	username, _ := c.Get("username")

	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	if err = h.forwardService.Delete(uint(id), userID.(uint), username.(string), ip, ua); err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// GetByID 获取转发规则详情
func (h *ForwardHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的转发 ID")
		return
	}

	forward, err := h.forwardService.GetByID(uint(id))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, forward)
}

// List 获取转发规则列表
func (h *ForwardHandler) List(c *gin.Context) {
	var req dto.ForwardListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	forwards, total, err := h.forwardService.List(&req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessPage(c, forwards, total, req.Page, req.PageSize)
}

// Start 启动转发
func (h *ForwardHandler) Start(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的转发 ID")
		return
	}

	userID, _ := c.Get("userID")
	username, _ := c.Get("username")

	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	if err = h.forwardService.Start(uint(id), userID.(uint), username.(string), ip, ua); err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessWithMessage(c, "启动成功", nil)
}

// Stop 停止转发
func (h *ForwardHandler) Stop(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的转发 ID")
		return
	}

	userID, _ := c.Get("userID")
	username, _ := c.Get("username")

	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	if err := h.forwardService.Stop(uint(id), userID.(uint), username.(string), ip, ua); err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessWithMessage(c, "停止成功", nil)
}

// GetStats 获取转发统计
func (h *ForwardHandler) GetStats(c *gin.Context) {
	stats, err := h.forwardService.GetStats()
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, stats)
}
