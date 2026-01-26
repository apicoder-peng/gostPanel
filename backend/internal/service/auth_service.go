// Package service 提供业务逻辑层服务
package service

import (
	stderrors "errors"

	"gost-panel/internal/dto"
	"gost-panel/internal/errors"
	"gost-panel/internal/model"
	"gost-panel/internal/repository"
	"gost-panel/pkg/jwt"
	"gost-panel/pkg/logger"

	"gorm.io/gorm"
)

// AuthService 认证服务
// 负责用户登录、Token 管理和密码修改
type AuthService struct {
	userRepo   *repository.UserRepository
	logService *LogService
	jwt        *jwt.JWT
}

// NewAuthService 创建认证服务
func NewAuthService(db *gorm.DB, jwtCfg *jwt.Config) *AuthService {
	return &AuthService{
		userRepo:   repository.NewUserRepository(db),
		logService: NewLogService(db),
		jwt:        jwt.New(jwtCfg),
	}
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token    string      `json:"token"`
	ExpireAt int64       `json:"expire_at"`
	User     *model.User `json:"user"`
}

// Login 用户登录
func (s *AuthService) Login(req *dto.LoginReq, ip, userAgent string) (*LoginResponse, error) {
	// 查询用户
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrInvalidCredentials
		}
		logger.Errorf("查询用户失败: %v", err)
		return nil, err
	}

	// 验证密码
	if !user.CheckPassword(req.Password) {
		return nil, errors.ErrInvalidCredentials
	}

	// 生成 Token
	token, err := s.jwt.GenerateToken(user.ID, user.Username)
	if err != nil {
		logger.Errorf("生成 Token 失败: %v", err)
		return nil, errors.ErrTokenGenerationFailed
	}

	// 记录登录日志
	s.logService.Record(
		user.ID,
		user.Username,
		model.ActionLogin,
		"",
		0,
		"",
		ip,
		userAgent)

	return &LoginResponse{
		Token:    token,
		ExpireAt: 0, // TODO: 从配置中获取
		User:     user,
	}, nil
}

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(userID uint, req *dto.ChangePasswordReq, ip, userAgent string) error {
	// 查询用户
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrUserNotFound
		}
		return err
	}

	// 验证旧密码
	if !user.CheckPassword(req.OldPassword) {
		return errors.ErrPasswordMismatch
	}

	// 设置新密码
	if err = user.SetPassword(req.NewPassword); err != nil {
		return err
	}

	// 更新密码
	if err = s.userRepo.UpdatePassword(userID, user.Password); err != nil {
		logger.Errorf("更新密码失败: %v", err)
		return err
	}

	// 记录操作日志
	s.logService.Record(
		userID,
		user.Username,
		model.ActionChangePassword,
		"",
		0,
		"",
		ip,
		userAgent)

	return nil
}

// RefreshToken 刷新 Token
func (s *AuthService) RefreshToken(tokenString string) (string, error) {
	return s.jwt.RefreshToken(tokenString)
}

// ParseToken 解析 Token
func (s *AuthService) ParseToken(tokenString string) (*jwt.Claims, error) {
	return s.jwt.ParseToken(tokenString)
}

// GetUserByID 根据 ID 获取用户
func (s *AuthService) GetUserByID(id uint) (*model.User, error) {
	return s.userRepo.FindByID(id)
}

// InitDefaultAdmin 初始化默认管理员
func (s *AuthService) InitDefaultAdmin(username, password string) error {
	// 检查是否已存在
	exists, err := s.userRepo.ExistsByUsername(username)
	if err != nil {
		return err
	}
	if exists {
		logger.Infof("管理员账号 %s 已存在", username)
		return nil
	}

	// 创建管理员
	admin := &model.User{
		Username: username,
		Password: password,
	}
	if err = s.userRepo.Create(admin); err != nil {
		return err
	}

	logger.Infof("默认管理员账号已创建: %s", username)
	return nil
}
