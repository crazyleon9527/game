package entities

// ////////////////////////////////////////////////////////////DB table ////////////////////////////////////////////////////////////////////////////////////////

type GmList struct {
	ID                  int     `gorm:"primaryKey;autoIncrement"`
	Name                string  `gorm:"size:50;not null"`
	SonGameList         string  `gorm:"type:text;not null"`
	Remarks             *string `gorm:"size:255"`
	Status              int8    `gorm:"not null;default:1"`
	Notice              string  `gorm:"size:255;not null;default:''"`
	Version             string  `gorm:"size:20;not null;default:1.0"`
	DownloadURL         string  `gorm:"size:100;not null;default:''"`
	MicroEndDownloadURL string  `gorm:"size:100;not null;default:''"`
	Platform            string  `gorm:"size:20;not null;default:''"`
	ImgURL              string  `gorm:"size:100;not null;default:''"`
	CreateTime          int     `gorm:"not null;default:0"`
	UpdateTime          int     `gorm:"not null;default:0"`
	HorseTimeInterval   int     `gorm:"default:0"`
	MinRecharge         float64 `gorm:"type:decimal(10,2);default:0.00"`
	MinWithdraw         float64 `gorm:"type:decimal(10,2);default:0.00"`
}

// 后台操作日志
type SystemOptionLog struct {
	ID         uint
	Type       uint8  //操作类型
	Result     string //操作结果
	Content    string //操作前内容
	Data       string //发送过来的数据
	UID        uint   `gorm:"column:uid" `         //操作的用户id
	Nickname   string `gorm:"column:nick_name" `   //操作的用户账号
	Username   string `gorm:"column:user_name" `   //操作的用户昵称
	OptionID   uint   `gorm:"column:oid" `         //操作者id
	OptionName string `gorm:"column:option_name" ` //操作者名字
	Remark     string //备注
	Time       int64  //操作时间
	IP         string //操作者IP
	Error      error  `gorm:"-" json:"-"`
}

func (t *SystemOptionLog) TableName() string {
	return "sys_option_log"
}

// ////////////////////////////////////////////////////////////DB table ////////////////////////////////////////////////////////////////////////////////////////

type FixInviteRelationReq struct {
	UID           uint `json:"uid"` //
	PromotionCode uint `json:"pc"`
	PID           uint `json:"pid"` //
	OptionID      uint `json:"optionID"`
	IP            string
}

type AddUserWithdrawCardReq struct {
	UID           uint   `json:"uid"`
	IFSC          string `json:"ifsc"`
	AccountNumber string `json:"accountNumber"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	OptionID      uint   `json:"optionID"`
	IP            string
}

type DelUserWithdrawCardReq struct {
	ID       uint `json:"id"`
	OptionID uint `json:"optionID"`
	IP       string
}

type FixUserWithdrawCardReq struct {
	ID            uint    `json:"id"`
	IFSC          *string `json:"ifsc"`
	AccountNumber *string `json:"accountNumber"`
	Name          *string `json:"name"`
	OptionID      uint    `json:"optionID"`
	IP            string
}

type RankStats struct {
	UID    uint    `gorm:"column:uid;"`
	Profit float64 `gorm:"column:profit;default:0;type:decimal(10,2)" `
	Code   int     `gorm:"column:pc;default:0" json:"-"`
}

type ReturnCash struct {
	PID  int
	Cash float64
}

type VerifyGoogleAuthCodeReq struct {
	Account  string `json:"account"`
	AuthCode string `json:"authCode"`
}
type SysUserArea struct {
	ID   uint `gorm:"column:uid;"`
	Area int  `gorm:"column:room;"`
}

type SysUserAdmin struct {
	ID               uint   `gorm:"column:uid;"`
	TelegramTeam     string `gorm:"column:telegram_team;"`
	TelegramCustomer string `gorm:"column:telegram_customer;"`
}

type MonthBackupAndCleaReq struct {
	TableNames string `json:"tableNames"`

	OptionID uint `json:"optionID"`
	IP       string
}
type CallChangePCReq struct {
	SRC uint `json:"pcSRC"`
	DST uint `json:"pcDST"`

	OptionID uint `json:"optionID"`
	IP       string
}
type ToggleSMSVerificationStateReq struct {
	OptionID uint `json:"optionID"`
	IP       string
}
type SMSVerificationStateResp struct {
	Disabled bool `json:"disabled"`
}
