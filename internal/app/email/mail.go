package email

// IProvider 邮箱提供商接口
type IProvider interface {
	Send(to []string, subject string, body string) error
}

type EmailData struct {
	Code         string // 验证码
	BusinessType string // 业务类型
}

// SMTPConfig SMTP 配置
type SMTPConfig struct {
	Server   string // SMTP 服务器地址
	Port     int    // SMTP 端口
	Username string // SMTP 用户名
	Password string // SMTP 密码
	From     string // 发件人邮箱
	Nickname string // 发件人昵称
}
