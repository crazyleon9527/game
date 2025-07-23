package entities

import (
	"regexp"
	"rk-api/pkg/math"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ////////////////////////////////////////////////////////////DB table ////////////////////////////////////////////////////////////////////////////////////////

type BetUser struct {
	UID            uint
	Balance        float64 `json:"withdraw_cash"` // 余额 // 可提现金额
	BetAmountLimit float64 `json:"betLimit"`      // 投注金额限制
	BetTimesLimit  int     `json:"timesLimit"`    // 次数限制
	UntilCash      float64 `json:"untilCash"`     // 游戏结算未到期的cash
	UntilTime      int64   `json:"untilTime"`     // 游戏结算到期时间 //到了期 untilCash 就可以领出来
}

// 用户信息
type UserProfile struct {
	*User
	Wallet  *UserWallet       `json:"wallet"`
	Summary *FinancialSummary `json:"summary"`
}

// 用户基础统计
type FinancialSummary struct {
	UID                uint    `gorm:"primarykey"`
	Interest           float64 `gorm:"column:interest;default:0;type:decimal(10,2)" redis:"interest" json:"interest"`          // 利息收益
	Withdraw           float64 `gorm:"type:decimal(18,2);default:0" redis:"withdraw"`                                          // 提现金额
	Recharge           float64 `gorm:"type:decimal(18,2);default:0" redis:"recharge"`                                          // 充值金额
	RedPacket          float64 `gorm:"type:decimal(18,2);default:0" redis:"red"`                                               // 红包金额
	GM                 float64 `gorm:"type:decimal(18,2);default:0" redis:"gm" json:"-"`                                       // GM赠送
	CommissionRecharge float64 `gorm:"column:cs_recharge;default:0;type:decimal(11,3)" redis:"cs_recharge" json:"cs_recharge"` // 代理用户 充值佣金
	CommissionGame     float64 `gorm:"column:cs_game;default:0;type:decimal(11,3)" redis:"cs_game" json:"cs_game"`             // 代理用户 投注游戏佣金
	PromoterCode       int     `gorm:"column:pc;index" redis:"pc" json:"pc"`                                                   // 所属推销码
}

func (t *FinancialSummary) AddRedCash(cash float64) {
	t.RedPacket = AddPrecise(t.RedPacket, cash)
}

func (t *FinancialSummary) AddCommissionRecharge(cash float64) {
	t.CommissionRecharge = AddPrecise(t.CommissionRecharge, cash)
}

func (t *FinancialSummary) AddCommissionGame(cash float64) {
	t.CommissionGame = AddPrecise(t.CommissionGame, cash)
}

func (t *FinancialSummary) AddRechargeAll(cash float64) {
	t.Recharge = AddPrecise(t.Recharge, cash)
}

func (t *FinancialSummary) AddWithdrawAll(cash float64) {
	t.Withdraw = AddPrecise(t.Withdraw, cash)
}

func (t *FinancialSummary) AddGMCash(cash float64) {
	t.GM = AddPrecise(t.GM, cash)
}

func (u *FinancialSummary) AddInterest(dailyInterest decimal.Decimal) {
	// 将日利息加到Interest字段上
	u.Interest = math.MustParsePrecFloat64(decimal.NewFromFloat(u.Interest).Add(dailyInterest).InexactFloat64(), 3)
}

type User struct {
	ID        uint   `gorm:"primarykey" redis:"id"  json:"uid"`
	CreatedAt int64  `json:"-"`
	UpdatedAt int64  `json:"-"`
	Username  string `gorm:"column:username;size:20;uniqueIndex" redis:"username" json:"username"` // 用户名
	Password  string `gorm:"column:password;size:64" json:"-"`                                     // 密码

	Mobile       string `gorm:"column:mobile;default:null;size:20;uniqueIndex" redis:"mobile" json:"mobile"` // 手机号
	Email        string `gorm:"column:email;default:null;size:32;" json:"email"`                             // 邮箱
	Nickname     string `gorm:"column:nickname;size:35" redis:"nickname" json:"nickname"`                    // 昵称
	IP           string `gorm:"column:ip;size:20" json:"-"`                                                  // 注册IP
	LoginIP      string `gorm:"column:login_ip;default:null;size:20" redis:"ip" json:"ip"`                   // 登录IP
	LoginTime    int64  `gorm:"default:null" redis:"loginTime" json:"-"`                                     // 登录时间
	Device       string `gorm:"column:device;size:20" json:"-"`                                              // 设备
	Plat         string `gorm:"column:plat;size:14" json:"-"`                                                // 平台
	Channel      string `gorm:"column:channel;size:20" json:"channel"`                                       // 渠道
	InviteCode   string `gorm:"column:ic;size:10" redis:"ic" json:"ic"`                                      // 邀请码
	PromoterCode int    `gorm:"column:pc;index" redis:"pc" json:"pc"`                                        // 所属推销码
	Gender       uint8  `gorm:"column:gender;default:0" json:"gender"`                                       // 性别
	Avatar       string `gorm:"column:avatar;default:null;size:100" json:"avatar"`                           // 头像

	// BetAmountLimit float64 `gorm:"column:bet_limit;default:0;type:decimal(10,1)" redis:"bet_limit" json:"betLimit"` // 投注金额限制
	// BetTimesLimit  int     `gorm:"column:times_limit;default:0" redis:"times_limit" json:"timesLimit"`              // 次数限制
	// FirstTen uint8 `gorm:"column:first_ten;default:0" redis:"first_ten" json:"first_ten"` // 充值一万第一次奖励

	Color       uint8  `gorm:"column:color;default:0" redis:"color" json:"color"`                     // 标记颜色
	IsRobot     uint8  `gorm:"column:is_robot;default:0" redis:"is_robot" json:"-"`                   // 是否为机器人
	InviteCount uint   `gorm:"column:invite_count;default:0" redis:"invite_count" json:"-"`           // 邀请数量
	Inviter     string `gorm:"column:inviter;size:20;default:null" redis:"inviter" json:"-"`          // 一级邀请人ID
	Promoter    string `gorm:"column:promoter;size:20;default:null" redis:"promoter" json:"-"`        // 分销推广人
	Telegram    string `gorm:"column:telegram;default:null;size:20" redis:"telegram" json:"telegram"` // 电报

	Status uint8 `gorm:"column:status;default:0" redis:"status" json:"status"` // 用户状态
}

func (t *User) BeforeCreate(tx *gorm.DB) (err error) {
	now := time.Now().Unix()
	t.CreatedAt = now
	t.UpdatedAt = now
	return nil
}

func (t *User) BeforeUpdate(tx *gorm.DB) (err error) {
	t.UpdatedAt = time.Now().Unix()
	return nil
}

// func (t *User) CheckInvalid() error {
// 	if t.BetAmountLimit > constant.BigNumber {
// 		return errors.WithCode(errors.InvalidParam)
// 	}
// 	if t.BetTimesLimit > constant.BigNumber {
// 		return errors.WithCode(errors.InvalidParam)
// 	}
// 	return nil
// }

// func (t *User) AddLockCash(cash float64) {
// 	t.LockCash = AddPrecise(t.LockCash, cash)
// }

// func (t *User) AddUntilCash(cash float64) {
// 	t.UntilCash = AddPrecise(t.UntilCash, cash)
// }

// 添加利息

func (t *User) GetUserID() string {
	return strconv.Itoa(int(t.ID))
}

func (t *User) TableName() string {
	return "user"
}

type SysUser struct {
	UID      int    `gorm:"column:uid;"`
	Username string `gorm:"column:user_name;"`
	Secret   string `gorm:"column:secret;"`
}

// 假定存在一个结构体，用来保存每日利息计算的记录
type InterestCalculationLog struct {
	Date         string `gorm:"primary_key"` // 使用日期作为主键
	IsCalculated bool
}

// ////////////////////////////////////////////////////////////DB table ////////////////////////////////////////////////////////////////////////////////////////

// 客服
type Customer struct {
	TelegramTeam     string `json:"telegram_team"`
	TelegramCustomer string `json:"telegram_customer"`
}

type LoginResp struct {
	Token string `json:"token"`
	UID   uint   `json:"uid"`
}

// binding:"required,max=32"

type Distribute struct {
	Plat          string `json:"plat"  example:"" binding:"omitempty"`    //  平台
	Channel       string `json:"channel"  example:"" binding:"omitempty"` //  渠道
	InviteCode    string `json:"ic"  example:"" binding:"omitempty"`      //邀请码
	PromotionCode string `json:"pc"  example:"" binding:"omitempty"`      //分销码
}

func (t *Distribute) ConvertPC() (int, error) {
	if t.PromotionCode == "" {
		return 0, nil
	}
	patterns := regexp.MustCompile(`\d+`) // 第一种
	st := patterns.FindStringSubmatch(t.PromotionCode)
	if len(st) > 0 {
		t.PromotionCode = st[0]
	}
	return strconv.Atoi(t.PromotionCode)
}

type EditNicknameReq struct {
	UID      uint   `json:"-"`
	Nickname string `json:"nickname" binding:"required"`
	Gender   uint8  `json:"gender"`
}
type EditAvatarReq struct {
	UID    uint   `json:"-"`
	Avatar string `json:"avatar" binding:"required"`
}
type BindTelegramReq struct {
	UID      uint   `json:"-"`
	Telegram string `json:"telegram" binding:"required"`
}

type BindEmailReq struct {
	UID   uint   `json:"-"`
	Email string `json:"email" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

type EditUserInfoReq struct {
	UID uint `json:"uid"`

	BalanceAdd float64 `json:"balanceAdd"` // 使用指针允许空值和非空值
	BetLimit   float64 `json:"betLimit"`   // 下注限制，使用指针
	TimesLimit int     `json:"timesLimit"` // 每日下注次数限制，使用指针
	Color      uint8   `json:"color"`      // 颜色标记，使用指针
	Password   *string `json:"password"`   // 密码，使用指针
	Status     uint8   `json:"status"`     // 状态，使用指针
	OptionID   uint    `json:"optionID"`   // 操作者
	IP         string  `json:"ip"`
}

type AddUserReq struct {
	Password string `json:"password"`
	Username string `json:"username"`
	OptionID uint   `json:"optionID"` // 操作者
	Status   uint8  `json:"status"`
	IP       string `json:"ip"`
}

type SearchUserReq struct {
	Username string `json:"username"`
}

type ClearUserCacheReq struct {
	UID uint `json:"uid"`
}

type SwitchChannelReq struct {
	Channel string `json:"channel"`
}
