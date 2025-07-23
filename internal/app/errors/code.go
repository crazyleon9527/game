package errors

const (
	SUCCESS     = 200 // 成功
	ERROR       = 500 // 基本错误
	ServerError = 501 // 服务器错误

	ServiceNotFound  = 601 //服务不存在
	ResourceNotExist = 602 // 资源不存在

	VerifyForbidden      = 700 //验证失败
	PermissionsForbidden = 701 //权限不够
	UnLogin              = 702 // 用户未登录
	InvalidParam         = 703 // 参数错误
	ValidationParamError = 704 // ginx验证参数错误
	UnsupportThirdLogin  = 705 // 不支持的三方登录

	// 用户错误

	UserNameNotExist = 10030001 // 用户名不存在
	UserNameExist    = 10030002 // 用户名已经存在 不能重复
	EmailNotExist    = 10030003 // 邮箱不存在
	EmailExist       = 10030004 // 邮箱已经存在
	MobileNotExist   = 10030005 // 手机号不存在
	MobileExist      = 10030006 // 手机号已经存在
	MobileNotBind    = 10030007 //手机号未绑定
	InvalidMobile    = 10030008 // 参数错误
	InvalidUsername  = 10030009 // 参数错误

	SMSVerificationDisabled = 10030010 // 短信验未开启

	AccountNotExist = 10020002 // 用户不存在
	AccountBlocked  = 10020009 // 用户被封禁
	InvalidPassword = 10020013 // 密码错误

	AccountLoginExpire = 10020014 // 登录已过期

	WalletNotExist      = 10020015 // 钱包不存在不存在
	WalletPasswordExist = 10020016 // 钱包密码已经存在

	ChatChannelNotSubscribed = 10020017 // 聊天频道未订阅
	ChatChannelNotExist      = 10020018 // 聊天频道不存在

	InvalidInviteCode = 10020019 //邀请码错误

	InvalidPromotionCode = 10020020 //分销码错误
	InsufficientBalance  = 10020021 //金额不足
	InviteRelationExist  = 10020022 //邀请关系已经存在

	DuplicatePassword = 10020023 // 重复密码

	RetryFrequenceLimit  = 10020115 //email 请求验证码频率太高
	RetryCountLimit      = 10020116 //email 请求验证码频率太高
	VerifiedCodeExpire   = 10020117 //验证码已过期
	VerifiedCodeNotMatch = 10020118 //验证码不一致

	RedEnvelopeNotExist     = 10020301 //红包信息不存在
	RedEnvelopeExpire       = 10020302 //红包信息过期
	InsufficientRedEnvelope = 10020303 //红包数量不足
	RedEnvelopeRepeatGet    = 10020304 //红包重复领取

	PinduoNotExist       = 10020401 //拼多多不存在
	PinduoHasGet         = 10020402 //拼多多已经领了
	PinduoNoGetCondition = 10020403 //拼多多无领取条件
	PinduoExpire         = 10020404 //拼多多已经过期

	MinWithdrawalCashLimit         = 10040023 //最低提现金额限制
	WithdrawCardNotExist           = 10040034 //提现卡不存在
	WithdrawalIntervalLimit        = 10040035 //已经有一笔提现订单
	WithdrawalOrderExists          = 10040036 //已经有一笔提现订单
	WithdralCardExist              = 10040037 //提现卡已经存在
	WithdralCardAccountNumberExist = 10040038 //提现卡账号已经存在
	WithdralCardIFSCExist          = 10040039 //提现卡IFSC已经存在
	WithdrawalDayCountLimit        = 10040040 //提现每日限制

	InvalidWithdrawalReview      = 10040050 //不可用的订单审核状态
	InsufficientWithdrawLockCash = 10040051 //提现冻结的资金不足

	MinRechargeCashLimit           = 10060027 //最低充值金额限制
	RechargeConfigNotExist         = 10060028 //支付配置不存在
	RechargeConfigNotAvailable     = 10060029 //支付配置不可用
	RechargeChannelSettingNotExist = 10060030 //支付渠道配置不存在
	InvalidRechargeOrderActType    = 10060041 //充值活动类型不匹配
	InvalidRechargeReturn          = 10060101 //支付配置不可用
	RechargeReturnNotExist         = 10060102 //支付配置不可用

	RoomNotExist                = 10050001 //房间不存在
	GameStrategyNotExist        = 10050002 //游戏策略不存在
	BettingNotAllowed           = 10050022 // 不允许下注
	UserBettingAmountLimit      = 10050023 //用户下注限制
	UserDayBettingTimesLimit    = 10060024 //用户每天下注次数限制
	UserBettingPeriodTimesLimit = 10060025 //用户每期下注次数限制
)

var ErrCodeMsg = map[int]string{
	SUCCESS:     "success",
	ERROR:       "basic-error",
	ServerError: "server-error",

	ServiceNotFound:  "service-not-found",
	ResourceNotExist: "resource-not-exist",

	VerifyForbidden:      "verification-failed",
	PermissionsForbidden: "insufficient-permissions",
	UnLogin:              "user-not-logged-in",
	InvalidParam:         "invalid-parameters",
	AccountLoginExpire:   "account-login-exipre",
	WalletNotExist:       "wallet-not-exist",
	WalletPasswordExist:  "wallet-password-exist",

	ChatChannelNotSubscribed: "chat-channel-not-subscribed",
	ChatChannelNotExist:      "chat-channel-not-exist",

	EmailNotExist:    "email-does-not-exist",
	EmailExist:       "email-already-exists",
	MobileNotExist:   "mobile-number-does-not-exist",
	MobileExist:      "mobile-number-already-exists",
	MobileNotBind:    "mobile-number-not-bound",
	UserNameExist:    "username-already-exists",
	UserNameNotExist: "username-does-not-exist",
	InvalidMobile:    "invalid-mobile-number",
	InvalidUsername:  "invalid-username",

	SMSVerificationDisabled: "sms-verification-disabled",

	AccountNotExist: "account-does-not-exist",
	AccountBlocked:  "account-blocked",
	InvalidPassword: "invalid-password",

	InvalidInviteCode:    "invalid-invite-code",
	InvalidPromotionCode: "invalid-promotion-code",
	InsufficientBalance:  "insufficient-funds",
	InviteRelationExist:  "invite-relation-exist",
	DuplicatePassword:    "duplicate-password",

	RetryFrequenceLimit:  "retry-frequency-too-high",
	RetryCountLimit:      "retry-count-limit",
	VerifiedCodeExpire:   "verification-code-expired",
	VerifiedCodeNotMatch: "verification-code-mismatch",

	RedEnvelopeNotExist:     "red-envelope-does-not-exist",
	RedEnvelopeExpire:       "red-envelope-expired",
	InsufficientRedEnvelope: "insufficient-red-envelope-quantity",
	RedEnvelopeRepeatGet:    "red-envelope-already-claimed",

	PinduoNotExist:       "pinduoduo-does-not-exist",
	PinduoHasGet:         "pinduoduo-already-claimed",
	PinduoNoGetCondition: "pinduoduo-condition-not-met",
	PinduoExpire:         "pinduoduo-expired",

	MinWithdrawalCashLimit:         "minimum-withdrawal-limit",
	WithdrawCardNotExist:           "withdrawal-card-does-not-exist",
	WithdrawalIntervalLimit:        "withdrawal-interval-too-short",
	WithdrawalOrderExists:          "withdrawal-order-already-exists",
	WithdralCardExist:              "withdrawal-card-already-exists",
	WithdralCardAccountNumberExist: "withdrawal-card-account-number-already-exists",
	WithdralCardIFSCExist:          "withdrawal-card-ifsc-already-exists",
	WithdrawalDayCountLimit:        "withdrawal-day-count-limit",

	InvalidWithdrawalReview:      "invalid-withdrawal-review-status",
	InsufficientWithdrawLockCash: "insufficient-withdrawal-locked-funds",

	MinRechargeCashLimit:           "minimum-recharge-limit",
	RechargeConfigNotExist:         "recharge-configuration-does-not-exist",
	RechargeConfigNotAvailable:     "recharge-configuration-not-available",
	RechargeChannelSettingNotExist: "recharge-channel-setting-not-exist",
	InvalidRechargeOrderActType:    "invalid-recharge-order-act-type",
	InvalidRechargeReturn:          "invalid-recharge-return",
	RechargeReturnNotExist:         "recharge-return-does-not-exist",

	RoomNotExist:           "room-does-not-exist",
	BettingNotAllowed:      "betting-not-allowed",
	UserBettingAmountLimit: "user-betting-amount-limit",

	UserDayBettingTimesLimit:    "user-day-betting-times-limit",    //用户每天下注次数限制
	UserBettingPeriodTimesLimit: "user-betting-period-times-limit", //用户每期下注次数限制
}

// var ErrCodeMsg = map[int]string{
// 	SUCCESS:     "Operation successful",                 // 操作成功
// 	ERROR:       "A general error occurred",             // 出现一个普通错误
// 	ServerError: "Server error, please try again later", // 服务器错误，请稍后重试

// 	ServiceNotFound:  "Requested service not found", // 未找到请求的服务
// 	ResourceNotExist: "Resource does not exist",     // 资源不存在

// 	VerifyForbidden:      "Verification failed, check your details",        // 验证失败，请检查你的信息
// 	PermissionsForbidden: "Insufficient permissions for this action",       // 此操作的权限不足
// 	UnLogin:              "You are not logged in, please log in first",     // 您尚未登录，请先登录
// 	InvalidParam:         "Invalid parameters, please check and try again", // 参数无效，请检查后重试

// 	EmailNotExist:    "The email address does not exist",              // 邮箱地址不存在
// 	EmailExist:       "Email address already exists",                  // 邮箱地址已存在
// 	MobileNotExist:   "The mobile number does not exist",              // 手机号不存在
// 	MobileExist:      "Mobile number already exists",                  // 手机号已存在
// 	MobileNotBind:    "Mobile number is not bound to any account",     // 手机号未绑定任何帐户
// 	UserNameExist:    "Username already taken, please choose another", // 用户名已被占用，请选择另一个
// 	UserNameNotExist: "Username does not exist",                       // 用户名不存在

// 	AccountNotExist: "Account does not exist",        // 帐户不存在
// 	AccountBlocked:  "Your account has been blocked", // 您的账户已被封锁
// 	InvalidPassword: "Incorrect password, try again", // 密码错误，请重试

// 	InvalidInviteCode:    "Invalid invite code",                      // 邀请码无效
// 	InvalidPromotionCode: "Invalid promotion code",                   // 推广码无效
// 	InsufficientBalance:  "Insufficient balance for the transaction", // 交易余额不足
// 	InviteRelationExist:  "invite relation exist",                    //邀请关系已经存在

// 	RetryFrequenceLimit:  "You're trying too often, please wait a while before retrying", // 您尝试的频率太高，请稍等一会再试
// 	RetryCountLimit:      "Too many attempts, please try again later",                    // 尝试次数过多，请稍后再试
// 	VerifiedCodeExpire:   "Your verification code has expired, please request a new one", // 您的验证码已过期，请请求新的验证码
// 	VerifiedCodeNotMatch: "Verification code does not match, check and try again",        // 验证码不匹配，请检查后重试

// 	RedEnvelopeNotExist:     "Red envelope does not exist",                // 红包不存在
// 	RedEnvelopeExpire:       "Red envelope has expired",                   // 红包已过期
// 	InsufficientRedEnvelope: "Not enough red envelopes left",              // 剩余的红包不足
// 	RedEnvelopeRepeatGet:    "You have already claimed this red envelope", // 您已经领取过这个红包

// 	PinduoNotExist:       "Pinduoduo offer does not exist",                               // 拼多多优惠不存在
// 	PinduoHasGet:         "You have already claimed this Pinduoduo offer",                // 您已领取过此拼多多优惠
// 	PinduoNoGetCondition: "You do not meet the conditions to claim this Pinduoduo offer", // 您不满足领取此拼多多优惠的条件
// 	PinduoExpire:         "This Pinduoduo offer has expired",                             // 此拼多多优惠已过期

// 	MinWithdrawalCashLimit:  "Amount is below the minimum withdrawal limit",     // 金额低于最低提款限制
// 	WithdrawCardNotExist:    "Withdrawal card does not exist",                   // 提现卡不存在
// 	WithdrawalIntervalLimit: "Please wait before making another withdrawal",     // 请等待一段时间再进行另一次提款
// 	WithdrawalOrderExists:   "You already have a pending withdrawal order",      // 您已有待处理的提款订单
// 	WithdralCardExist:       "This withdrawal card has already been registered", // 此提款卡已注册

// 	InvalidWithdrawalReview:      "Invalid status for withdrawal review",     // 提现审核状态无效
// 	InsufficientWithdrawLockCash: "Insufficient locked funds for withdrawal", // 提现锁定资金不足

// 	MinRechargeCashLimit:       "Amount is below the minimum recharge limit",                        // 金额低于最低充值限制
// 	RechargeConfigNotExist:     "Recharge configuration does not exist",                             // 充值配置不存在
// 	RechargeConfigNotAvailable: "Recharge service is currently unavailable, please try again later", // 充值服务目前不可用，请稍后再试
// 	InvalidRechargeReturn:      "Invalid return from recharge",                                      // 充值返回无效
// 	RechargeReturnNotExist:     "Recharge transaction does not exist",                               // 充值事务不存在

// 	RoomNotExist:      "The room does not exist",          // 房间不存在
// 	BettingNotAllowed: "Betting not allowed at this time", // 目前不允许下注
// }

func GetErrorMsg(ec int) string {
	return ErrCodeMsg[ec]
}

func IsAccountBlocked(err error) bool {
	if appError, ok := err.(*Error); ok { //是账号封禁的错误可以继续  下面的逻辑处理
		if appError.Code == AccountBlocked {
			return true
		}
	}
	return false
}
