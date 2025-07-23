package entities

import "encoding/json"

// ////////////////////////////////////////////////////////////DB table start////////////////////////////////////////////////////////////////////////////////////////

// 流水表
type Flow struct {
	BaseModel    `json:"-"`
	UID          uint    `gorm:"column:uid;default:0;index" json:"uid"`
	FlowType     uint16  `gorm:"column:type;default:0;index" json:"type"`
	Currency     string  `gorm:"column:currency;size:12" json:"currency"`
	IsRobot      uint8   `gorm:"column:is_robot;default:0" json:"isRobot"`
	Number       float64 `gorm:"column:number;type:decimal(20,2);default:0" json:"number"`
	Balance      float64 `gorm:"column:balance;type:decimal(20,3);default:0" json:"balance"`
	Remark       string  `gorm:"column:remark;size:72" json:"remark"`
	PromoterCode int     `gorm:"column:pc" json:"pc"`
}

// // 自定义JSON序列化功能
func (wp *Flow) MarshalJSON() ([]byte, error) {
	type Alias Flow
	return json.Marshal(&struct {
		CreatedAt int64 `json:"createdAt"`
		*Alias
	}{
		CreatedAt: wp.CreatedAt,
		Alias:     (*Alias)(wp),
	})
}

// 游戏下注返利
type RefundGameFlow struct {
	BaseModel    `json:"-"`
	UID          uint    `gorm:"column:uid;default:0;index"`
	FlowID       uint    `gorm:"column:fid;default:0"`
	FlowType     uint8   `gorm:"column:type;default:0;index"`
	Currency     string  `gorm:"column:currency;size:12"`
	Number       float64 `gorm:"column:number;type:decimal(20,2);default:0"`
	Status       uint8   `gorm:"column:status;default:0;index:index_status"`
	PromoterCode int     `gorm:"column:pc" json:"pc"`
}

// 游戏下注返利
type RefundLinkGameFlow struct {
	BaseModel    `json:"-"`
	UID          uint    `gorm:"column:uid;default:0;index"`
	FlowID       uint    `gorm:"column:fid;default:0"`
	FlowType     uint16  `gorm:"column:type;default:0"`
	Currency     string  `gorm:"column:currency;size:12"`
	Number       float64 `gorm:"column:number;type:decimal(20,2);default:0"`
	Status       uint8   `gorm:"column:status;default:0;index:index_status"`
	PromoterCode int     `gorm:"column:pc;index" json:"pc"`
}

// ////////////////////////////////////////////////////////////DB table end////////////////////////////////////////////////////////////////////////////////////////
type GetFlowListReq struct {
	Paginator
	UID uint
}
