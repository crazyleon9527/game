package entities

//消息通知
type Notification struct {
	BaseModel
	UID        uint    `gorm:"index" json:"uid"`
	Type       uint8   `gorm:"type:varchar(20);index" json:"type"` // 0  普通消息 1  模板消息
	Title      string  `gorm:"type:varchar(64)" json:"title"`
	TemplateID *uint   `gorm:"index" json:"-"`                     // 模板消息ID
	Params     *string `gorm:"type:text" json:"-"`                 // 模板消息参数
	Message    string  `gorm:"type:text" json:"message,omitempty"` // 消息内容  前端只需要关心消息内容
	Read       uint8   `json:"read"`                               //是否已读
	CreatedAt  int64   `json:"created_at"`
}

//消息模板
type NotificationTemplate struct {
	BaseModel
	Language string `gorm:"type:varchar(20)" json:"language"` // 语言
	Name     string `gorm:"type:varchar(32)" json:"name"`     // 模板类型，例如 sms、payment 等
	Title    string `gorm:"type:varchar(64)" json:"title"`    // 模板归属标题
	Content  string `gorm:"type:text" json:"content"`
}

type GetNotificationListReq struct {
	Paginator
	UID uint
}

type SendNotificationReq struct {
	UID        uint    `json:"uid"`
	Type       uint8   `json:"type"`        // 0: 普通消息, 1: 模板消息
	TemplateID *uint   `json:"template_id"` // 模板消息ID
	Params     *string `json:"params"`      // 模板消息参数
	Message    string  `json:"message"`     // 消息内容
}
