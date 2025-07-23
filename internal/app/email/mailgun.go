package email

import "gopkg.in/mail.v2"

// MailgunProvider Mailgun 邮箱提供商
type MailgunProvider struct {
	config SMTPConfig
}

// NewMailgunProvider 创建一个新的 MailgunProvider 实例
func NewMailgunProvider(config SMTPConfig) *MailgunProvider {
	return &MailgunProvider{config: config}
}

// Send 发送邮件
func (p *MailgunProvider) Send(to []string, subject string, body string) error {
	// 创建邮件消息
	m := mail.NewMessage()
	m.SetHeader("From", p.config.Nickname+" <"+p.config.From+">") // 设置发件人
	m.SetHeader("To", to...)                                      // 设置收件人
	m.SetHeader("Subject", subject)                               // 设置邮件主题
	m.SetBody("text/html", body)                                  // 设置邮件正文（HTML 格式）

	// 创建 SMTP 拨号器
	d := mail.NewDialer(p.config.Server, p.config.Port, p.config.Username, p.config.Password)

	// 强制使用 TLS 加密连接
	d.StartTLSPolicy = mail.MandatoryStartTLS

	// 发送邮件
	return d.DialAndSend(m)
}
