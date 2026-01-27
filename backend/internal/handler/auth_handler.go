package handler

import (
	"gost-panel/internal/dto"
	"gost-panel/internal/service"
	"gost-panel/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证控制器
// 处理用户登录、Token 刷新和密码修改等请求
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 创建认证控制器
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	result, err := h.authService.Login(&req, ip, userAgent)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, result)
}

// GetUserInfo 获取当前用户信息
func (h *AuthHandler) GetUserInfo(c *gin.Context) {
	userID, _ := c.Get("userID")

	user, err := h.authService.GetUserByID(userID.(uint))
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, dto.UserInfoResp{
		ID:       user.ID,
		Username: user.Username,
		Role:     "admin",
	})
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, _ := c.Get("userID")
	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	if err := h.authService.ChangePassword(userID.(uint), &req, ip, userAgent); err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessWithMessage(c, "密码修改成功", nil)
}

// RefreshToken 刷新 Token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		response.Unauthorized(c, "缺少 Authorization 头")
		return
	}

	// 从 Bearer token 中提取 token
	tokenString := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	} else {
		tokenString = authHeader
	}

	newToken, err := h.authService.RefreshToken(tokenString)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{
		"token": newToken,
	})
}
