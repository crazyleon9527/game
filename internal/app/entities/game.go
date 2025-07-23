package entities

//-------------------------------------------------------------------------------------------------------------------------------------------------------------

// ---------------------------------------------------------------------------jhsz---------------------------------------------------------------------------------------------------------------
// Game 游戏模型
type Game struct {
	BaseModel
	Name string `gorm:"size:128" json:"name"` // 游戏名

	Category       string `gorm:"size:64" json:"category"`          // 游戏类别（公司棋牌、区块链、外接游戏等），限制为64字符
	GameCode       string `gorm:"size:64" json:"game_code"`         // 游戏代号，限制为64字符
	IconURL        string `gorm:"size:512" json:"icon_url"`         // 游戏图标地址
	CompanyLogoURL string `gorm:"size:512" json:"company_logo_url"` // 游戏公司logo地址
	OnlineCount    int    `json:"online_count"`                     // 当前在线玩家数

	GameplayType  string  `gorm:"size:64" json:"gameplay_type"`    // 游戏玩法类型，限制为64字符
	Popularity    int     `json:"popularity"`                      // 游戏热度
	Rating        float64 `gorm:"type:decimal(3,2)" json:"rating"` // 游戏评分（0到10的评分，保留两位小数）
	Description   string  `gorm:"size:1024" json:"description"`    // 游戏描述，限制为1024个字符
	IsActive      bool    `json:"is_active"`                       // 是否上架
	IsDeleted     bool    `json:"is_deleted"`                      // 是否删除
	ServiceStatus int8    `json:"service_status"`                  // 服务状态
	Language      string  `gorm:"size:8" json:"language"`          // 游戏支持的语言类型
	Priority      int     `json:"priority"`                        // 游戏优先级(较高数字表示较高优先级)
}

type GameIdentification struct {
	Category string `gorm:"size:64" json:"category"` // 游戏类别（公司棋牌、区块链、外接游戏等）
	Name     string `gorm:"size:64" json:"name"`     // 游戏名
}

type GameRefresh struct {
	GameCode    string `gorm:"size:64" json:"game_code"` // 游戏编码
	OnlineCount int    `json:"online_count"`
}

type GetGameListReq struct {
	Paginator
	Category string `json:"category"`
	Status   int8   `json:"status"`
}

type SearchGameReq struct {
	Paginator
	Category string `json:"category"`
	Status   int8   `json:"status"`
	Name     string `json:"name"` // 游戏名
}

//---------------------------------------------------------------------------jhsz---------------------------------------------------------------------------------------------------------------
