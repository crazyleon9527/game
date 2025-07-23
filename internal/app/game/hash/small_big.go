package hash

import (
	"errors"
	"rk-api/internal/app/entities"
)

// -----------------------------------------------------------大小---------------------------------------------------------------------------------------------------------

const (
	SmallBigResultSmall uint8 = iota + 1
	SmallBigResultBig
)

type SmallBigStrategy struct {
	BaseStrategy // 可嵌入公共逻辑
}

func (s *SmallBigStrategy) ParseResult(hash string) interface{} {
	// 从后往前找到第一个数字
	var digit, _ = findLastDigit(hash)
	return judgeBigOrSmall(digit)
}

func judgeBigOrSmall(digit int) interface{} {
	if digit < 5 {
		return SmallBigResultSmall
	} else {
		return SmallBigResultBig
	}
}

func (s *SmallBigStrategy) CalculatePayout(bet entities.IHashBetRequest, result interface{}) (float64, float64) {

	r, _ := result.(uint8)
	if bet.GetPrediction() == r {
		return bet.GetBetAmount() * 1.95, 0 // 无手续费
	}
	return 0, 0
}

func (s *SmallBigStrategy) ValidateBet(bet entities.IHashBetRequest) error {

	if bet.GetBetAmount() <= 0 {
		return errors.New("invalid bet amount")
	}
	if bet.GetPrediction() != SmallBigResultSmall && bet.GetPrediction() != SmallBigResultBig {
		return errors.New("prediction error")
	}
	return nil
}
