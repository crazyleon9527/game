package hash

import (
	"errors"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/service"
	"rk-api/pkg/logger"

	"go.uber.org/zap"
)

//---------------------------------------- 单双 ---------------------------------------

const (
	SingleDoubleResultOdd uint8 = iota + 1
	SingleDoubleResultEven
)

type SingleDoubleStrategy struct {
	BaseStrategy // 可嵌入公共逻辑
}

func (s *SingleDoubleStrategy) ParseResult(hash string) interface{} {
	// 从后往前找到第一个数字
	var digit, _ = findLastDigit(hash)
	return judgeOddOrEven(digit)
}

func judgeOddOrEven(digit int) interface{} {
	if digit%2 == 0 {
		return SingleDoubleResultEven
	} else {
		return SingleDoubleResultOdd
	}
}

func (s *SingleDoubleStrategy) ValidateBet(bet entities.IHashBetRequest) error {

	if bet.GetBetAmount() <= 0 {
		return errors.New("invalid bet amount")
	}
	if bet.GetPrediction() != SingleDoubleResultOdd && bet.GetPrediction() != SingleDoubleResultEven {
		return errors.New("prediction error")
	}
	return nil
}

func (s *SingleDoubleStrategy) CalculatePayout(bet entities.IHashBetRequest, result interface{}) (float64, float64) {

	r, _ := result.(uint8)
	if bet.GetPrediction() == r {
		return bet.GetBetAmount() * 1.95, 0 // 无手续费
	}
	return 0, 0
}

// ------------------------------------ room ------------------------------------

type SDGameRoom struct {
	*BaseGameRoom
}

// NewSDGameRoom 创建单双游戏房间
func NewSDGameRoom(srv *service.HashGameService, strategy GameStrategy) IGameRoom {
	return &SDGameRoom{
		BaseGameRoom: NewBaseGameRoom(srv, strategy, &SDGameRoom{}),
	}
}

func (g *SDGameRoom) buildHashGameRound(round *entities.BaseHashGameRound) entities.IHashGameRound {
	logger.ZInfo("SDGameRoom buildHashGameRound", zap.Any("round", round))
	return &entities.HashSDGameRound{BaseHashGameRound: round}
}

func (g *SDGameRoom) buildHashGameOrder(order *entities.BaseHashGameOrder) entities.IHashGameOrder {
	logger.ZInfo("SDGameRoom buildHashGameOrder", zap.Any("order", order))
	return &entities.HashSDGameOrder{BaseHashGameOrder: order}
}
