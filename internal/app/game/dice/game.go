package dice

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/service"
	"rk-api/internal/app/utils"
	"rk-api/pkg/logger"
	"strconv"
	"time"

	"github.com/google/wire"
	"go.uber.org/zap"
)

var DiceGameSet = wire.NewSet(
	NewDiceGame, // 直接提供结构体指针
)

type DiceGame struct {
	Srv     *service.DiceGameService
	setting *DiceSetting
}

type DiceSetting struct {
	// 抽水比例 1%
	Rate uint8

	// 增量值1：客户端种子+“:0:0”
	Add1 string
}

func NewDiceGame(srv *service.DiceGameService) *DiceGame {
	m := &DiceGame{
		Srv: srv,
		setting: &DiceSetting{
			Rate: 10,
			Add1: ":0:0",
		},
	}
	return m
}

func (m *DiceGame) getDiceOrder(uid uint) (*entities.DiceGameOrder, error) {
	order, err := m.Srv.GetUserDiceGameOrder(uid)
	if err != nil {
		return nil, err
	}
	seed2, _ := utils.GenerateSecureHex()
	if order == nil {
		seed1, _ := utils.GenerateSecureHex()
		order = &entities.DiceGameOrder{
			UID:        uid,
			RoundID:    1,
			ClientSeed: seed1,
			ServerSeed: seed2,
		}
		err = m.Srv.CreateDiceGameOrder(order)
		if err != nil {
			return nil, err
		}
	} else if order.Settled == 1 {
		seed1 := order.ClientSeed
		order = &entities.DiceGameOrder{
			UID:        uid,
			RoundID:    order.RoundID + 1,
			ClientSeed: seed1,
			ServerSeed: seed2,
		}
		err = m.Srv.CreateDiceGameOrder(order)
		if err != nil {
			return nil, err
		}
	}

	return order, nil
}

// buildDiceState
func (m *DiceGame) buildDiceState(order *entities.DiceGameOrder) (*entities.DiceGameState, error) {
	originalHash := fmt.Sprintf("%s%s", order.ClientSeed, order.ServerSeed)
	return &entities.DiceGameState{
		RoundID:      order.RoundID,
		ClientSeed:   order.ClientSeed,
		ServerSeed:   order.ServerSeed,
		OpenHash:     fmt.Sprintf("%x", sha256.Sum256([]byte(originalHash))),
		Target:       order.Target,
		Result:       order.Result,
		IsAbove:      order.IsAbove,
		Multiple:     order.Multiple,
		BetTime:      order.BetTime,
		BetAmount:    order.BetAmount,
		RewardAmount: order.RewardAmount,
		Settled:      order.Settled,
		EndTime:      order.EndTime,
	}, nil
}

// GetOrderList
func (m *DiceGame) GetOrderList(uid uint) ([]*entities.DiceGameState, error) {
	list, err := m.Srv.GetUserDiceGameOrderList(uid)
	if err != nil {
		return nil, err
	}
	states := make([]*entities.DiceGameState, 0, len(list))
	for i := range list {
		order := list[i]
		state, err := m.buildDiceState(order)
		if err != nil {
			return nil, err
		}
		states = append(states, state)
	}
	return states, nil
}

// PlaceBet
func (m *DiceGame) PlaceBet(req *entities.DiceGamePlaceBetReq) (*entities.DiceGameState, error) {
	if err := m.checkPlaceBetReq(req); err != nil {
		return nil, err
	}
	order, err := m.getDiceOrder(req.UID)
	if err != nil {
		return nil, err
	}
	order.Target = req.Target
	order.IsAbove = req.IsAbove
	order.BetTime = time.Now().Unix()
	order.BetAmount = req.BetAmount

	// place order
	if err := m.Srv.PlaceOrder(order); err != nil {
		return nil, err
	}

	// settle win
	// generate dice result
	order.Result = m.generateDiceResult(order.ClientSeed, order.ServerSeed)
	order.Multiple = m.calcMultiple(order.Target, order.Result, order.IsAbove, int(m.setting.Rate))
	order.RewardAmount = m.calcRewardAmount(order)
	logger.ZInfo("PlaceBet betting result", zap.Any("order", order))
	if err := m.Srv.SettleOrder(order); err != nil {
		return nil, err
	}

	state, err := m.buildDiceState(order)
	if err != nil {
		return nil, err
	}
	return state, nil
}

func (m *DiceGame) checkPlaceBetReq(req *entities.DiceGamePlaceBetReq) error {
	if req.BetAmount <= 0 {
		return errors.New("bet amount must greater than 0")
	}
	if req.Target <= 0 || req.Target >= 100 {
		return errors.New("target must be between 0 and 100")
	}
	if req.IsAbove != 0 && req.IsAbove != 1 {
		return errors.New("is_above must be 0 or 1")
	}
	return nil
}

// generateDiceResult
func (m *DiceGame) generateDiceResult(clientSeed, serverSeed string) float64 {
	crashHash1 := utils.HmacSHA256(serverSeed, clientSeed+m.setting.Add1)
	// 计算结果
	k, _ := strconv.ParseInt(crashHash1[:8], 16, 64)
	r := math.Floor(float64(k)/(math.MaxUint32+1)*float64(10001)) / 100
	return r
}

// calcMultiple
func (m *DiceGame) calcMultiple(target, result float64, isAbove, rate int) float64 {
	k := 1.0
	if isAbove == 0 {
		if target >= result {
			k = target / 100
		} else {
			return 0
		}
	} else {
		if target <= result {
			k = (100 - target) / 100
		} else {
			return 0
		}
	}
	p := (100 - float64(rate)/10) / (k * 100)
	p = math.Round(p*10000) / 10000
	return p
}

// calcRewardAmount
func (m *DiceGame) calcRewardAmount(order *entities.DiceGameOrder) float64 {
	return math.Round(order.Delivery*order.Multiple*100) / 100
}

// ChangeSeed
func (m *DiceGame) ChangeSeed(req *entities.DiceGameChangeSeedReq) (*entities.DiceGameChangeSeedRsp, error) {
	if err := m.checkChangeSeedReq(req); err != nil {
		return nil, err
	}
	order, err := m.getDiceOrder(req.UID)
	if err != nil {
		return nil, err
	}

	// update order
	order.ClientSeed = req.ClientSeed
	order.ServerSeed, _ = utils.GenerateSecureHex()
	originalHash := fmt.Sprintf("%s%s", order.ClientSeed, order.ServerSeed)
	if err := m.Srv.UpdateDiceGameOrder(order); err != nil {
		return nil, err
	}
	return &entities.DiceGameChangeSeedRsp{
		ClientSeed: order.ClientSeed,
		OpenHash:   fmt.Sprintf("%x", sha256.Sum256([]byte(originalHash))),
	}, nil
}

// checkChangeSeedReq
func (m *DiceGame) checkChangeSeedReq(req *entities.DiceGameChangeSeedReq) error {
	if req.ClientSeed == "" {
		return errors.New("client seed is empty")
	}
	return nil
}

func FairCheck(req *entities.FairCheckReq) (*entities.FairCheckRsp, error) {
	m := NewDiceGame(nil)
	r := m.generateDiceResult(req.ClientSeed, req.ServerSeed)
	return &entities.FairCheckRsp{Result: r}, nil
}
