package utils

import (
	"errors"
	"regexp"
	"rk-api/internal/app/constant"
)

func ValidateTRC20Address(address string) bool {
	// TRC20 地址通常是以 "T" 开头，且总长度为 34 个字符
	re := regexp.MustCompile("^T[A-Za-z0-9]{33}$")
	return re.MatchString(address)
}

// 验证 ERC20 地址
func ValidateERC20Address(address string) bool {
	// ERC20 地址通常是 40 个十六进制字符（以 "0x" 开头）
	re := regexp.MustCompile("^0x[A-Fa-f0-9]{40}$")
	return re.MatchString(address)
}

// 验证地址（根据地址类型选择不同的验证方式）
func ValidateAddress(address string, blockchainType string) error {
	switch blockchainType {
	case constant.TRC20:
		if !ValidateTRC20Address(address) {
			return errors.New("invalid TRC20 address")
		}
	case constant.ERC20:
		if !ValidateERC20Address(address) {
			return errors.New("invalid ERC20 address")
		}
	default:
		return errors.New("unknown blockchain type")
	}
	return nil
}
