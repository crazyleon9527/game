// -----------------------------------------------------------庄闲牛牛---------------------------------------------------------------------------------------------------------
package hash

import "rk-api/internal/app/entities"

const (
	BullBullResultDealerBull uint8 = iota + 1
	BullBullResultPlayerBull
	BullBullResultDealerNine
	BullBullResultPlayerNine
	BullBullResultDealerWin
	BullBullResultPlayerWin
	BullBullResultDraw
)

type BullBullStrategy struct {
	BaseStrategy // 可嵌入公共逻辑
}

func (s *BullBullStrategy) ParseResult(hash string) interface{} {
	// 从后往前找到最后5位字符
	lastFive := findLastFiveChars(hash)

	// 解析庄家和闲家牌面
	dealerCards, playerCards := lastFive[:3], lastFive[2:]

	// 计算点数
	dealerSum := calculateSum(dealerCards)
	playerSum := calculateSum(playerCards)

	// 判断结果
	return judgeBullBullResult(dealerSum, playerSum)
}

// 判断庄闲牛牛结果
func judgeBullBullResult(dealer, player int) interface{} {
	switch {
	case dealer == 0:
		return BullBullResultDealerBull
	case dealer == 9:
		return BullBullResultDealerNine
	case player == 0:
		return BullBullResultPlayerBull
	case player == 9:
		return BullBullResultPlayerNine
	case dealer > player:
		return BullBullResultDealerWin
	case player > dealer:
		return BullBullResultPlayerWin
	default:
		return BullBullResultDraw
	}
}

func (s *BullBullStrategy) CalculatePayout(bet entities.IHashBetRequest, result interface{}) (float64, float64) {

	bbResult := result.(uint8)
	switch bbResult {
	case BullBullResultDealerBull:
		if bet.GetPrediction() == BullBullResultDealerBull {
			return bet.GetBetAmount() * 10 * 0.9, 0 // 牛牛抽10%手续费
		}
		return 0, 0
	case BullBullResultPlayerBull:
		if bet.GetPrediction() == BullBullResultPlayerBull {
			return bet.GetBetAmount() * 10 * 0.9, 0 // 牛牛抽10%手续费
		}
		return 0, 0
	case BullBullResultDealerNine:
		if bet.GetPrediction() == BullBullResultDealerNine {
			return bet.GetBetAmount() * 9 * 0.9, 0 // 牛九抽10%手续费
		}
		return 0, 0
	case BullBullResultPlayerNine:
		if bet.GetPrediction() == BullBullResultPlayerNine {
			return bet.GetBetAmount() * 9 * 0.9, 0 // 牛九抽10%手续费
		}
		return 0, 0
	case BullBullResultDealerWin:
		if bet.GetPrediction() == BullBullResultDealerWin {
			return bet.GetBetAmount() * 1.95, 0
		}
		return 0, 0
	case BullBullResultPlayerWin:
		if bet.GetPrediction() == BullBullResultPlayerWin {
			return bet.GetBetAmount() * 1.95, 0
		}
		return 0, 0
	case BullBullResultDraw:
		if bet.GetPrediction() == BullBullResultDealerWin || bet.GetPrediction() == BullBullResultPlayerWin {
			return bet.GetBetAmount() * 0.99, 0 // 和局时扣1%手续费退还
		}
		return 0, 0
	default:
		return 0, 0
	}
}

func (s *BullBullStrategy) ValidateBet(bet entities.IHashBetRequest) error {

	return nil
}
