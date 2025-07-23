package test

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"testing"
)

func TestMD5Sign(t *testing.T) {
	// 原始签名串
	originalStr := "amount=20000&appId=66a4bfba57b697e04f026c7d&channelOrderNo=2024072721203179893938&clientIp=13.201.103.123&createdAt=1722095432000&currency=INR&customerEmail=default@email.com&customerName=6375667129&customerPhone=6375667129&ifCode=paypay&mchNo=M1722073018&mchOrderNo=p80508281722095431932374&payOrderId=P1817226068204462082&reqTime=1722095624479&state=2&wayCode=PAYY_TTPAY2"
	finalSignature := "BD6AD2F933BD3D178B32D574EFB9F899"

	privateKeyValue := "&key=2vM0BRQvUUBlXPq66lq6gfKOypYRtOPwfyuNsLDdd2DxMyHFf7uiHKGQlD81yGvoNInxaYZ0Q7YEh5x7Wff8oNV9jlRkTptobFNiNM3q4STG88PBm3b8HZT6rT9HJcR2"

	// 不能去掉的key
	mustKeepKeys := []string{"amount"}

	// 解析签名串
	values, _ := url.ParseQuery(originalStr)

	// 获取所有可以去掉的key
	removableKeys := make([]string, 0)
	for key := range values {
		if !contains(mustKeepKeys, key) {
			removableKeys = append(removableKeys, key)
		}
	}

	// 尝试去掉不同的key组合
	for i := 0; i < (1 << uint(len(removableKeys))); i++ {
		newValues := make(url.Values)
		for key, value := range values {
			if contains(mustKeepKeys, key) || (i&(1<<uint(indexOf(removableKeys, key))) == 0) {
				newValues[key] = value
			}
		}

		// 生成新的签名串
		newStr := generateSignatureString(newValues)

		newStr += privateKeyValue
		fmt.Println("签名串+私钥:", newStr)

		// 计算MD5签名
		signature := calculateMD5(newStr)

		if signature == finalSignature {
			fmt.Println("找到匹配的签名串:", newStr)
			return
		}
	}

	fmt.Println("没有找到匹配的签名串")
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func indexOf(slice []string, item string) int {
	for i, s := range slice {
		if s == item {
			return i
		}
	}
	return -1
}

func generateSignatureString(values url.Values) string {
	var sb strings.Builder
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for i, key := range keys {
		if i > 0 {
			sb.WriteString("&")
		}
		sb.WriteString(key)
		sb.WriteString("=")
		sb.WriteString(values.Get(key))
	}
	return sb.String()
}

func calculateMD5(str string) string {
	hash := md5.Sum([]byte(str))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}
