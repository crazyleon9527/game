package entities

type WalletAddress struct {
	BaseModel
	UID            uint   `json:"uid"`          // 关联的用户ID
	Address        string `json:"address"`      // 区块链钱包地址
	BlockchainType string `json:"chain_type"`   // 地址类型（如 TRC20, ERC20）
	TokenSymbol    string `json:"token_symbol"` // 代币类型（如 USDT, ETH）
	Active         bool   `json:"active"`       // 是否激活（例如：是否允许交易）
	// BlockchainTokenID uint   `json:"blockchain_token_id"` // 关联的 BlockchainToken ID
}

type BlockchainToken struct {
	BaseModel
	BlockchainType   string `json:"blockchain_type"`                 // 区块链类型（如 Ethereum, TRON 等）
	TokenName        string `json:"token_name"`                      // 代币名称（如 Tether）
	TokenSymbol      string `json:"token_symbol" binding:"required"` // 代币类型（如 USDT、ETH）
	ContractAddress  string `json:"contract_address"`                // 代币合约地址（如 ERC-20 或 TRC-20 地址）
	Description      string `json:"description"`                     // 代币描述（例如：USDT 在 TRON 测试网上的实现）
	Active           bool   `json:"active"`                          // 是否激活（例如：是否允许交易）
	TransactionSpeed string `json:"transaction_speed"`               // 交易速度（例如：约 3-5 秒）
	Fee              string `json:"fee"`                             // 手续费（例如：每笔交易约 0.1 TRX）
}

type CreateWalletAddressReq struct {
	BlockchainType string `json:"chain_type" binding:"required"`   // 区块链网络类型（如 Ethereum、TRON 等）
	TokenSymbol    string `json:"token_symbol" binding:"required"` // 代币类型（如 USDT、ETH）
	Address        string `json:"address" binding:"required"`      // 用户钱包地址
}

type GetBlockchainTokenReq struct {
	TokenSymbol string `json:"token_symbol"` // 代币类型
	ChainType   string `json:"chain_type"`   // 区块链类型
}

type GetWalletAddressReq struct {
	UID         uint   `json:"-"`            // 用户 ID
	TokenSymbol string `json:"token_symbol"` // 代币类型
	ChainType   string `json:"chain_type"`   // 区块链类型
}

// BlockchainToken 表示区块链上某个代币的详细信息

// Ethereum:USDT, TRON:USDT, Ethereum:ETH

// type PlatMerchant struct {
// 	BaseModel
// 	MerchantName string `json:"merchant_name"`
// 	MerchantCode string `json:"merchant_code"`
// 	Secret       string `json:"secret"`
// 	SignSecret   string `json:"sign_secret"`
// 	TransferUrl  string `json:"transfer_url"`
// 	BalanceUrl   string `json:"balance_url"`
// 	ProtectedUrl string `json:"protected_url"`
// 	CompeteID    int64  `json:"compete_id" gorm:"default:0"`
// 	PromoteID    int64  `json:"promote_id" gorm:"default:0"`
// }

// type ChainTokenReq struct {
// 	MerchantCode string `json:"merchant_code" binding:"required"`
// 	SecureKey    string `json:"secure_key" binding:"required"`
// 	Sign         string `json:"sign" binding:"required"`
// }

// type ChainTokenResp struct {
// 	AuthToken string `json:"auth_token"`
// 	Timeout   int64  `json:"timeout"`
// }

// type ChainLaunchGameReq struct {
// 	AuthToken    string `json:"auth_token"`
// 	MerchantCode string `json:"merchant_code"`
// 	Language     string `json:"language"`
// 	Username     string `json:"username"`
// 	GameCode     string `json:"game_code"`
// 	RoomId       string `json:"room_id"`
// 	HomeLink     string `json:"home_link"`
// 	Sign         string `json:"sign"`
// }

// func (req *ChainLaunchGameReq) GetSignMap() map[string]string {
// 	return map[string]string{
// 		"merchant_code": req.MerchantCode,
// 		"language":      req.Language,
// 		"auth_token":    req.AuthToken,
// 		"username":      req.Username,
// 		"game_code":     req.GameCode,
// 		"home_link":     req.HomeLink,
// 	}
// }

// type ChainLaunchGameResp struct {
// 	GameLink string `json:"game_link"`
// }

// type ChainTransaction struct {
// 	BaseModel
// 	Status       string  `json:"status" gorm:"index"`
// 	Amount       float64 `json:"amount"`
// 	Currency     string  `json:"currency"`
// 	GameCode     string  `json:"game_code" gorm:"index"`
// 	RecordID     string  `json:"record_id"` //
// 	RoundID      string  `json:"round_id"`
// 	TransferNo   string  `json:"transfer_no" gorm:"uniqueIndex"` // 设置为唯一索引
// 	Action       string  `json:"action"`
// 	MerchantCode string  `json:"merchant_code"`         //
// 	Username     string  `json:"username" gorm:"index"` // 用户名，添加索引
// }

// type ChainAuthToken struct {
// 	Token string `json:"token"`
// }

// type ChainWithdrawResp struct {
// 	Balance float64 `json:"balance"`
// }
