package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"log"
	"net/url"
	"sort"
	"strings"
)

func GenerateSign(data map[string]string, privateKey string) string {
	// 对map的键进行排序
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建待签名的字符串
	var queryStringBuilder strings.Builder
	for _, k := range keys {
		val := data[k]
		if val != "" {
			// 对 value 进行 URL 编码
			escapedValue := url.QueryEscape(val)
			queryStringBuilder.WriteString(k)
			queryStringBuilder.WriteString("=")
			queryStringBuilder.WriteString(escapedValue)
			queryStringBuilder.WriteString("&")
		}
	}

	// 去除最后一个"&"
	queryString := strings.TrimRight(queryStringBuilder.String(), "&")

	log.Printf("queryString: %s", queryString)

	h := hmac.New(sha256.New, []byte(privateKey))
	h.Write([]byte(queryString))
	return fmt.Sprintf("%x", h.Sum(nil))
}
