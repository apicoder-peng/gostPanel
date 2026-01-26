package middleware

import (
	"strings"

	"gost-panel/pkg/jwt"
	"gost-panel/pkg/response"

	"github.com/gin-gonic/gin"
)

// Auth JWT 认证中间件
func Auth(jwtInstance *jwt.JWT) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 Authorization 头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		// 检查 Bearer 前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Unauthorized(c, "认证格式错误")
			c.Abort()
			return
		}

		// 解析 Token
		claims, err := jwtInstance.ParseToken(parts[1])
		if err != nil {
			switch err {
			case jwt.ErrTokenExpired:
				response.Unauthorized(c, "登录已过期，请重新登录")
			case jwt.ErrTokenMalformed:
				response.Unauthorized(c, "Token 格式错误")
			default:
				response.Unauthorized(c, "Token 无效")
			}
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}
