package email

import "gopkg.in/mail.v2"

// GmailProvider Gmail 邮箱提供商
type GmailProvider struct {
	config SMTPConfig
}

func NewGmailProvider(config SMTPConfig) *GmailProvider {
	return &GmailProvider{config: config}
}

func (p *GmailProvider) Send(to []string, subject string, body string) error {
	m := mail.NewMessage()
	m.SetHeader("From", p.config.Nickname+" <"+p.config.From+">")
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := mail.NewDialer(p.config.Server, p.config.Port, p.config.Username, p.config.Password)
	return d.DialAndSend(m)
}
