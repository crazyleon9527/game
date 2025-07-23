package entities

import "time"

type GetGameRecordListReq struct {
	Paginator
	Category string `json:"category"` // 游戏类别
	UID      uint   `json:"-"`        // 用户ID
}

type GetCategoryStatsListReq struct {
	Paginator
	Category string `json:"category"` // 游戏类别
	UID      uint   `json:"-"`        // 用户ID
}

type GetGamerDailyStatsListReq struct {
	Paginator
	Currency string `json:"currency"`  // 货币类型
	DateType string `json:"date_type"` // 日期类型  (1d,7d,30d)
	UID      uint   `json:"-"`         // 用户ID
}

type GetGamerDailyStatsListRsp struct {
	Page     int         `json:"page" form:"page,default=1"`
	PageSize int         `json:"pageSize" form:"pageSize,default=30"`
	Count    int64       `json:"count" `
	List     interface{} `json:"list" `
	Stat     *GameStats  `json:"stat"`
}

// 分类游戏统计
type GameCategoryStats struct {
	Category  string  `json:"category"`   // 游戏类别（公司棋牌、区块链、外接游戏等）
	BetCount  int     `json:"bet_count"`  // 投注单数
	Profit    float64 `json:"profit"`     // 盈亏
	BetAmount float64 `json:"bet_amount"` // 投注额
}

// 游戏统计
type GameStats struct {
	TotalBetCount  int     `json:"total_bet_count"`  // 总投注单数
	TotalBetAmount float64 `json:"total_bet_amount"` // 总投注额
	TotalProfit    float64 `json:"total_profit"`     // 总输赢
}

// 第一步：修正模型定义，添加唯一索引
type GamerDailyStats struct {
	Source string    `gorm:"uniqueIndex:idx_date_uid_source;size:50" json:"source"`    //第三方标识
	Date   time.Time `gorm:"uniqueIndex:idx_date_uid_source" json:"date"`              // 日期
	UID    uint      `gorm:"uniqueIndex:idx_date_uid_source;index:idx_uid" json:"uid"` // 用户ID
	// 其他字段保持不变...
	BetCount  int     `json:"bet_count"`               // 投注单数
	BetAmount float64 `json:"bet_amount"`              // 投注额
	Profit    float64 `json:"profit"`                  // 盈亏
	Currency  string  `gorm:"size:10" json:"currency"` // 货币类型
}

// // 游戏详情记录
type GameRecord struct {
	BaseModel
	Category     string    `gorm:"index:idx_uid_category,priority:2;size:50" json:"category"`  // 分类字段
	RecordId     string    `gorm:"uniqueIndex;size:40" json:"record_id"`                       // 唯一记录ID
	BetTime      time.Time `json:"bet_time"`                                                   // 下注时间
	BetAmount    float64   `gorm:"bet_amount;type:decimal(10,2)" json:"bet_amount"`            // 投注金额
	Amount       float64   `gorm:"amount;type:decimal(10,2)" json:"amount"`                    // 有效流水
	Profit       float64   `gorm:"profit;type:decimal(20,2)" json:"profit"`                    // 盈亏
	Game         string    `gorm:"size:100" json:"game"`                                       // 游戏名
	Status       uint8     `json:"status"`                                                     // 是否0:未结算 1:已结算 2:已返利
	UID          uint      `gorm:"index:idx_uid;index:idx_uid_category,priority:1" json:"uid"` //   // 用户ID
	Currency     string    `gorm:"size:10" json:"currency"`                                    // 货币类型
	PromoterCode int       `gorm:"column:pc;index" redis:"pc" json:"pc"`                       // 所属推销码
}

type GameRecordRefund struct {
	ID           uint    `gorm:"primarykey" redis:"id"  json:"id"`
	Amount       float64 `gorm:"amount;type:decimal(10,2)" json:"amount"`                    // 有效流水
	UID          uint    `gorm:"index:idx_uid;index:idx_uid_category,priority:1" json:"uid"` //   // 用户ID
	PromoterCode int     `gorm:"column:pc;index" redis:"pc" json:"pc"`                       // 所属推销码
}

func (GameRecordRefund) TableName() string {
	return "game_record"
}

type GameProfitStats struct {
	UID         uint    `json:"uid"`          // 用户ID
	Game        string  `json:"game"`         // 游戏名
	TotalProfit float64 `json:"total_profit"` // 游戏盈利总额
}

type GlobalSyncStatus struct {
	Source         string // 第三方标识
	LastSyncTime   int64  // 最后同步结束时间
	LastRecordTime int64  // 最后处理的记录时间（断点续传）
	SyncWindow     int    // 最大同步窗口（防大时间跨度）
}
