package entities

type PlatSetting struct {
	ID              int     `json:"id"`
	Name            string  `json:"name"`
	Price           int     `json:"price"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
	USDExchangeRate float64 `json:"usd_exchange_rate"` // 美元汇率
	MinWithdrawal   float64 `json:"min_withdrawal"`    // 最小提现额
}
