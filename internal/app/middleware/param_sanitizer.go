package middleware

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	scriptTag  = regexp.MustCompile(`(?i)<script.*?>.*?</script>`)
	sqlKeyword = regexp.MustCompile(`(?i)\b(select|insert|delete|drop|update|truncate|exec|union)\b`)
)

// ParamSanitizerMiddleware 校验并过滤请求中的危险参数
func ParamSanitizerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != "POST" && c.Request.Method != "PUT" && c.Request.Method != "PATCH" {
			c.Next()
			return
		}

		// 提取表单和 JSON 数据
		var input map[string]interface{}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "参数解析失败，请检查请求格式",
			})
			return
		}

		for key, val := range input {
			strVal, ok := val.(string)
			if !ok {
				continue
			}
			if scriptTag.MatchString(strVal) || sqlKeyword.MatchString(strVal) || hasIllegalChars(strVal) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error":   "包含非法参数或危险内容",
					"field":   key,
					"message": "参数被拒绝，请重新输入",
				})
				return
			}
			// 可选规则：禁止空字符串
			if strings.TrimSpace(strVal) == "" {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error":   "字段不能为空",
					"field":   key,
					"message": "请填写必要内容",
				})
				return
			}
		}

		// 请求通过，继续处理
		c.Next()
	}
}

func hasIllegalChars(input string) bool {
	// 可自定义：过滤特殊符号，或使用更强的 regex
	illegals := []string{";", "\"", "'", "--", "\\"}
	for _, c := range illegals {
		if strings.Contains(input, c) {
			return true
		}
	}
	return false
}
