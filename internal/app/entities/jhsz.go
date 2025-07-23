package entities

import (
	"fmt"
)

type JhszGameLoginReq struct {
	GameCode string `json:"gameCode"`
	RoomId   string `json:"roomId"`
	HomeLink string `json:"HomeLink"`
	Language string `json:"Language"`
	UID      uint
	Ip       string
	DeviceId string
	DeviceOS string
}
type JhszGameLoginResp struct {
	GameLink string `json:"game_link"`
}

type ChainLaunchReq struct {
	Account      string `json:"Account"`
	Name         string `json:"Name"`
	GameCode     string `json:"GameCode"`
	Language     string `json:"Language"`
	HomeLink     string `json:"HomeLink"`
	MerchantCode string `json:"MerchantCode"`
	Ip           string `json:"Ip"`
	DeviceId     string `json:"DeviceId"`
	DeviceOS     string `json:"DeviceOS"`
	Sex          string `json:"Sex"`
	RoomId       string `json:"RoomId"`
	Nonce        string `json:"Nonce"`
	Signature    string `json:"Signature"`
}

func (r *ChainLaunchReq) ToMap() map[string]string {
	data := make(map[string]string)
	data["Account"] = r.Account
	data["Name"] = r.Name
	data["GameCode"] = r.GameCode
	data["Language"] = r.Language
	data["HomeLink"] = r.HomeLink
	data["MerchantCode"] = r.MerchantCode
	data["Ip"] = r.Ip
	data["DeviceId"] = r.DeviceId
	data["DeviceOS"] = r.DeviceOS
	data["Sex"] = r.Sex
	data["RoomId"] = r.RoomId
	data["Nonce"] = r.Nonce
	return data
}

type ChainResp struct {
	Error   int    `json:"Error"`
	Message string `json:"Message"`
}

type ChainLaunchResp struct {
	ChainResp
	GameUrl string `json:"GameUrl"`
}

type JhszTransferOrder struct {
	BaseModel
	UID        uint    `gorm:"index;not null"`             // 用户id，非空
	GameCode   string  `gorm:"type:varchar(32);not null;"` // 游戏代码
	Amount     float64 `gorm:"type:decimal(10,2)"`         // 金额，十进制
	TransferNo string  `json:"transfer_no"`                // 设置为唯一索引
	Action     string  `json:"action"`
	UniqueID   string  `json:"unique_id"`
}

type UseFreeCardReq struct {
	Timestamp    int    `json:"timestamp"`
	MerchantCode string `json:"merchant_code"`
	Account      string `json:"account"`
	UniqueID     string `json:"unique_id"`
	Sign         string `json:"sign"`
	Id           uint   `json:"id"`
}

func (req *UseFreeCardReq) GetSignMap() map[string]string {
	return map[string]string{
		"timestamp": fmt.Sprintf("%d", req.Timestamp),
		"uniqueId":  req.UniqueID,
		"account":   req.Account,
		"id":        fmt.Sprintf("%d", req.Id),
	}
}

type GetAvailableFreeCardReq struct {
	Timestamp    int    `json:"timestamp"`
	MerchantCode string `json:"merchant_code"`
	Account      string `json:"account"`
	Sign         string `json:"sign"`
}

func (req *GetAvailableFreeCardReq) GetSignMap() map[string]string {
	return map[string]string{
		"account": req.Account,
	}
}

type JhszTransferReq struct {
	Timestamp    int64   `json:"timestamp"`
	MerchantCode string  `json:"merchantCode"`
	UniqueID     string  `json:"uniqueId"`
	Sign         string  `json:"sign"`
	Account      string  `json:"account"`
	GameCode     string  `json:"gameCode"`
	RecordId     string  `json:"recordId"`
	RoundId      string  `json:"roundId"`
	Amount       float64 `json:"amount"`
	Currency     string  `json:"currency"`
	Action       string  `json:"action"`

	FlowType uint
}

func (req *JhszTransferReq) GetSignMap() map[string]string {
	return map[string]string{
		"timestamp": fmt.Sprintf("%d", req.Timestamp),
		"uniqueId":  req.UniqueID,
		"account":   req.Account,
		"gameCode":  req.GameCode,
		"recordId":  req.RecordId,
		"roundId":   req.RoundId,
		"amount":    fmt.Sprintf("%.2f", req.Amount),
		"currency":  req.Currency,
		"action":    req.Action,
	}
}

type JhszTransferResp struct {
	TransferAmount float64 `json:"TransferAmount"`
	Balance        float64 `json:"Balance"`
}

type JhszBalanceReq struct {
	MerchantCode string `json:"merchant_code"`
	Sign         string `json:"sign"`
	Account      string `json:"account"`
	Currency     string `json:"currency"`
}

type JhszBalanceResp struct {
	Currency string  `json:"Currency"`
	Balance  float64 `json:"Balance"`
}

func (req *JhszBalanceReq) GetSignMap() map[string]string {
	return map[string]string{
		"merchant_code": req.MerchantCode,
		"account":       req.Account,
		"currency":      req.Currency,
	}
}

type JhszNotificationReq struct {
	Account string `json:"account"`
	Message string `json:"message"` // 消息内容
	Title   string `json:"title"`   // 消息标题
}

type PlatLoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Sign     string `json:"sign"`
	LoginIP  string `json:"-"`
}

func (req *PlatLoginReq) GetSignMap() map[string]string {
	return map[string]string{
		"username": req.Username,
		"password": req.Password,
	}
}

type PlatLoginResp struct {
	Account      string `json:"Account"`
	Name         string `json:"Name"`
	MerchantCode string `json:"MerchantCode"`
	Sex          string `json:"Sex"`
}

// type ChainLaunchReq struct {
// 	Account      string `json:"Account"`
// 	Name         string `json:"Name"`
// 	GameCode     string `json:"GameCode"`
// 	Language     string `json:"Language"`
// 	HomeLink     string `json:"HomeLink"`
// 	MerchantCode string `json:"MerchantCode"`
// 	Ip           string `json:"Ip"`
// 	DeviceId     string `json:"DeviceId"`
// 	DeviceOS     string `json:"DeviceOS"`
// 	Sex          string `json:"Sex"`
// 	RoomId       string `json:"RoomId"`
// 	Nonce        string `json:"Nonce"`
// 	Signature    string `json:"Signature"`
// }

// type JhszBetRecord struct {
// 	GameCode  string  `json:"game_code" gorm:"type:varchar(32);not null;"` // 游戏代码
// 	RecordID  string  `json:"record_id" gorm:"type:varchar(36)"`           // 局
// 	Account   string  `json:"account" gorm:"type:varchar(10)"`             // 账号
// 	BetAmount float64 `json:"bet_amount" gorm:"type:decimal(10,2)"`        // 投注金额
// 	CreatedAt int64   `json:"created_at" `                                 // 创建时间
// }
