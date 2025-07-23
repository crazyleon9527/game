package game

// 颜色索引
const (
	COLOR_GREEN  = 10
	COLOR_VIOLET = 20
	COLOR_RED    = 30
)

const (
	STATE_INIT    = "STATE_INIT"
	STATE_BETTING = "STATE_BETTING"
	STATE_WAITING = "STATE_WAITING"
	STATE_SETTLE  = "STATE_SETTLE"

	EVENT_START        = "EVENT_START"
	EVENT_STOP_BETTING = "EVENT_STOP_BETTING"
	EVENT_SETTLE       = "EVENT_SETTLE"
	EVENT_STOP         = "EVENT_STOP"
)

const (
	StateBetting     = 1
	StateStopBetting = 2
	StateSettle      = 3
)

const (
	PlayerOrderHistoryMax = 20
	PeriodHistoryMax      = 10
)
