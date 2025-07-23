package entities

////////////////////////////////////////////////////////////DB table ////////////////////////////////////////////////////////////////////////////////////////

// INSERT INTO `pinduo_setting` VALUES (1, '拼多多活动', 999.00,  1651140650, 1653746250, 20, 13.300, 2.800)
// PinduoSetting 结构体存储了拼多多活动设置
type PinduoSetting struct {
	ID        int     `json:"id" gorm:"primaryKey"` // 主键
	Name      string  //名称
	Amount    float64 `gorm:"column:amount;default:0;type:decimal(10,2)" json:"amount"`      // 总金额
	StartTime int64   `json:"startTime"`                                                     // 开始时间
	EndTime   int64   `json:"endTime"`                                                       // 结束时间
	Base      float64 `gorm:"column:base;default:0;type:decimal(10,2)" json:"base"`          // 基础金额
	AbxRatio  float64 `gorm:"column:abx_ratio;default:0;type:decimal(10,2)" json:"abxRatio"` // ABX比例
	CbxRatio  float64 `gorm:"column:cbx_ratio;default:0;type:decimal(10,2)" json:"cbxRatio"` // CBX比例
}

// PinduoRecord 结构体记录了拼多多活动的用户数据
type PinduoRecord struct {
	BaseModel
	UID         uint    `json:"uid" gorm:"index"`                                                    // 用户ID
	UserName    string  `json:"userame"`                                                             // 用户名
	LockCash    float64 `gorm:"column:lock_cash;default:0;type:decimal(10,2)" json:"lockCash"`       // 锁定现金
	SuggestCash float64 `gorm:"column:suggest_cash;default:0;type:decimal(10,2)" json:"suggestCash"` // 建议现金
	InviteCount int     `json:"inviteCount"`                                                         // 邀请数量
	PinID       int     `json:"pinID"`                                                               // 拼团ID
	Created     int64   `json:"created"`                                                             // 创建时间
	Status      int     `json:"status" gorm:"index"`                                                 // 状态
}

// 红包设置表
type HongbaoSetting struct {
	BaseModel     `json:"-"`
	Name          string  `gorm:"column:name;size:32;;unique"`                //红包名
	Amount        float64 `gorm:"column:amount;type:decimal(10,2);default:0"` //红包金额
	Number        uint    `gorm:"column:number;default:0"`                    //红包数量
	ReceiveNumber uint    `gorm:"column:receive_number;default:0"`            //接收数目
	Remark        string  `gorm:"column:remark"`                              //备注
	Type          uint8   `gorm:"column:type;default:0"`                      //领取类型
	SYSUID        uint    `gorm:"column:sysuid;default:0"`                    //添加人
	Status        uint8   `gorm:"column:status;default:0"`                    //状态
}

// 红包设置表
type HongbaoRecord struct {
	BaseModel `json:"-"`
	UID       uint    `gorm:"column:uid;default:0"`                        //用户ID
	Username  string  `gorm:"column:username;size:32"`                     //用户名
	HongID    uint    `gorm:"column:hong_id;default:0"`                    //红包ID
	HongName  string  `gorm:"column:hong_name;size:64"`                    //红包名
	Amount    float64 `gorm:"column:amount;default:0;type:decimal(10,2)" ` //抢到金额
	// Remark     string  `gorm:"column:remark;size:64;"`                      //备注
	PromoterCode int `gorm:"column:pc;default:0" json:"-"`
}

// ////////////////////////////////////////////////////////////DB table ////////////////////////////////////////////////////////////////////////////////////////

type Activity struct {
	BaseModel
	Title          string `gorm:"size:255;not null" json:"title"`  // 活动标题
	Name           string `gorm:"size:255;not null" json:"name"`   // 活动名字
	Description    string `gorm:"type:text" json:"description"`    // 活动描述
	StartTime      int64  `gorm:"not null" json:"start_time"`      // 活动开始时间
	EndTime        int64  `gorm:"not null" json:"end_time"`        // 活动结束时间
	Status         uint8  `gorm:"size:50;not null" json:"status"`  // 活动激活状态 (例如 "active", "inactive", "paused")
	Priority       int    `gorm:"default:0" json:"priority"`       // 活动优先级 (较高数字表示较高优先级)
	PromotionImage string `gorm:"size:255" json:"promotion_image"` // 活动宣传图 URL
}

// Banner 结构体定义了 banner 的信息
type Banner struct {
	BaseModel
	Title    string `gorm:"size:255;not null" json:"title"`            // 活动标题
	Status   uint8  `gorm:"column:status;default:1" json:"status"`     // 状态 0 1
	Priority uint32 `gorm:"column:priority;default:0" json:"priority"` // 优先级 (较高数字表示较高优先级)
	JumpURL  string `gorm:"column:jump_url;size:255" json:"jumpURL"`   // 跳转 URL
	Image    string `gorm:"column:image;size:255" json:"image"`        // 图片
}

// Logo 结构体定义了 logo 的信息
type Logo struct {
	BaseModel
	Title    string `gorm:"size:255;not null" json:"title"`            // 活动标题
	Status   uint8  `gorm:"column:status;default:1" json:"status"`     // 状态 0 1
	Priority uint32 `gorm:"column:priority;default:0" json:"priority"` // 优先级 (较高数字表示较高优先级)
	JumpURL  string `gorm:"column:jump_url;size:255" json:"jumpURL"`   // 跳转 URL
	Image    string `gorm:"column:image;size:255" json:"image"`        // 图片
	Type     uint8  `gorm:"column:type;default:0" json:"type"`         // 类型 1.平台主界面左上角 2.平台跳转游戏加载页 3.游戏加载页)
}

// PinduoInfo 结构体定义了拼多多活动的信息
type PinduoInfo struct {
	EndTime     int64   `json:"endTime"`     // 结束时间
	InviteCount uint    `json:"inviteCount"` // 邀请数量
	Status      int     `json:"status"`      // 状态
	LockCash    float64 `json:"lockCash"`    // 锁定现金
	SuggestCash float64 `json:"suggestCash"` // 建议现金
}

type HallInvitePinduo struct {
	InviteID   uint
	InviteName string
	UID        uint
}

type GetRedEnvelopeReq struct {
	RedName string `json:"red_name" binding:"required"`
	Type    uint8  `json:"type" binding:"gte=0"`
	UID     uint
}

type AddRedEnvelopeReq struct {
	UID      uint
	Remark   string  `json:"remark"`
	Type     uint8   `json:"type"`
	Amount   float64 `json:"amount"`   //红包金额
	Number   uint    `json:"number"`   //红包数量
	OptionID uint    `json:"optionID"` //添加人
	IP       string
}

type DelRedEnvelopeReq struct {
	Name     string `json:"name"`     //唯一标识名字
	OptionID uint   `json:"optionID"` //添加人
	IP       string
}

type GetPinduoReq struct {
	Type uint8 `json:"type" binding:"gte=0"`
	UID  uint
}
