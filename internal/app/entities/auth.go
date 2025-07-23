package entities

type RegisterCredentials struct {
	Distribute
	Username string `json:"username" example:"zhangsan" binding:"omitempty,gte=3,lte=30"`     // 用户名
	Password string `json:"password" example:"123456" binding:"required,gte=6,lte=64"`        //密码
	Email    string `json:"email" example:"187884342432@gmail.com" binding:"omitempty,gte=8"` //手机号，这里假设手机号是按照e164规则验证
	Mobile   string `json:"mobile" example:"+91187884342432" binding:"omitempty,gte=9"`       //手机号，这里假设手机号是按照e164规则验证
	VerCode  string `json:"verCode" binding:"required,numeric"`                               //验证码
	LoginIP  string `json:"loginIP" binding:"omitempty,ip"`                                   //登录IP，这里假设是IP格式验证
	IsRobot  uint8  `json:"isRobot"`
	Telegram string `json:"telegram"` // 电报

	Device string `json:"-"` // 设备信息

	IsOAuth bool `json:"-"` //是否是第三方登录注册
	Status  uint8
}

type VerifyCredentials struct {
	Username string `json:"username" binding:"omitempty"`
	Password string `json:"password" binding:"omitempty"`
}

type LoginCredentials struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	LoginIP  string `json:"-"`
}

type MobileLoginCredentials struct {
	Mobile  string `json:"mobile" binding:"required"`
	VerCode string `json:"verCode" binding:"required,numeric"` //验证码，这里假设是6位数的验证码
	LoginIP string `json:"-"`                                  //登录IP
}

type ChangePasswordCredentials struct {
	Username        string `json:"username" binding:"required"`                             //用户名
	Password        string `json:"password" binding:"required"`                             //旧密码
	NewPassword     string `json:"newPassword" binding:"required,gte=6,lte=30"`             //新密码，这里假设密码长度应在6到30个字符之间
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"` //确认密码，确保和新密码相同
}

type ResetPasswordCredentials struct {
	Password string `json:"password" binding:"required,gte=6,lte=30"` //密码
	Mobile   string `json:"emailOrMobile" binding:"required,gte=8"`   //手机号
	VerCode  string `json:"verCode" binding:"required,numeric"`       //验证码，这里假设是6位数的验证码
	LoginIP  string `json:"-"`                                        //登录IP
}
