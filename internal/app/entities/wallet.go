package entities

import (
	"rk-api/pkg/logger"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type UserWallet struct {
	BaseModel
	UID           uint    `json:"uid" redis:"uid" gorm:"uniqueIndex"` // 为 UID 字段添加唯一索引
	Cash          float64 `gorm:"type:decimal(18,2);default:0" redis:"cash"`
	Diamond       uint    `gorm:"default:0" redis:"diamand"`                   // 钻石
	Card          uint    `gorm:"default:0" redis:"card"`                      // 卡
	PromoterCode  int     `gorm:"column:pc;index" redis:"pc" json:"pc"`        // 所属推销码
	Password      string  `gorm:"column:password;size:64" json:"-"`            // 密码
	SecurityLevel uint8   `gorm:"column:security_level;" json:"securityLevel"` // 安全等级，值为0、1、2表示不同的安全设置
}

func (u *UserWallet) CalculateDailyInterest(interestRate decimal.Decimal) decimal.Decimal {
	// 设定利息率为8千分之一，使用decimal防止浮点数运算问题
	currentBalance := decimal.NewFromFloat(u.Cash)    // 用户当前余额
	dailyInterest := currentBalance.Mul(interestRate) // 计算日利息

	return dailyInterest
}

func (t *UserWallet) SafeAdjustCash(cash float64) {
	t.Cash = AddPrecise(t.Cash, cash)
	if t.Cash < 0 { //出现异常
		logger.ZError("user balance exception:",
			zap.Uint("uid", t.ID),
			zap.Float64("cash", t.Cash),
		)
	}
}

type FundFreeze struct {
	BaseModel
	RecordID     string  `gorm:"type:varchar(32);"` // 记录ID (唯一标识符)
	UID          uint    `gorm:"index;"`            // 账户ID (用户ID)
	FreezeAmount float64 `gorm:"not null"`          // 冻结金额
	Currency     string  `gorm:"type:varchar(16);"` // 币种
	Reason       string  `gorm:"type:varchar(64);"` // 冻结原因 (限制长度)
	Status       int     // 冻结状态 (例如: 1=已冻结, 2=已解冻)
	FreezeType   int     // 冻结类型
	Remark       string  `gorm:"type:varchar(255);"` // 备注 (限制长度)
}

type ReportFreezeReq struct {
	UID          uint    `json:"uid"`           // 账户ID (用户ID)
	RecordID     string  `json:"record_id"`     // 记录ID (唯一标识符)
	FreezeAmount float64 `json:"freeze_amount"` // 冻结金额
	Reason       string  `json:"reason"`        // 冻结原因 (限制长度)
	FreezeType   int     `json:"freeze_type"`   // 冻结类型
	Remark       string  `json:"remark"`        // 备注 (限制长度)
}

type UnFreezeReq struct {
	UID      uint   `json:"uid"`       // 账户ID (用户ID)
	RecordID string `json:"record_id"` // 记录ID (唯一标识符)
}

type UpdateWalletPasswordReq struct {
	Password        string `json:"password" binding:"required"`                             //旧密码
	NewPassword     string `json:"newPassword" binding:"required,gte=6,lte=30"`             //新密码，这里假设密码长度应在6到30个字符之间
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"` //确认密码，确保和新密码相同
}

type EnableWalletPasswordReq struct {
	Password string `json:"password" binding:"required"` //密码
}
