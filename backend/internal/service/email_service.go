package service

import (
	"crypto/tls"
	"fmt"
	"gost-panel/internal/dto"
	"gost-panel/internal/errors"
	"net/smtp"
)

// SendTestEmail 发送测试邮件
func (s *SystemConfigService) SendTestEmail(req *dto.EmailConfigReq) error {
	// 验证必要参数
	if req.Host == "" || req.Port == 0 || req.FromEmail == "" {
		return errors.ErrSMTPConfigIncomplete
	}

	auth := smtp.PlainAuth("", req.Username, req.Password, req.Host)

	toEmail := req.FromEmail
	if req.ToEmail != "" {
		toEmail = req.ToEmail
	}

	to := []string{toEmail}

	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: Gost Panel 测试邮件\r\n"+
		"\r\n"+
		"这是一封来自 Gost Panel 的测试邮件。\r\n"+
		"如果您收到这封邮件，说明您的 SMTP 配置正确。\r\n", toEmail))

	addr := fmt.Sprintf("%s:%d", req.Host, req.Port)

	// 如果端口是 465，通常使用 SMTPS (隐式 TLS)
	// 如果端口是 587，通常使用 STARTTLS

	if req.Port == 465 {
		// SMTPS 需自定义 dialer
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true, // 允许自签名证书，生产环境建议关闭
			ServerName:         req.Host,
		}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return errors.ErrSMTPConnectFailed
		}

		client, err := smtp.NewClient(conn, req.Host)
		if err != nil {
			return errors.ErrSMTPClientFailed
		}
		defer func(client *smtp.Client) {
			_ = client.Close()
		}(client)

		if req.Username != "" && req.Password != "" {
			if err = client.Auth(auth); err != nil {
				return errors.ErrSMTPAuthFailed
			}
		}

		if err = client.Mail(req.FromEmail); err != nil {
			return errors.ErrSMTPSenderFailed
		}

		if err = client.Rcpt(to[0]); err != nil {
			return errors.ErrSMTPRecipientFailed
		}

		w, err := client.Data()
		if err != nil {
			return errors.ErrSMTPDataFailed
		}

		_, err = w.Write(msg)
		if err != nil {
			return errors.ErrSMTPWriteFailed
		}

		err = w.Close()
		if err != nil {
			return errors.ErrSMTPCloseFailed
		}

		return client.Quit()

	}

	// 标准 SMTP / STARTTLS
	return smtp.SendMail(addr, auth, req.FromEmail, to, msg)
}
