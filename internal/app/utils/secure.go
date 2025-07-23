package utils

import (
	"crypto/des"
	"encoding/base64"
	"fmt"
)

// DES 加密
func DESEncrypt(input string, key string) (string, error) {
	// 确保密钥是 8 字节
	if len(key) != 8 {
		return "", fmt.Errorf("key must be 8 bytes long")
	}

	// 创建 DES 加密器
	block, err := des.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	// 使用 ECB 模式加密
	// 因为 ECB 模式没有初始化向量（IV），所以是简单的加密
	// 填充数据（这里使用零填充）
	data := []byte(input)
	padding := 8 - len(data)%8
	padData := append(data, make([]byte, padding)...)

	encrypted := make([]byte, len(padData))
	for i := 0; i < len(padData); i += 8 {
		block.Encrypt(encrypted[i:i+8], padData[i:i+8])
	}

	// 返回 Base64 编码的加密结果
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DES 解密
func DESDecrypt(input string, key string) (string, error) {
	// 确保密钥是 8 字节
	if len(key) != 8 {
		return "", fmt.Errorf("key must be 8 bytes long")
	}

	// 创建 DES 解密器
	block, err := des.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	// 解码 Base64 编码的输入
	encryptedData, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return "", err
	}

	// 使用 ECB 模式解密
	decrypted := make([]byte, len(encryptedData))
	for i := 0; i < len(encryptedData); i += 8 {
		block.Decrypt(decrypted[i:i+8], encryptedData[i:i+8])
	}

	// 去除填充
	padding := decrypted[len(decrypted)-1]
	decrypted = decrypted[:len(decrypted)-int(padding)]

	// 返回解密后的字符串
	return string(decrypted), nil
}

// func main() {
// 	// 测试字符串和密钥
// 	originalText := "Hello, World!"
// 	key := "12345678" // DES 密钥必须是 8 字节

// 	// 打印原始文本
// 	fmt.Println("Original Text:", originalText)

// 	// 加密
// 	encryptedText, err := DESEncrypt(originalText, key)
// 	if err != nil {
// 		log.Fatal("Encryption error:", err)
// 	}
// 	fmt.Println("Encrypted Text:", encryptedText)

// 	// 解密
// 	decryptedText, err := DESDecrypt(encryptedText, key)
// 	if err != nil {
// 		log.Fatal("Decryption error:", err)
// 	}
// 	fmt.Println("Decrypted Text:", decryptedText)

// 	// 检查解密是否正确
// 	if originalText == decryptedText {
// 		fmt.Println("Encryption and decryption worked successfully!")
// 	} else {
// 		fmt.Println("Encryption and decryption failed.")
// 	}
// }
