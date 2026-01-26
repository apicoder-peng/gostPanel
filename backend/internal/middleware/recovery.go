package middleware

import (
	"fmt"
	"gost-panel/internal/errors"
	"gost-panel/pkg/logger"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Recovery 全局异常恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录堆栈信息
				stack := string(debug.Stack())
				logger.Errorf("系统发生 Panic: %v\nStack: %s", err, stack)

				// 返回统一的错误响应
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    errors.ErrInternal.Code,
					"message": fmt.Sprintf("系统内部错误: %v", err),
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// ErrorHandler 全局错误处理中间件
// 用于捕获 c.Error() 添加的错误并统一格式化
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		//如果没有发生panic，但有 errors
		if len(c.Errors) > 0 {
			// 获取最后一个错误
			err := c.Errors.Last()

			// 此时 response 可能已经写入了，如果没写入则写入 JSON
			if !c.Writer.Written() {
				// 尝试转换由于 checkError 等工具函数可能产生的自定义错误
				// 这里简单处理，如果有 bizError 则使用，否则包装

				// 默认 500
				status := http.StatusInternalServerError
				resp := gin.H{
					"code":    50000,
					"message": err.Error(),
				}

				// 检查是否是自定义业务错误
				if bizErr, ok := err.Err.(*errors.BizError); ok {
					status = bizErr.HTTPCode
					resp["code"] = bizErr.Code
					resp["message"] = bizErr.Message
				}

				c.JSON(status, resp)
			}
		}
	}
}
