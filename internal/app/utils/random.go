package utils

import (
	"bytes"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	encoding   = base32.NewEncoding("ybndrfg8ejkmcpqxot1uwisza345h769")
	// randString = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func GetRandomSalt() string {
	return EncodeMD5(NewRandomString(12))
}

// 生成八位数字
func GenerateSerialNumber() string {
	return fmt.Sprintf("1%07v", rand.New(randSource).Int31n(10000000))
}

func NewRandomString(length int) string {
	var b bytes.Buffer
	str := make([]byte, length+8)
	rand.New(randSource).Read(str)
	encoder := base32.NewEncoder(encoding, &b)
	encoder.Write(str)
	encoder.Close()
	b.Truncate(length) // removes the '==' padding
	return b.String()
}

func NewNonce() string {
	// 生成一个16字节的随机数
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	uuid := hex.EncodeToString(b)
	return uuid
}

// func NewId() string {
// 	var b bytes.Buffer
// 	encoder := base32.NewEncoder(encoding, &b)
// 	encoder.Write(uuid.NewRandom())
// 	encoder.Close()
// 	b.Truncate(26) // removes the '==' padding
// 	return b.String()
// }

// 生成64字符长度的十六进制字符串
func GenerateSecureHex() (string, error) {
	bytes := make([]byte, 32) // 32字节=256位，hex编码后为64字符
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("随机数生成失败: %v", err)
	}
	return hex.EncodeToString(bytes), nil
}
