package response

import (
	"gost-panel/internal/errors"
	"gost-panel/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`    // 业务状态码
	Message string      `json:"message"` // 响应消息
	Data    interface{} `json:"data"`    // 响应数据
}

// PageData 分页数据结构
type PageData struct {
	List     interface{} `json:"list"`     // 数据列表
	Total    int64       `json:"total"`    // 总数
	Page     int         `json:"page"`     // 当前页
	PageSize int         `json:"pageSize"` // 每页大小
}

// 业务状态码定义
const (
	CodeSuccess       = 0     // 成功
	CodeBadRequest    = 40000 // 请求参数错误
	CodeUnauthorized  = 40100 // 未授权
	CodeForbidden     = 40300 // 禁止访问
	CodeNotFound      = 40400 // 资源不存在
	CodeInternalError = 50000 // 服务器内部错误
)

// 常用响应消息
const (
	MsgSuccess       = "success"
	MsgBadRequest    = "请求参数错误"
	MsgUnauthorized  = "未授权，请先登录"
	MsgForbidden     = "禁止访问"
	MsgNotFound      = "资源不存在"
	MsgInternalError = "服务器内部错误"
)

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: MsgSuccess,
		Data:    data,
	})
}

// SuccessWithMessage 成功响应（自定义消息）
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

// SuccessPage 分页成功响应
func SuccessPage(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: MsgSuccess,
		Data: PageData{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	})
}

// Error 错误响应
func Error(c *gin.Context, httpCode int, businessCode int, message string) {
	c.JSON(httpCode, Response{
		Code:    businessCode,
		Message: message,
		Data:    nil,
	})
}

// BadRequest 请求参数错误响应
func BadRequest(c *gin.Context, message string) {
	if message == "" {
		message = MsgBadRequest
	}
	Error(c, http.StatusBadRequest, CodeBadRequest, message)
}

// Unauthorized 未授权响应
func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = MsgUnauthorized
	}
	Error(c, http.StatusUnauthorized, CodeUnauthorized, message)
}

// Forbidden 禁止访问响应
func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = MsgForbidden
	}
	Error(c, http.StatusForbidden, CodeForbidden, message)
}

// NotFound 资源不存在响应
func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = MsgNotFound
	}
	Error(c, http.StatusNotFound, CodeNotFound, message)
}

// InternalError 服务器内部错误响应
func InternalError(c *gin.Context, message string) {
	if message == "" {
		message = MsgInternalError
	}
	Error(c, http.StatusInternalServerError, CodeInternalError, message)
}

// HandleError 统一处理业务错误
// 自动识别 BizError 类型并返回对应的 HTTP 状态码和错误信息
func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	// 判断是否为业务错误
	if bizErr, ok := err.(*errors.BizError); ok {
		c.JSON(bizErr.HTTPCode, Response{
			Code:    bizErr.Code,
			Message: bizErr.Message,
			Data:    nil,
		})
		return
	}

	// 未知错误，记录日志并返回通用错误
	logger.Errorf("未知错误: %v", err)
	// 将实际错误信息返回给前端，而不是 generic message
	c.JSON(http.StatusInternalServerError, Response{
		Code:    CodeInternalError,
		Message: err.Error(),
		Data:    nil,
	})
}
