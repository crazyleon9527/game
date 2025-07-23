package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func GenerateJWT(secretKey string, expireTime time.Duration, userID string) (string, error) {
	// 计算过期时间

	// 创建 JWT 令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userID,
		"exp":    time.Now().Add(expireTime).Unix(),
	})

	// 使用密钥签名令牌
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseJWT 解析 JWT 并返回 Claims
func ParseJWT(tokenString, secretKey string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// VerifyJWTExpiration 验证 JWT 是否过期
func VerifyJWTExpiration(tokenString, secretKey string) (bool, error) {
	// 解析 Token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return false, err
	}

	// 提取 Claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false, fmt.Errorf("invalid token claims")
	}

	// 检查 exp 声明
	exp, ok := claims["exp"].(float64)
	if !ok {
		return false, fmt.Errorf("invalid exp claim")
	}

	// 将 exp 转换为时间戳
	expirationTime := time.Unix(int64(exp), 0)

	// 判断是否过期
	if time.Now().After(expirationTime) {
		return true, nil // Token 已过期
	}

	return false, nil // Token 未过期
}
