package middleware

import (
	"rk-api/internal/app/config"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	corsSetting := config.Get().CorsSettings
	return cors.New(cors.Config{
		AllowOrigins:     corsSetting.AllowOrigins,
		AllowMethods:     corsSetting.AllowMethods,
		AllowHeaders:     corsSetting.AllowHeaders,
		AllowCredentials: corsSetting.AllowCredentials,
		MaxAge:           time.Second * time.Duration(corsSetting.MaxAge),
	})
}
