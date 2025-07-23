package middleware

import (
	"net/http"
	"path/filepath"
	"rk-api/pkg/logger"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// staticFileHandler 处理静态文件请求
func UnityFileHandler(c *gin.Context) {
	path := c.Request.URL.Path

	// 处理压缩扩展名和内容类型
	var (
		encoding    string
		basePath    = path
		contentType string
	)

	// 识别并设置压缩编码类型
	if strings.HasSuffix(path, ".gz") {
		encoding = "gzip"
		basePath = strings.TrimSuffix(path, ".gz")
	} else if strings.HasSuffix(path, ".br") {
		encoding = "br"
		basePath = strings.TrimSuffix(path, ".br")
	}
	logger.ZInfo("UnityFileHandler", zap.String("path", path), zap.String("encoding", encoding), zap.String("basePath", basePath), zap.String("contentType", contentType))

	// 根据基础路径获取内容类型
	contentType = getContentType(basePath)

	// 设置响应头
	if contentType != "" {
		c.Header("Content-Type", contentType)
	}
	if encoding != "" {
		c.Header("Content-Encoding", encoding)
	}
	// 使用标准库文件服务器处理请求
	fs := http.FileServer(http.Dir("wwwroot"))
	fs.ServeHTTP(c.Writer, c.Request)
	// 中止后续处理
	c.Abort()
}

// getContentType 根据文件扩展名返回 Content-Type。
func getContentType(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".wasm":
		return "application/wasm"
	case ".js":
		return "application/javascript"
	case ".data":
		return "application/octet-stream"
	case ".br":
		return "application/octet-stream" // 匹配 C# 代码行为
	default:
		return ""
	}
}
