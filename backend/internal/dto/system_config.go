package dto

// ==================== 系统配置相关 ====================

// SystemConfigResp 系统配置响应
type SystemConfigResp struct {
	Panel  PanelConfigResp  `json:"panel"`
	Email  EmailConfigResp  `json:"email"`
	Config PanelSettingResp `json:"config"`
	Log    LogConfigResp    `json:"log"`
	Backup BackupConfigResp `json:"backup"`
}

type PublicSystemConfigResp struct {
	SiteTitle string `json:"siteTitle"`
	LogoURL   string `json:"logoUrl"`
	Copyright string `json:"copyright"`
}

type PanelConfigResp struct {
	PanelURL string `json:"panelUrl"`
}

type EmailConfigResp struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	FromEmail string `json:"fromEmail"`
}

type PanelSettingResp struct {
	SiteTitle string `json:"siteTitle"`
	LogoURL   string `json:"logoUrl"`
	Copyright string `json:"copyright"`
}

type LogConfigResp struct {
	RetentionDays int    `json:"retentionDays"`
	Level         string `json:"level"`
}

type BackupConfigResp struct {
	AutoBackup     bool `json:"autoBackup"`
	RetentionCount int  `json:"retentionCount"`
}

// UpdateSystemConfigReq 更新系统配置请求
type UpdateSystemConfigReq struct {
	Panel  PanelConfigReq  `json:"panel"`
	Email  EmailConfigReq  `json:"email"`
	Config PanelSettingReq `json:"config"`
	Log    LogConfigReq    `json:"log"`
	Backup BackupConfigReq `json:"backup"`
}

type PanelConfigReq struct {
	PanelURL string `json:"panelUrl"`
}

type EmailConfigReq struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	FromEmail string `json:"fromEmail"`
	ToEmail   string `json:"toEmail"` // 测试邮件接收人
}

type PanelSettingReq struct {
	SiteTitle string `json:"siteTitle"`
	LogoURL   string `json:"logoUrl"`
	Copyright string `json:"copyright"`
}

type LogConfigReq struct {
	RetentionDays int    `json:"retentionDays"`
	Level         string `json:"level"`
}

type BackupConfigReq struct {
	AutoBackup     bool `json:"autoBackup"`
	RetentionCount int  `json:"retentionCount"`
}
