package entities

type ZfTransferOrder struct {
	BaseModel
	UID          uint    `gorm:"index;not null"`             // 用户id，非空
	RoundID      int     `gorm:"column:round_id"`            // 局数
	BetID        int     `gorm:"column:bet_id"`              // 投注ID
	GameCode     string  `gorm:"type:varchar(32);not null;"` // 游戏代码
	Amount       float64 `gorm:"type:decimal(10,2)"`         // 金额，十进制
	Type         int     `gorm:"type:int(3)"`                // 类型
	RewardAmount float64 `gorm:"type:decimal(10,2)"`
	// Opencode     string  `gorm:"type:varchar(32);not null;default:'0.00'"`
	Status uint8 `gorm:"column:status;default:0" json:"status"` // 是否已经结算
}

type ZfBetRecord struct {
	RecordID  string  `json:"uniqueid" gorm:"type:varchar(32);not null;"`
	Account   string  `json:"username"  gorm:"type:varchar(36)"`
	GameCode  string  `json:"game_code"  gorm:"type:varchar(36)"`
	BetAmount float64 `json:"bet_amount" gorm:"type:decimal(10,2)"`
	CreatedAt int64   `json:"bet_time"`
}

type ZfGameLoginReq struct {
	GameCode string `json:"gameCode"`
	UID      uint
}

type ZfKickReq struct {
	GameCode string `json:"gameCode"`
	UID      uint
}

type ZfGameReq struct {
	UniqueID     string `json:"unique_id"`
	Timestamp    int    `json:"timestamp"`
	MerchantCode string `json:"merchant_code"`
}

type ZfRefund struct {
	ZfGameReq

	Username string  `json:"username"`
	Sign     string  `json:"sign"`
	GameCode string  `json:"game_code"`
	BetID    int     `json:"bet_id"`
	RoundID  int     `json:"round_id"`
	Amount   float64 `json:"amount"`
	Type     int     `json:"type"` //1: refund 2: payout failed 3: issue cancel; 1:退回 2:派彩失败 3:取消
	Currency string  `json:"currency"`
	FlowType uint
}

type ZfBetReq struct {
	ZfGameReq

	Sign     string  `json:"sign"`
	Username string  `json:"username"`
	GameCode string  `json:"game_code"`
	BetID    int     `json:"bet_id"`
	RoundID  int     `json:"round_id"`
	Currency string  `json:"currency"`
	Number   string  `json:"number"`
	Amount   float64 `json:"amount"`
}

type ZfPayoutReq struct {
	ZfGameReq

	Sign     string  `json:"sign"`
	Username string  `json:"username"`
	GameCode string  `json:"game_code"`
	BetID    int     `json:"bet_id"`
	RoundID  int     `json:"round_id"` // RoundID可以是string或number
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Number   string  `json:"number"`
}

type ZfBalanceReq struct {
	ZfGameReq
	Sign     string `json:"sign"`
	Username string `json:"username"`
}

type ZfSettleReq struct {
	ZfGameReq
	RoundID  int         `json:"round_id"` // RoundID可以是string或number
	GameCode string      `json:"game_code"`
	Opencode interface{} `json:"opencode"` // 文档写着string，实际是个number
}
