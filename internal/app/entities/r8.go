package entities

type R8TransferOrder struct {
	BaseModel
	UID        uint    `gorm:"index;not null"`                                         // 用户id，非空
	RoundID    string  `gorm:"type:varchar(32)"`                                       //
	RecordID   string  `gorm:"type:varchar(36)"`                                       // 局
	TransferNo string  `gorm:"type:varchar(36);index:idx_transfer_no_action"`          // 投注id，
	GameCode   string  `gorm:"type:varchar(32);not null;"`                             // 游戏代码
	Amount     float64 `gorm:"type:decimal(10,2)"`                                     // 金额，十进制
	Action     string  `gorm:"type:varchar(10);not null;index:idx_transfer_no_action"` // 行动
	CreateTime int64   `gorm:"type:int(11)"`
}

type R8BetRecord struct {
	GameCode  string  `json:"game_code" gorm:"type:varchar(32);not null;"` // 游戏代码
	RecordID  string  `json:"record_id" gorm:"type:varchar(36)"`           // 局
	Account   string  `json:"account" gorm:"type:varchar(10)"`             // 账号
	BetAmount float64 `json:"bet_amount" gorm:"type:decimal(10,2)"`        // 投注金额
	CreatedAt int64   `json:"created_at" `                                 // 创建时间
}

type R8ActivityOrder struct {
	BaseModel            // 加入gorm.Model包括了基本的ID、CreatedAt等字段
	UID          uint    `gorm:"index;not null"`                         // 用户id，非空
	EventID      string  `json:"event_id" gorm:"type:varchar(42);"`      // 使用varchar(100)类型，并增加索引
	AwardID      string  `json:"award_id" gorm:"type:varchar(64);index"` // 索引有助于提高查询效率
	ActivityType string  `json:"activity_type" gorm:"type:varchar(50);"` // 根据实际需求，设置类型和索引
	Account      string  `json:"account" gorm:"type:varchar(100);"`      // 类型和索引
	Action       string  `json:"action" gorm:"type:varchar(50)"`         // 可能不需要索引
	Currency     string  `json:"currency" gorm:"type:varchar(10)"`       // 指定类型，这里按照Currency2类型来
	Money        float64 `json:"money" gorm:"type:decimal(10,2)"`        // decimal类型用于金钱，保留两位小数
}

type GameSetting struct {
	BaseModel `json:"-"`
}

type GameLauncherReq struct {
	UID uint

	GameID  string
	TableID string
}

type R8GameLoginReq struct {
	GameCode string `json:"gameCode"`
	UID      uint
}

type R8GameLoginResp struct {
	Url string
}

type R8Transfer struct {
	Account    string  `json:"account"`
	Action     string  `json:"action"`
	Money      float64 `json:"money"`
	GameCode   string  `json:"game_code"`
	TransferNo string  `json:"transfer_no"`
	RecordID   string  `json:"record_id"`
	RoundID    string  `json:"round_id"`
	FlowType   uint
}

type R8Activity struct {
	EventID      string  `json:"event_id"`
	AwardID      string  `json:"award_id"`
	ActivityType string  `json:"activity_type"`
	Account      string  `json:"account"`
	Action       string  `json:"action"`
	Currency     string  `json:"currency"`
	Money        float64 `json:"money"` //如果交易货币为VND，比值为 1000:1，如果派奖金额 money = 5，则平台方需对该玩家增加 5000 VND。
}
