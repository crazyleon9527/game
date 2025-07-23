package entities

import "time"

type UserOauth struct {
	BaseModel
	UID              uint   `gorm:"column:uid"`
	OauthType        uint64 `gorm:"type:bigint unsigned;" json:"oauthType"`
	OauthID          string `gorm:"type:varchar(64);unique;" json:"oauthId"`
	OauthAccessToken string `gorm:"type:longtext;" json:"oauthAccessToken"`
	OauthExpires     string `gorm:"type:longtext;" json:"oauthExpires"`
	Status           uint64 `gorm:"type:bigint unsigned;" json:"status"`
}

func (*UserOauth) TableName() string {
	return "user_oauth"
}

type OauthState struct {
	Plat        string `json:"plat"`
	State       string `json:"state"`
	OauthUrl    string `json:"oauth_url"`
	RedirectUrl string `json:"redirect_url"`
	Extra       string `json:"extra"`
}

// 判断RedirectUrl为空的情况
func (t *OauthState) GetUrl() string {
	url := t.RedirectUrl
	if t.Extra != "" {
		url += "?extra=" + t.Extra
	}
	return url
}

type OAuthLoginReq struct {
	Plat        string `json:"plat" example:"github" binding:"required"`                         //第三方平台(google,facebook,twister,github[可用])
	RedirectUrl string `json:"redirect_url" example:"https://www.baidu.com" binding:"omitempty"` //第三方登录成功后跳转地址
	Extra       string `json:"extra" example:"" binding:"omitempty"`                             //第三方登录成功后跳转地址携带参数
}

type OAuthToken struct {
	AccessToken  string        `json:"access_token,omitempty"`
	RefreshToken string        `json:"refresh_token,omitempty"`
	Scope        string        `json:"scope,omitempty"`
	TokenType    string        `json:"token_type,omitempty"`
	ExpiresIn    time.Duration `json:"expires_in,omitempty"`
	RedirectUrl  string        `json:"redirect_url,omitempty"`
	UID          uint          `json:"uid,omitempty"`
}
