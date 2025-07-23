// -----------------------------------------------------------庄闲和---------------------------------------------------------------------------------------------------------
package hash

import "rk-api/internal/app/entities"

const (
	BankerPlayerTieResultBankerWin uint8 = iota + 1
	BankerPlayerTieResultPlayerWin
	BankerPlayerTieResultTie
)

type BankerPlayerTieStrategy struct {
	BaseStrategy // 可嵌入公共逻辑
}

func (s *BankerPlayerTieStrategy) ParseResult(hash string) interface{} {
	// 从后往前找到最后5位字符
	lastFive := findLastFiveChars(hash)

	// 解析庄家和闲家牌面
	dealerCards, playerCards := lastFive[:2], lastFive[3:]

	// 计算点数
	dealerSum := calculateSum(dealerCards)
	playerSum := calculateSum(playerCards)

	// 判断结果
	return judgeBankerPlayerTieResult(dealerSum, playerSum)
}

// 判断庄闲和结果
func judgeBankerPlayerTieResult(dealer, player int) interface{} {
	switch {
	case dealer > player:
		return BankerPlayerTieResultBankerWin
	case player > dealer:
		return BankerPlayerTieResultPlayerWin
	default:
		return BankerPlayerTieResultTie
	}
}

func (s *BankerPlayerTieStrategy) CalculatePayout(bet entities.IHashBetRequest, result interface{}) (float64, float64) {

	r, _ := result.(uint8)
	if bet.GetPrediction() == r {
		switch r {
		case BankerPlayerTieResultBankerWin:
			return bet.GetBetAmount() * 1.95, 0 // 无手续费
		case BankerPlayerTieResultPlayerWin:
			return bet.GetBetAmount() * 1.95, 0 // 无手续费
		case BankerPlayerTieResultTie:
			return bet.GetBetAmount() * 8, 0 // 无手续费
		default:
			return 0, 0
		}
	} else if r == BankerPlayerTieResultTie && (bet.GetPrediction() == BankerPlayerTieResultBankerWin || bet.GetPrediction() == BankerPlayerTieResultPlayerWin) {
		return bet.GetBetAmount() * 0.5, 0 // 和局时退50%
	}
	return 0, 0
}

func (s *BankerPlayerTieStrategy) ValidateBet(bet entities.IHashBetRequest) error {

	return nil
}
