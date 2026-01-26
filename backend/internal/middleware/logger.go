package middleware

import (
	"time"

	"gost-panel/pkg/logger"

	"github.com/gin-gonic/gin"
)

// Logger 请求日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()

		// 执行时间
		latency := endTime.Sub(startTime)

		// 请求信息
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		// 记录日志
		if statusCode >= 400 {
			logger.Warnf("[%d] %s %s - %s - %v - %d bytes",
				statusCode, method, path, clientIP, latency, bodySize)
		} else {
			logger.Infof("[%d] %s %s - %s - %v - %d bytes",
				statusCode, method, path, clientIP, latency, bodySize)
		}
	}
}
