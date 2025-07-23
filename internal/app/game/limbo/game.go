package limbo

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

var LimboGameSet = wire.NewSet(
	NewLimboGame,
)

type LimboGame struct {
	Srv     *service.LimboGameService
	setting *LimboSetting
}

type LimboSetting struct {
	Rate uint8
	Add1 string
}

func NewLimboGame(srv *service.LimboGameService) *LimboGame {
	m := &LimboGame{
		Srv: srv,
		setting: &LimboSetting{
			Rate: 10,
			Add1: ":0:0",
		},
	}
	return m
}

func (m *LimboGame) getLimboOrder(uid uint) (*entities.LimboGameOrder, error) {
	order, err := m.Srv.GetUserLimboGameOrder(uid)
	if err != nil {
		return nil, err
	}
	seed2, _ := utils.GenerateSecureHex()
	if order == nil {
		seed1, _ := utils.GenerateSecureHex()
		order = &entities.LimboGameOrder{
			UID:        uid,
			RoundID:    1,
			ClientSeed: seed1,
			ServerSeed: seed2,
		}
		err = m.Srv.CreateLimboGameOrder(order)
		if err != nil {
			return nil, err
		}
	} else if order.Settled == 1 {
		seed1 := order.ClientSeed
		order = &entities.LimboGameOrder{
			UID:        uid,
			RoundID:    order.RoundID + 1,
			ClientSeed: seed1,
			ServerSeed: seed2,
		}
		err = m.Srv.CreateLimboGameOrder(order)
		if err != nil {
			return nil, err
		}
	}

	return order, nil
}

func (m *LimboGame) buildLimboState(order *entities.LimboGameOrder) (*entities.LimboGameState, error) {
	originalHash := fmt.Sprintf("%s%s", order.ClientSeed, order.ServerSeed)
	return &entities.LimboGameState{
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

func (m *LimboGame) GetOrderList(uid uint) ([]*entities.LimboGameState, error) {
	list, err := m.Srv.GetUserLimboGameOrderList(uid)
	if err != nil {
		return nil, err
	}
	states := make([]*entities.LimboGameState, 0, len(list))
	for i := range list {
		order := list[i]
		state, err := m.buildLimboState(order)
		if err != nil {
			return nil, err
		}
		states = append(states, state)
	}
	return states, nil
}

func (m *LimboGame) PlaceBet(req *entities.LimboGamePlaceBetReq) (*entities.LimboGameState, error) {
	if err := m.checkPlaceBetReq(req); err != nil {
		return nil, err
	}
	order, err := m.getLimboOrder(req.UID)
	if err != nil {
		return nil, err
	}
	order.Target = req.Target
	order.IsAbove = req.IsAbove
	order.BetTime = time.Now().Unix()
	order.BetAmount = req.BetAmount

	if err := m.Srv.PlaceOrder(order); err != nil {
		return nil, err
	}

	order.Result = m.generateLimboResult(order.ClientSeed, order.ServerSeed)
	order.Multiple = m.calcMultiple(order.Target, order.Result, order.IsAbove, int(m.setting.Rate))
	order.RewardAmount = m.calcRewardAmount(order)
	logger.ZInfo("PlaceBet betting result", zap.Any("order", order))
	if err := m.Srv.SettleOrder(order); err != nil {
		return nil, err
	}

	state, err := m.buildLimboState(order)
	if err != nil {
		return nil, err
	}
	return state, nil
}

func (m *LimboGame) checkPlaceBetReq(req *entities.LimboGamePlaceBetReq) error {
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

func (m *LimboGame) generateLimboResult(clientSeed, serverSeed string) float64 {
	crashHash1 := utils.HmacSHA256(serverSeed, clientSeed+m.setting.Add1)
	k, _ := strconv.ParseInt(crashHash1[:8], 16, 64)
	r := math.Floor(float64(k)/(math.MaxUint32+1)*float64(10001)) / 100
	return r
}

func (m *LimboGame) calcMultiple(target, result float64, isAbove, rate int) float64 {
	k := 1.0
	if isAbove == 0 {
		if target <= result {
			k = target / 100
		} else {
			return 0
		}
	} else {
		if target >= result {
			k = (100 - target) / 100
		} else {
			return 0
		}
	}
	p := (100 - float64(rate)/10) / (k * 100)
	p = math.Round(p*10000) / 10000
	return p
}

func (m *LimboGame) calcRewardAmount(order *entities.LimboGameOrder) float64 {
	return math.Round(order.Delivery*order.Multiple*100) / 100
}

func (m *LimboGame) ChangeSeed(req *entities.LimboGameChangeSeedReq) (*entities.LimboGameChangeSeedRsp, error) {
	if err := m.checkChangeSeedReq(req); err != nil {
		return nil, err
	}
	order, err := m.getLimboOrder(req.UID)
	if err != nil {
		return nil, err
	}

	order.ClientSeed = req.ClientSeed
	order.ServerSeed, _ = utils.GenerateSecureHex()
	originalHash := fmt.Sprintf("%s%s", order.ClientSeed, order.ServerSeed)
	if err := m.Srv.UpdateLimboGameOrder(order); err != nil {
		return nil, err
	}
	return &entities.LimboGameChangeSeedRsp{
		ClientSeed: order.ClientSeed,
		OpenHash:   fmt.Sprintf("%x", sha256.Sum256([]byte(originalHash))),
	}, nil
}

func (m *LimboGame) checkChangeSeedReq(req *entities.LimboGameChangeSeedReq) error {
	if req.ClientSeed == "" {
		return errors.New("client seed is empty")
	}
	return nil
}

func FairCheck(req *entities.FairCheckReq) (*entities.FairCheckRsp, error) {
	m := NewLimboGame(nil)
	r := m.generateLimboResult(req.ClientSeed, req.ServerSeed)
	return &entities.FairCheckRsp{Result: r}, nil
}