package email

import "gopkg.in/mail.v2"

// ZohoProvider Zoho 邮箱提供商
type ZohoProvider struct {
	config SMTPConfig
}

// NewZohoProvider 创建一个新的 ZohoProvider 实例
func NewZohoProvider(config SMTPConfig) *ZohoProvider {
	return &ZohoProvider{config: config}
}

// Send 发送邮件
func (p *ZohoProvider) Send(to []string, subject string, body string) error {
	// 创建邮件消息
	m := mail.NewMessage()
	m.SetHeader("From", p.config.Nickname+" <"+p.config.From+">") // 设置发件人
	m.SetHeader("To", to...)                                      // 设置收件人
	m.SetHeader("Subject", subject)                               // 设置邮件主题
	m.SetBody("text/html", body)                                  // 设置邮件正文（HTML 格式）

	// Zoho SMTP 服务器地址和端口
	// smtp.zoho.com:587 (TLS) 或 smtp.zoho.com:465 (SSL)
	d := mail.NewDialer(p.config.Server, p.config.Port, p.config.Username, p.config.Password)

	// Zoho 使用强制 TLS
	d.StartTLSPolicy = mail.MandatoryStartTLS

	// 发送邮件
	return d.DialAndSend(m)
}
