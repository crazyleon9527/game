package email

import (
	"crypto/tls"

	"gopkg.in/mail.v2"
)

type OutlookProvider struct {
	config SMTPConfig
}

func NewOutlookProvider(cfg SMTPConfig) *OutlookProvider {
	return &OutlookProvider{config: cfg}
}

func (p *OutlookProvider) Send(to []string, subject string, body string) error {
	m := mail.NewMessage()
	m.SetHeader("From", m.FormatAddress(p.config.From, p.config.Nickname))
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := mail.NewDialer(
		"smtp.office365.com", // Outlook SMTP地址
		587,                  // 必须使用TLS端口
		p.config.Username,    // 完整邮箱（如your@outlook.com）
		p.config.Password,    // 此处填应用密码（非邮箱登录密码）
	)
	d.TLSConfig = &tls.Config{ServerName: "smtp.office365.com"} // 必需TLS配置

	return d.DialAndSend(m)
}

// hfeyhiijxvnlhgdh
