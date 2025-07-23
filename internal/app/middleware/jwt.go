package middleware

import (
	"net/http"
	"rk-api/internal/app/config"
	"rk-api/internal/app/constant"
	"rk-api/internal/app/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// func JWTMiddleware() gin.HandlerFunc {

// 	return func(c *gin.Context) {
// 		user := c.GetHeader("X-User")
// 		if user == config.Get().ServiceSettings.JwtSignKey { //如果是授权者直接验证通过
// 			c.Set("user", user) //记录操作者ID
// 			c.Next()
// 			return
// 		}

// 		authHeader := c.GetHeader("Authorization")
// 		if authHeader == "" {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid Authorization header"})
// 			c.Abort()
// 			return
// 		}

// 		tokenString := authHeader[7:] // Remove "Bearer " prefix
// 		signingKey := []byte(config.Get().ServiceSettings.JwtSignKey)
// 		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 			// Verify the signing method
// 			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
// 			}
// 			return signingKey, nil
// 		})

// 		if err != nil {
// 			// Check for expired JWT
// 			if ve, ok := err.(*jwt.ValidationError); ok {
// 				if ve.Errors&jwt.ValidationErrorExpired != 0 {
// 					c.JSON(http.StatusUnauthorized, gin.H{"error": "JWT token has expired"})
// 				} else {
// 					c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid JWT token"})
// 				}
// 			} else {
// 				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid JWT token"})
// 			}
// 			c.Abort()
// 			return
// 		}

// 		// Extract userID from JWT claims
// 		if claims, ok := token.Claims.(jwt.MapClaims); ok {
// 			c.Set("userID", claims["userID"])
// 			c.Next()
// 		}

// 		c.Next()
// 	}
// }

// AuthMiddleware 处理 access_token 和 refresh_token 的中间件
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//从链接参数获取token
		accessToken := c.Query("token")
		//解析token
		// 从请求头中获取 access_token
		if accessToken == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
				return
			}
			// 去掉 "Bearer " 前缀
			accessToken = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// 解析 access_token
		claims, err := utils.ParseJWT(accessToken, config.Get().ServiceSettings.JwtSignKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid access token"})
			return
		}

		// 检查 access_token 是否过期
		exp, ok := claims["exp"].(float64)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid exp claim"})
			return
		}
		expirationTime := time.Unix(int64(exp), 0)
		if time.Now().After(expirationTime) {
			// 如果 access_token 过期，尝试使用 refresh_token 刷新
			var refreshToken string

			// 1. 从 Cookie 中获取
			refreshToken, err = c.Cookie(constant.REFRESH_TOKEN)
			if err != nil {
				// 2. 从 Header 中获取
				refreshToken = c.GetHeader(constant.REFRESH_TOKEN)
			}

			// 如果仍然没有获取到 refresh_token
			if refreshToken == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "Refresh token is required. Please enable cookies or provide a refresh token in the header.",
				})
				return
			}

			// 验证 refresh_token 是否有效
			refreshClaims, err := utils.ParseJWT(refreshToken, config.Get().ServiceSettings.JwtSignKey)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
				return
			}

			// 检查 refresh_token 是否过期
			refreshExp, ok := refreshClaims["exp"].(float64)
			if !ok {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid exp claim in refresh token"})
				return
			}
			refreshExpirationTime := time.Unix(int64(refreshExp), 0)
			if time.Now().After(refreshExpirationTime) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Expired refresh token"})
				return
			}

			// 检查 refresh_token 是否被踢下线
			userID := refreshClaims["userID"].(string)
			// 生成新的 access_token
			newAccessToken, err := utils.GenerateJWT(config.Get().ServiceSettings.JwtSignKey, time.Hour, userID)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new access token"})
				return
			}

			// 将新的 access_token 返回给客户端
			c.Header("Authorization", "Bearer "+newAccessToken)
		}

		// 检查 access_token 是否被踢下线
		userID := claims["userID"].(string)
		// 将用户信息存储到上下文中
		c.Set("userID", userID)
		// 继续处理请求
		c.Next()
	}
}

// accessToken

// refreshToken
