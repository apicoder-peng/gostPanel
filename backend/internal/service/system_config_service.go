package service

import (
	"gost-panel/internal/dto"
	"gost-panel/internal/repository"
)

// SystemConfigService 系统配置服务
type SystemConfigService struct {
	repo *repository.SystemConfigRepository
}

// NewSystemConfigService 创建系统配置服务
func NewSystemConfigService(repo *repository.SystemConfigRepository) *SystemConfigService {
	return &SystemConfigService{repo: repo}
}

// GetConfig 获取配置DTO
func (s *SystemConfigService) GetConfig() (*dto.SystemConfigResp, error) {
	config, err := s.repo.Get()
	if err != nil {
		return nil, err
	}

	return &dto.SystemConfigResp{
		Panel: dto.PanelConfigResp{
			PanelURL: config.PanelURL,
		},
		Email: dto.EmailConfigResp{
			Host:      config.SMTPHost,
			Port:      config.SMTPPort,
			Username:  config.SMTPUsername,
			Password:  config.SMTPPassword,
			FromEmail: config.SMTPFrom,
		},
		Config: dto.PanelSettingResp{
			SiteTitle: config.SiteTitle,
			LogoURL:   config.LogoURL,
			Copyright: config.Copyright,
		},
		Log: dto.LogConfigResp{
			RetentionDays: config.LogRetentionDays,
			Level:         config.LogLevel,
		},
		Backup: dto.BackupConfigResp{
			AutoBackup:     config.AutoBackup,
			RetentionCount: config.BackupRetentionCount,
		},
	}, nil
}

// GetPublicConfig 获取公开配置
func (s *SystemConfigService) GetPublicConfig() (*dto.PublicSystemConfigResp, error) {
	config, err := s.repo.Get()
	if err != nil {
		return nil, err
	}

	return &dto.PublicSystemConfigResp{
		SiteTitle: config.SiteTitle,
		LogoURL:   config.LogoURL,
		Copyright: config.Copyright,
	}, nil
}

// UpdateConfig 更新配置
func (s *SystemConfigService) UpdateConfig(req *dto.UpdateSystemConfigReq) error {
	config, err := s.repo.Get()
	if err != nil {
		return err
	}

	// 映射 Panel
	config.PanelURL = req.Panel.PanelURL

	// 映射 Email
	config.SMTPHost = req.Email.Host
	config.SMTPPort = req.Email.Port
	config.SMTPUsername = req.Email.Username
	config.SMTPPassword = req.Email.Password
	config.SMTPFrom = req.Email.FromEmail

	// 映射 Config
	config.SiteTitle = req.Config.SiteTitle
	config.LogoURL = req.Config.LogoURL
	config.Copyright = req.Config.Copyright

	// 映射 Log
	config.LogRetentionDays = req.Log.RetentionDays
	config.LogLevel = req.Log.Level

	// 映射 Backup
	config.AutoBackup = req.Backup.AutoBackup
	config.BackupRetentionCount = req.Backup.RetentionCount

	return s.repo.Update(config)
}
