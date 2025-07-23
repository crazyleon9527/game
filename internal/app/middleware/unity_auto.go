package middleware

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// UnityFileHandler 处理静态文件请求，支持根据 Accept-Encoding 返回压缩文件
// UnityAutoFileHandler 中间件
func UnityAutoFileHandler(wwwRoot string) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// 检查是否是目标文件类型
		isTargetFile := false
		switch {
		case strings.HasSuffix(path, ".data"),
			strings.HasSuffix(path, ".wasm"),
			strings.HasSuffix(path, "framework.js"),
			strings.HasSuffix(path, "symbols.json"):
			isTargetFile = true
		}

		// 如果是目标文件且支持压缩
		if isTargetFile {
			acceptEncoding := strings.ToLower(c.GetHeader("Accept-Encoding"))
			fs := http.Dir(wwwRoot) // 使用传入的 wwwroot 路径

			// 按 br -> gzip 优先级检查
			if strings.Contains(acceptEncoding, "br") {
				compressedPath := path + ".br"
				if fileExists(fs, compressedPath) {
					c.Request.URL.Path = compressedPath
					c.Header("Content-Encoding", "br")
				} else if strings.Contains(acceptEncoding, "gzip") {
					compressedPath := path + ".gz"
					if fileExists(fs, compressedPath) {
						c.Request.URL.Path = compressedPath
						c.Header("Content-Encoding", "gzip")
					}
				}
			} else if strings.Contains(acceptEncoding, "gzip") {
				compressedPath := path + ".gz"
				if fileExists(fs, compressedPath) {
					c.Request.URL.Path = compressedPath
					c.Header("Content-Encoding", "gzip")
				}
			}

			contentType := getUnityAutoContentType(c.Request.URL.Path)
			if contentType != "" {
				c.Header("Content-Type", contentType)
			}
		}

		// 使用修改后的路径处理文件请求
		fs := http.FileServer(http.Dir(wwwRoot)) // 使用传入的 wwwroot 路径
		fs.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}

// 检查文件是否存在且是普通文件
func fileExists(fs http.FileSystem, path string) bool {
	f, err := fs.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	stat, err := f.Stat()
	return err == nil && !stat.IsDir()
}

// 获取基础文件类型（自动去除压缩扩展名）
func getUnityAutoContentType(path string) string {
	// 去除压缩扩展名
	basePath := path
	for _, ext := range []string{".br", ".gz"} {
		if strings.HasSuffix(basePath, ext) {
			basePath = strings.TrimSuffix(basePath, ext)
			break
		}
	}

	// 根据实际文件扩展名判断类型
	switch ext := filepath.Ext(basePath); ext {
	case ".wasm":
		return "application/wasm"
	case ".js":
		return "application/javascript"
	case ".data":
		return "application/octet-stream"
	case ".json":
		return "application/json"
	default:
		return ""
	}
}
