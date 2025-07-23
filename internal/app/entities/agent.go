package entities

// ////////////////////////////////////////////////////////////DB table ////////////////////////////////////////////////////////////////////////////////////////
// 代理关系表
type HallInviteRelation struct {
	BaseModel  `json:"-"` // 这里假设 BaseModel 已经有正确的 JSON 标签或不需要序列化
	PID        uint       `json:"pid" gorm:"column:pid;default:0"`
	UID        uint       `json:"uid" gorm:"column:uid;default:0;index"`
	Level      uint8      `json:"level" gorm:"column:level;default:0"`
	ReturnCash float64    `json:"return_cash" gorm:"column:return_cash;default:0;type:decimal(11,3)"`
	Mobile     string     `json:"mobile" gorm:"column:mobile;default:null;size:20"`
}

func (u *HallInviteRelation) AddReturnCash(cash float64) {
	u.ReturnCash = AddPrecise(u.ReturnCash, cash)
}

// 返利
type GameReturn struct {
	BaseModel    `json:"-"`
	UID          uint    `gorm:"column:uid;default:0"`
	Cash         float64 `gorm:"column:cash;default:0"`
	ReturnCash   float64 `gorm:"column:return_cash;type:decimal(10,3);default:0"`
	PID          uint    `gorm:"column:pid;default:0;index"`
	Level        uint8   `gorm:"column:level;default:0"`
	Percent      int     `gorm:"column:percent;default:0"`
	GetTime      int64   `gorm:"column:get_time"`
	Status       uint8   `gorm:"column:status;default:0"`
	Remark       string  `gorm:"column:remark;size:20"`
	PromoterCode int     `gorm:"column:pc"`
}

type RechargeReturn struct {
	BaseModel    `json:"-"`
	UID          uint    `gorm:"column:uid;default:0;index"`
	Cash         float64 `gorm:"column:cash;default:0"`
	ReturnCash   float64 `gorm:"column:return_cash;type:decimal(10,3);default:0"`
	PID          uint    `gorm:"column:pid;default:0;index"`
	Level        uint8   `gorm:"column:level;default:0"`
	Percent      int     `gorm:"column:percent;default:0"`
	GetTime      int64   `gorm:"column:get_time"`
	Status       uint8   `gorm:"column:status;default:0"`
	Remark       string  `gorm:"column:remark;size:36"`
	PromoterCode int     `gorm:"column:pc"`
}

type GameRebateReceipt struct {
	BaseModel   `json:"-"`
	UID         uint    `gorm:"column:uid;default:0"`
	Cash        float64 `gorm:"column:cash;default:0"`
	Status      uint8   `gorm:"column:status;default:0"`       //1已审核
	ReceiveTime int64   `gorm:"column:receive_time;default:0"` //领取时间
}

type GetGameRebateReceiptListReq struct {
	Paginator
	UID uint `json:"uid"`
}

// 返利比例
type RakeBack struct {
	Level  uint8 `gorm:"column:level;default:0"`
	Value  uint  `gorm:"column:value;default:0"`
	Status uint8 `gorm:"column:status;default:0"`
}

type LevelCountGroup struct {
	PID   uint `gorm:"column:pid"`
	Count int
}

// ////////////////////////////////////////////////////////////DB table ////////////////////////////////////////////////////////////////////////////////////////
type PromotionProfit struct {
	HasGet        float64        `json:"has_get"`         //已经领取的
	NotGet        float64        `json:"not_get"`         //未领取的
	LevelMap      map[string]int `json:"level_map"`       //每个级别拥有的人数
	Code          string         `json:"code"`            //邀请码
	TodayValidBet float64        `json:"today_valid_bet"` //今日有效投注
	TotalValidBet float64        `json:"total_valid_bet"` //总有效投注
	TodayLink     int            `json:"today_link"`      //今日邀请数
	InviteCount   int            `json:"invite_count"`    //总邀请数
}

type LevelRelationInfo struct {
	Num   int
	Level int
}

type GetPromotionListReq struct {
	Paginator
	UID   uint  `json:"uid"`
	Level uint8 `json:"level"`
}

type PromotionLink struct {
	InviteCode    string `json:"ic"`
	PromotionCode string `json:"pc"`
}

type FinalizeRechargeReturnReq struct {
	ID       uint `json:"id"`       // ID is the primary key identifier of the withdrawal request.
	OptionID uint `json:"optionID"` // SysUID refers to the system user ID associated with the request.
	IP       string
}
