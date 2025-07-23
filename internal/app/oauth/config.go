package oauth

import (
	"rk-api/internal/app/config"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

var (
	GithubOauthConfig   *oauth2.Config
	GoogleOauthConfig   *oauth2.Config
	FacebookOauthConfig *oauth2.Config
	TwitterOauthConfig  *oauth2.Config

	once sync.Once
)

// InitConfig 初始化 OAuth2 配置
func InitConfig() {
	once.Do(func() {
		var CallbackHost = config.Get().ServiceSettings.OAuth2CallbackHost

		// GitHub OAuth2 配置
		GithubOauthConfig = &oauth2.Config{
			ClientID:     "98d5a2772a8a866aaff1",                     // 替换为你的 GitHub Client ID
			ClientSecret: "a6c91ec5b310f41a44bb187725d5a70e9a9c6bae", // 替换为你的 GitHub Client Secret
			Scopes:       []string{"user"},
			Endpoint:     github.Endpoint,
			RedirectURL:  CallbackHost + "/oauth/github/callback",
		}

		// Google OAuth2 配置
		GoogleOauthConfig = &oauth2.Config{
			ClientID:     "1075149644401-qh9r4i37ftk2sjf7hj63epve8crhu7p6.apps.googleusercontent.com", // 替换为你的 Google Client ID
			ClientSecret: "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",                                       // 替换为你的 Google Client Secret
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
			RedirectURL:  CallbackHost + "/oauth/google/callback",
			Endpoint:     google.Endpoint,
		}

		// Facebook OAuth2 配置
		FacebookOauthConfig = &oauth2.Config{
			ClientID:     "", // 替换为你的 Facebook Client ID
			ClientSecret: "", // 替换为你的 Facebook Client Secret
			Scopes:       []string{"openid"},
			RedirectURL:  CallbackHost + "/oauth/facebook/callback",
			Endpoint:     facebook.Endpoint,
		}

		// Twitter OAuth2 配置
		TwitterOauthConfig = &oauth2.Config{
			ClientID:     "", // 替换为你的 Twitter Client ID
			ClientSecret: "", // 替换为你的 Twitter Client Secret
			Endpoint: oauth2.Endpoint{
				AuthURL:   "https://twitter.com/i/oauth2/authorize",
				TokenURL:  "https://api.twitter.com/2/oauth2/token",
				AuthStyle: oauth2.AuthStyleInHeader,
			},
			RedirectURL: CallbackHost + "/oauth/twitter/callback",
			Scopes:      []string{"tweet.read", "users.read", "tweet.write"},
		}
	})
}

type GoogleUser struct {
	Name               string `json:"name"`
	Email              string `json:"email"`
	EmailVerified      bool   `json:"email_verified"`
	Picture            string `json:"picture"`
	HostedGsuiteDomain string `json:"hd"`
}
