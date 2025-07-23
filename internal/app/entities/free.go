package entities

type FreeCard struct {
	BaseModel
	UID    uint    `gorm:"index" json:"uid"` //用户id
	Amount float64 `json:"amount"`           //免单卡的金额 N
	Used   bool    `json:"used"`             //是否已经使用
}
