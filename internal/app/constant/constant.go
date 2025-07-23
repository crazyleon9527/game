package constant

const (
	CURRENCY = "CNY"
)

const (
	CURRENCY_CASH = "CASH"
)

const (
	SUCCESS = "SUCCESS"
	FAIL    = "FAIL"
)

const (
	AuthorizationFixKey = "test_api_session_id_rk_api"
)

const (
	HEADER_REQUEST_ID      = "X-Request-ID"
	HEADER_VERSION_ID      = "X-Version-ID"
	HEADER_ETAG_SERVER     = "ETag"
	HEADER_ETAG_CLIENT     = "If-None-Match"
	HEADER_FORWARDED       = "X-Forwarded-For"
	HEADER_REAL_IP         = "X-Real-IP"
	HEADER_FORWARDED_PROTO = "X-Forwarded-Proto"

	HEADER_TOKEN  = "Token"
	HEADER_BEARER = "BEARER"
	HEADER_AUTH   = "Authorization"

	ACCESS_TOKEN         = "X-Access-Token"
	REFRESH_TOKEN        = "X-Refresh-Token"
	COOKIE_TOKEN         = "ID"
	SESSION_COOKIE_TOKEN = "AUTHTOKEN"

	HEADER_REQUESTED_WITH     = "X-Requested-With"
	HEADER_REQUESTED_WITH_XML = "XMLHttpRequest"
	STATUS                    = "status"
	STATUS_OK                 = "OK"

	SESSION_PROP_PLATFORM = "platform"
	SESSION_PROP_OS       = "os"
	SESSION_PROP_BROWSER  = "browser"
)

const (
	DateLayout    = "2006-01-02 15:04:05"
	DateLayout2   = "20060102150405"
	PeriodLayout  = "20060102"
	PeriodLayout2 = "2006-01-02"

	PreciseZero = 0.001 //精确后 0
	PreciseOne  = 1     //精确后 0
	SmallNumber = 0.000001
	BigNumber   = 10000000 //大数 1000W
)

const (
	TOKEN_SIZE                  = 64
	MAX_TOKEN_EXIPRY_TIME       = 60 * 5 // token超时时间 second
	MAX_EMAIL_TOKEN_EXIPRY_TIME = 60 * 5 // token超时时间 second
	MAX_TOKEN_RETRY_EXIPRY_TIME = 60     // token重发间隔 second
	VERIFICATION_CODE_DAY_LIMIT = 10     //验证码每天次数限制
)

const (
	DATE_TYPE_ONE_DAYS    = "1d"
	DATE_TYPE_SEVEN_DAYS  = "7d"  //七天
	DATE_TYPE_THIRTY_DAYS = "30d" //30天
)

// 验证类型
const (
	VERIFICATION_TYPE_EMAIL = 1 // 邮箱验证
	VERIFICATION_TYPE_SMS   = 2 // 短信验证
)

// 验证码业务类型
const (
	CODE_TYPE_REGISTER            = 1 // 用户注册
	CODE_TYPE_LOGIN               = 2 // 用户登录
	CODE_TYPE_RESET_USER_PASSWORD = 3 // 重置用户密码
	CODE_TYPE_BIND_MOBILE         = 4 // 绑定手机号
	CODE_TYPE_BIND_CARD           = 5 // 绑定银行卡
	CODE_TYPE_BIND_EMAIL          = 6 // 绑定邮箱
	CODE_TYPE_BIND_TELEGRAM       = 7 // 绑定telegram
)

// 验证码状态
const (
	CODE_STATUS_UNUSED  = 0 // 未使用
	CODE_STATUS_USED    = 1 // 已使用
	CODE_STATUS_EXPIRED = 2 // 已过期
)

const (
	NOTIFICATION_TYPE_NORMAL   = 0 // 普通消息
	NOTIFICATION_TYPE_TEMPLATE = 1 // 模板消息
)

const (
	AGENT_LEVEL_MAX      = 1 //最大1级代理
	RETURN_TYPE_RECHARGE = 0 //充值返利类型
	RETURN_TYPE_GAME     = 1 //游戏返利类型
)

const (
	REDIS_KEY_USER               = "USER_%d"
	REDIS_KEY_USER_ACCESS_TOKEN  = "USER_ACCESS_TOKEN_%d"
	REDIS_KEY_USER_REFRESH_TOKEN = "USER_REFRESH_TOKEN_%d"
	REDIS_USER_EXPIRE_TIME       = 3600 * 24 * 30 // seconds   1 month
	REDIS_WINGO_PRESET           = "winGo:presetValue:%s"
	REDIS_NINE_PRESET            = "nine:presetValue:%s"
)

// 以前老的 已经废弃// /5(后台添加) 27注册赠送 30（申请提现扣除） 31房间内输赢，35 （红包），37 充值，42 提现（回调 记录），45（提现驳回），50（邀请）,56（返利 记录） 66 （下级首充返利）70 （旧的返利 记录），127（利息）
//
//	//加入流水表，驳回type=45
const (

	//游戏流水类型大于200
	FLOW_TYPE_WINGO        = 201 //wingo 下注
	FLOW_TYPE_WINGO_REWARD = 202 //wingo 中奖
	FLOW_TYPE_NINE         = 211 //九星 下注
	FLOW_TYPE_NINE_REWARD  = 212 //九星 中奖
	//
	// ------------------------------外部链接游戏大于300---------------------------------------------------------
	FLOW_TYPE_R8_WITHDRAW        = 301 //R8 withdraw
	FLOW_TYPE_R8_WITHDRAW_POKDEN = 302 //r8 withdraw
	FLOW_TYPE_R8_DEPOSIT         = 303 // r8 deposit
	FLOW_TYPE_R8_ROLLBACK        = 304 // r8 回滚
	FLOW_TYPE_R8_ACTIVITY_AWARD  = 305 //

	FLOW_TYPE_ZF_BET         = 311 // zf 投注
	FLOW_TYPE_ZF_PAYOUT      = 312 // zf 派彩，奖励
	FLOW_TYPE_ZF_REFUND      = 313 //zf  1: refund 2: payout failed 3: issue cancel; 1:退回 2:派彩失败 3:取消
	FLOW_TYPE_ZF_PAYOUT_FAIL = 314 //
	FLOW_TYPE_ZF_CANCEL      = 315 //

	FLOW_TYPE_JHSZ_WITHDRAW = 321 //jhsz withdraw
	FLOW_TYPE_JHSZ_DEPOSIT  = 322 // jhsz deposit
	FLOW_TYPE_JHSZ_ROLLBACK = 323 // jhsz 回滚

	FLOW_TYPE_JHSZ_FREEZE   = 324 //jhsz freeze
	FLOW_TYPE_JHSZ_UNFREEZE = 325 // jhsz unfreeze

	FlOW_TYPE_SD        = 331 //sd 下注
	FlOW_TYPE_SD_REWARD = 332 //sd 结算奖励

	FLOW_TYPE_CRASH        = 351 //crash 下注
	FlOW_TYPE_CRASH_REWARD = 352 //crash 结算奖励
	FLOW_TYPE_CRASH_CANCEL = 353 //crash 取消

	FLOW_TYPE_MINE        = 361 //mine 下注
	FlOW_TYPE_MINE_REWARD = 362 //mine 结算奖励

	FLOW_TYPE_DICE        = 371 //dice 下注
	FLOW_TYPE_DICE_REWARD = 372 //dice 结算奖励

	FLOW_TYPE_RETURN_CASH             = 3  // 返利
	FLOW_TYPE_RECHARGE_RETURN_CASH    = 8  //充值返利
	FLOW_TYPE_GET_RED_ENVELOPE        = 4  //红包收益
	FLOW_TYPE_RECHARGE_CASH           = 5  //充值
	FLOW_TYPE_RECHARGE_ACT_10000      = 6  //充值10000
	FLOW_TYPE_APPLY_FOR_WITHDRAW_CASH = 9  //提现
	FLOW_TYPE_WITHDRAW_LOCK_CASH      = 12 //提现锁定金额 返回
	FLOW_TYPE_GM_CASH                 = 11 //gm操作给的

	FLOW_TYPE_INTEREST = 7  //利息
	FLOW_TYPE_PINDUO   = 20 //拼多多
)

const (
	FUND_STATUS_FREEZE = 1 // 冻结
	FUND_STATUS_THAW   = 2 // 解冻
)

const (
	FREEZE_TYPE_REPORT   = 1 // 举报冻结
	FREEZE_TYPE_WITHDRAW = 2 // 提现冻结
)

const (
	STATUS_CREATE = 0 // 未结算
	STATUS_SETTLE = 1 // 已经结算
	STATUS_CANCEL = 2 // 已经取消
)

const (
	RETURN_STATE_WAIT_REVIEW = 0 //待审核
	RETURN_STATE_REVIEWED    = 1 //审核通过
	RETURN_STATE_REJECTED    = 2 //驳回
)

const (
	USER_STATE_BLOCKED = 2 //用户被封
	USER_STATE_NORMAL  = 1 //用户态正常

	WITHDRAW_STATE_WAIT_REVIEW = 0 //待审核
	WITHDRAW_STATE_REVIEWED    = 1 //审核通过
	WITHDRAW_STATE_REJECTED    = 2 //驳回
	WITHDRAW_STATE_TRADE_SUCC  = 3 //3打款成功
	WITHDRAW_STATE_TRADE_FAIL  = 4 //4打款失败

	WITHDRAW_REVIEW_OPT_REJECT         = 0 //提现审核拒绝
	WITHDRAW_REVIEW_OPT_APPROVE        = 1 //提现批准
	WITHDRAW_REVIEW_OPT_RESERVE_FAILED = 2 //提现审核 成功 转失败， 提现失败冲正。

	RECHARGE_STATE_SUCC  = 1 //支付成功
	RECHARGE_CASH_CANCEL = 2 //取消支付
	RECHARGE_STATE_FAIL  = 3 //支付失败

	WITHDRAW_DAY_MAX_COUNT = 3 //提现一天最大的次数

	//支付状态 0未支付 1支付成功 2取消 3支付失败 4谷歌已支付但未消耗'

	//状态 0待审核 1通过 2驳回 3打款成功 4打款失败

)
const (
	RECHARGE_ORDER_ACT_TYPE_10000 = 1 //充值 1W 送 2000 的订单类型
)

// /系统操作类型
const (
	SYS_OPTION_TYPE_ADD_USER            = 1  // 添加用户
	SYS_OPTION_TYPE_EDIT_USER_INFO      = 2  // 操作类型  修改用户
	SYS_OPTION_TYPE_WITHDRAWAL_REVIEW   = 79 // 操作类型  提现审核
	SYS_OPTION_TYPE_WITHDRAWAL_CARD_ADD = 7  // 操作类型  提现卡添加
	SYS_OPTION_TYPE_INVITE_RELATION_FIX = 5  //邀请关系修改
)

const (
	RED_GET_TYPE_NORMAL = 1
)

const (
	StateGameBetAreaLimit        = "StateGameBetAreaLimit"        // 游戏下注限制中
	StateMonthBackupAndClean     = "StateMonthBackupAndClean"     // 游戏清理数据
	StateChangePC                = "StateChangePC"                // 处理业务员合并
	StateSMSVerificationDisabled = "StateSMSVerificationDisabled" // 短信发送禁止
)

// 定义游戏状态常量
const (
	ServiceStatusMaintenance int8 = 0 // 维护中
	ServiceStatusNormal      int8 = 1 // 正常
)

const (
	GameCategoryCompanyChess string = "CompanyCasino" // 公司棋牌
	GameCategoryBlockchain   string = "Blockchain"    // 区块链
	GameCategoryExternal     string = "external"      // 外接游戏
	GameCategorySports       string = "sports"        // 体育游戏
	GameCategoryHash         string = "hash"          // 哈希游戏
	GameCategoryCrash        string = "crash"         // crash游戏
	GameCategoryMine         string = "mine"          // 挖矿游戏
	GameCategoryDice         string = "dice"          // 骰子游戏
	// GameCategoryPuzzle       string = "puzzle"        // 益智游戏
	// GameCategoryAction       string = "action"        // 动作游戏
	// GameCategoryAdventure    string = "adventure"     // 冒险游戏
	// GameCategoryStrategy     string = "strategy"      // 策略游戏
	// GameCategorySimulation   string = "simulation"    // 模拟游戏
	// GameCategoryRacing       string = "racing"        // 竞速游戏
)

const (
	GameNameCrash string = "Crash"
	GameNameMine  string = "Mine"
	GameNameDice  string = "Dice"
)

const (
	CurrencyCNY string = "CNY"
)

const (
	TRC20 = "TRC20"
	ERC20 = "ERC20"
	BTC   = "BTC"
)
