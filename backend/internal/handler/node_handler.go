// Package handler 提供 HTTP 请求处理器
package handler

import (
	"strconv"

	"gost-panel/internal/dto"
	"gost-panel/internal/service"
	"gost-panel/pkg/response"

	"github.com/gin-gonic/gin"
)

// NodeHandler 节点控制器
// 处理节点相关的 HTTP 请求
type NodeHandler struct {
	nodeService *service.NodeService
}

// NewNodeHandler 创建节点控制器
func NewNodeHandler(nodeService *service.NodeService) *NodeHandler {
	return &NodeHandler{nodeService: nodeService}
}

// Create 创建节点
// @Summary 创建节点
// @Tags 节点管理
// @Accept json
// @Produce json
// @Param body body dto.CreateNodeReq true "创建节点请求"
// @Success 200 {object} response.Response
// @Router /api/v1/nodes [post]
func (h *NodeHandler) Create(c *gin.Context) {
	var req dto.CreateNodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, _ := c.Get("userID")
	username, _ := c.Get("username")

	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	node, err := h.nodeService.Create(&req, userID.(uint), username.(string), ip, ua)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, node)
}

// Update 更新节点
// @Summary 更新节点
// @Tags 节点管理
// @Accept json
// @Produce json
// @Param id path int true "节点ID"
// @Param body body dto.UpdateNodeReq true "更新节点请求"
// @Success 200 {object} response.Response
// @Router /api/v1/nodes/{id} [put]
func (h *NodeHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的节点 ID")
		return
	}

	var req dto.UpdateNodeReq
	if err = c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, _ := c.Get("userID")
	username, _ := c.Get("username")

	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	node, err := h.nodeService.Update(uint(id), &req, userID.(uint), username.(string), ip, ua)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, node)
}

// Delete 删除节点
// @Summary 删除节点
// @Tags 节点管理
// @Param id path int true "节点ID"
// @Success 200 {object} response.Response
// @Router /api/v1/nodes/{id} [delete]
func (h *NodeHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的节点 ID")
		return
	}

	userID, _ := c.Get("userID")
	username, _ := c.Get("username")

	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	if err = h.nodeService.Delete(uint(id), userID.(uint), username.(string), ip, ua); err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// GetByID 获取节点详情
// @Summary 获取节点详情
// @Tags 节点管理
// @Param id path int true "节点ID"
// @Success 200 {object} response.Response
// @Router /api/v1/nodes/{id} [get]
func (h *NodeHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的节点 ID")
		return
	}

	node, err := h.nodeService.GetByID(uint(id))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, node)
}

// List 获取节点列表
// @Summary 获取节点列表
// @Tags 节点管理
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param status query string false "状态筛选"
// @Param keyword query string false "关键词搜索"
// @Success 200 {object} response.Response
// @Router /api/v1/nodes [get]
func (h *NodeHandler) List(c *gin.Context) {
	var req dto.NodeListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	nodes, total, err := h.nodeService.List(&req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessPage(c, nodes, total, req.Page, req.PageSize)
}

// GetStats 获取节点统计
// @Summary 获取节点统计
// @Tags 节点管理
// @Success 200 {object} response.Response
// @Router /api/v1/nodes/stats [get]
func (h *NodeHandler) GetStats(c *gin.Context) {
	stats, err := h.nodeService.GetStats()
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, stats)
}

// GetConfig 获取节点配置
// @Summary 获取节点 Gost 配置
// @Tags 节点管理
// @Param id path int true "节点ID"
// @Success 200 {object} response.Response
// @Router /api/v1/nodes/{id}/config [get]
func (h *NodeHandler) GetConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的节点 ID")
		return
	}

	config, err := h.nodeService.GetConfig(uint(id))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, config)
}
