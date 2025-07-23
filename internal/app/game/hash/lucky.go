// -----------------------------------------------------------幸运---------------------------------------------------------------------------------------------------------
package hash

import "rk-api/internal/app/entities"

const (
	LuckyResultNotLucky uint8 = iota + 1
	LuckyResultLucky
)

type LuckyStrategy struct {
	BaseStrategy // 可嵌入公共逻辑
}

func (s *LuckyStrategy) ParseResult(hash string) interface{} {
	// 从后往前找到第一个数字
	// 检查哈希值长度是否至少为2
	if len(hash) < 2 {
		return LuckyResultNotLucky
	}

	// 获取哈希值的最后两位字符
	lastTwoChars := hash[len(hash)-2:]

	// 检查第一位是否是数字
	if !isDigit(lastTwoChars[0]) {
		return LuckyResultNotLucky
	}

	// 检查第二位是否是字母
	if !isLetter(lastTwoChars[1]) {
		return LuckyResultNotLucky
	}

	return LuckyResultLucky
}

func (s *LuckyStrategy) CalculatePayout(bet entities.IHashBetRequest, result interface{}) (float64, float64) {

	r, _ := result.(uint8)
	if bet.GetPrediction() == r {
		return bet.GetBetAmount() * 1.95, 0 // 无手续费
	}
	return 0, 0
}

func (s *LuckyStrategy) ValidateBet(bet entities.IHashBetRequest) error {

	return nil
}
