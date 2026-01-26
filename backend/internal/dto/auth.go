// Package dto 定义数据传输对象
package dto

// ==================== 认证相关 ====================

// LoginReq 登录请求
type LoginReq struct {
	Username string `json:"username" binding:"required"` // 用户名
	Password string `json:"password" binding:"required"` // 密码
}

// LoginResp 登录响应
type LoginResp struct {
	Token    string `json:"token"`     // JWT Token
	ExpireAt int64  `json:"expire_at"` // 过期时间戳
}

// ChangePasswordReq 修改密码请求
type ChangePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required"`       // 原密码
	NewPassword string `json:"new_password" binding:"required,min=6"` // 新密码
}

// UserInfoResp 用户信息响应
type UserInfoResp struct {
	ID       uint   `json:"id"`       // 用户 ID
	Username string `json:"username"` // 用户名
	Role     string `json:"role"`     // 角色
}
