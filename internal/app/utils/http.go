package utils

import (
	"net/http"
	"rk-api/internal/app/constant"
	"strings"

	"github.com/gin-gonic/gin"
)

// 判断是否https
func IsHttps(c *gin.Context) bool {
	if c.GetHeader(constant.HEADER_FORWARDED_PROTO) == "https" || c.Request.TLS != nil {
		return true
	}
	return false
}

func GetProtocol(r *http.Request) string {
	if r.Header.Get(constant.HEADER_FORWARDED_PROTO) == "https" || r.TLS != nil {
		return "https"
	} else {
		return "http"
	}
}

func GetFullURL(r *http.Request) string {
	// 获取协议
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	// 获取主机
	host := r.Host
	// 构建完整 URL
	return scheme + "://" + host
}

func GetDeviceOS(userAgent string) string {
	if strings.Contains(userAgent, "Windows NT") {
		return "Windows"
	} else if strings.Contains(userAgent, "Macintosh") {
		return "macOS"
	} else if strings.Contains(userAgent, "Linux") {
		return "Linux"
	} else if strings.Contains(userAgent, "Android") {
		return "Android"
	} else if strings.Contains(userAgent, "iOS") || strings.Contains(userAgent, "iPhone") || strings.Contains(userAgent, "iPad") {
		return "iOS"
	}
	return "Unknown"
}
