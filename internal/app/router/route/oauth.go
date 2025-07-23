package route

import (
	"github.com/gin-gonic/gin"
	"rk-api/internal/app/api"
)

func RegisterOAuthRoutes(r *gin.RouterGroup, oauthAPI *api.OauthAPI, app *gin.Engine) {
	oauth := r.Group("/oauth")
	{
		oauth.POST("/login", oauthAPI.Login)
		app.GET("/oauth/google/callback", oauthAPI.GoogleCallBack)
		app.GET("/oauth/github/callback", oauthAPI.CallBackGithub)
	}
}
