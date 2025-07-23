package entities

type VerifyCodeReq struct {
	Target           string `json:"target" binding:"required,gte=8" example:"48576410@gmail.com"` // 手机号或者邮箱
	VerificationType uint8  `json:"verification_type" binding:"required" example:"1" `            // 验证类型：1-邮箱，2-短信
	BusinessType     uint8  `json:"business_type" binding:"required" example:"1"`                 // 业务类型 用户注册1,用户登录2,重置密码3,绑定手机4,绑定银行卡5,绑定邮箱6
}

// VerifyCode 验证码表结构
type VerifyCode struct {
	BaseModel        `json:"-"`
	Target           string `gorm:"column:target;index;size:42"`        // 手机号或邮箱
	VerificationType uint8  `gorm:"column:verification_type;default:0"` // 验证类型：1-邮箱，2-短信
	Code             string `gorm:"column:code;size:8"`                 // 验证码
	BusinessType     uint8  `gorm:"column:business_type;default:0"`     // 业务类型
	Status           uint8  `gorm:"column:status;default:0"`            // 验证码状态：0-未使用，1-已使用，2-已过期
	Count            uint8  `gorm:"column:count;default:0"`             // 当天发送次数
}

func (t *VerifyCode) TableName() string {
	return "verify_code"
}
