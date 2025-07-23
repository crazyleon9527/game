package mine

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"rk-api/internal/app/constant"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/service"
	"rk-api/internal/app/utils"
	"rk-api/pkg/cjson"
	"rk-api/pkg/logger"
	"strconv"
	"time"

	"github.com/google/wire"
	"github.com/shopspring/decimal"
	"github.com/spf13/cast"
	"go.uber.org/zap"
)

// 回合状态定义
const (
	GameStatusPreparing string = "preparing"
	GameStatusPlaying   string = "playing"
	GameStatusGameOver  string = "gameover"
)

var MineGameSet = wire.NewSet(
	NewMineGame, // 直接提供结构体指针
)

type MineGame struct {
	Srv     *service.MineGameService
	setting *MineSetting
}

type MineSetting struct {
	// 抽水比例 1%
	Rate uint8

	// 增量值1：客户端种子+“:0:0”
	Add1 string
	// 增量值2：客户端种子+“:0:1”
	Add2 string
	// 增量值3：客户端种子+“:0:2”
	Add3 string
}

func NewMineGame(srv *service.MineGameService) *MineGame {
	m := &MineGame{
		Srv: srv,
		// blockFetcher: srv.BlockFetcher,
		setting: &MineSetting{
			Rate: 10,
			Add1: ":0:0",
			Add2: ":0:1",
			Add3: ":0:2",
		},
	}
	return m
}

// GetState
func (m *MineGame) GetState(uid uint) (*entities.MineGameState, error) {
	order, err := m.getMineOrder(uid)
	if err != nil {
		return nil, err
	}

	return m.buildMineState(order)
}

func (m *MineGame) getMineOrder(uid uint) (*entities.MineGameOrder, error) {
	order, err := m.Srv.GetUserMineGameOrder(uid)
	if err != nil {
		return nil, err
	}
	if order == nil {
		seed1, _ := utils.GenerateSecureHex()
		seed2, _ := utils.GenerateSecureHex()
		order = &entities.MineGameOrder{
			UID:          uid,
			RoundID:      1,
			Status:       GameStatusPreparing,
			ClientSeed:   seed1,
			ServerSeed:   seed2,
			MinePosition: "[]",
			OpenPosition: "[]",
		}
		err = m.Srv.CreateMineGameOrder(order)
		if err != nil {
			return nil, err
		}
	} else if order.Status == GameStatusGameOver {
		seed1 := order.ClientSeed
		seed2, _ := utils.GenerateSecureHex()
		order = &entities.MineGameOrder{
			UID:          uid,
			RoundID:      order.RoundID + 1,
			Status:       GameStatusPreparing,
			ClientSeed:   seed1,
			ServerSeed:   seed2,
			MinePosition: "[]",
			OpenPosition: "[]",
		}
		err = m.Srv.CreateMineGameOrder(order)
		if err != nil {
			return nil, err
		}
	}

	return order, nil
}

// buildMineState
func (m *MineGame) buildMineState(order *entities.MineGameOrder) (*entities.MineGameState, error) {
	serverSeed := ""
	var minePosition []int
	if order.Settled == constant.STATUS_SETTLE {
		serverSeed = order.ServerSeed
		if order.MinePosition != "" {
			if err := json.Unmarshal([]byte(order.MinePosition), &minePosition); err != nil {
				return nil, err
			}
		}
	}
	var openPosition []*entities.MineGamePosition
	if order.OpenPosition != "" {
		if err := json.Unmarshal([]byte(order.OpenPosition), &openPosition); err != nil {
			return nil, err
		}
	}

	originalHash := fmt.Sprintf("%s%s", order.ClientSeed, order.ServerSeed)
	return &entities.MineGameState{
		RoundID:      order.RoundID,
		Status:       order.Status,
		ClientSeed:   order.ClientSeed,
		ServerSeed:   serverSeed,
		OpenHash:     fmt.Sprintf("%x", sha256.Sum256([]byte(originalHash))),
		MineCount:    order.MineCount,
		DiamondLeft:  order.DiamondLeft,
		MinePosition: minePosition,
		OpenPosition: openPosition,
		Multiple:     order.Multiple,
		BetTime:      order.BetTime,
		BetAmount:    order.BetAmount,
		RewardAmount: order.RewardAmount,
		Settled:      order.Settled,
		EndTime:      order.EndTime,
	}, nil
}

// GetOrderList
func (m *MineGame) GetOrderList(uid uint) ([]*entities.MineGameState, error) {
	list, err := m.Srv.GetUserMineGameOrderList(uid)
	if err != nil {
		return nil, err
	}
	states := make([]*entities.MineGameState, 0, len(list))
	for i := range list {
		order := list[i]
		state, err := m.buildMineState(order)
		if err != nil {
			return nil, err
		}
		states = append(states, state)
	}
	return states, nil
}

// PlaceBet
func (m *MineGame) PlaceBet(req *entities.MineGamePlaceBetReq) (*entities.MineGameState, error) {
	if err := m.checkPlaceBetReq(req); err != nil {
		return nil, err
	}
	order, err := m.getMineOrder(req.UID)
	if err != nil {
		return nil, err
	}
	if err := m.checkOrderStatus(order, GameStatusPreparing); err != nil {
		return nil, err
	}
	order.Status = GameStatusPlaying
	order.MineCount = req.MineCount
	order.DiamondLeft = 25 - req.MineCount
	order.Multiple = 1
	order.BetTime = time.Now().Unix()
	order.BetAmount = req.BetAmount

	// generate mine position
	minePosition := m.gererateMinePosition(order.ClientSeed, order.ServerSeed, order.MineCount)
	order.MinePosition = cjson.StringifyIgnore(minePosition)

	state, err := m.buildMineState(order)
	if err != nil {
		return nil, err
	}

	// place order
	if err := m.Srv.PlaceOrder(order); err != nil {
		return nil, err
	}

	return state, nil
}

func (m *MineGame) checkPlaceBetReq(req *entities.MineGamePlaceBetReq) error {
	if req.BetAmount <= 0 {
		return errors.New("bet amount must greater than 0")
	}
	if req.MineCount <= 0 || req.MineCount >= 25 {
		return errors.New("mine count must in [1,24]")
	}
	return nil
}

func (m *MineGame) checkOrderStatus(order *entities.MineGameOrder, status string) error {
	if order.Status != status {
		return errors.New("invalid order status: " + order.Status)
	}
	if order.Settled == constant.STATUS_SETTLE {
		return errors.New("order has settled")
	}
	return nil
}

// gererateMinePosition
func (m *MineGame) gererateMinePosition(clientSeed, serverSeed string, count int) []int {
	minePosition := make([]int, 0, count)
	crashHash1 := utils.HmacSHA256(serverSeed, clientSeed+m.setting.Add1)
	crashHash2 := utils.HmacSHA256(serverSeed, clientSeed+m.setting.Add2)
	crashHash3 := utils.HmacSHA256(serverSeed, clientSeed+m.setting.Add3)

	// 随机位置
	randompos1 := m.generateMineRandomPosition(crashHash1, 25)
	randompos2 := m.generateMineRandomPosition(crashHash2, 17)
	randompos3 := m.generateMineRandomPosition(crashHash3, 9)
	randompos1 = append(randompos1, randompos2...)
	randompos1 = append(randompos1, randompos3...)

	// 计算炸弹位置
	customShuffled := make([]int, 25)
	for i := 0; i < 25; i++ {
		customShuffled[i] = i
	}
	for i := 0; i < count; i++ {
		// 取出随机数对应的元素
		randomIndex := randompos1[i]
		if randomIndex < 1 || randomIndex+i >= len(customShuffled) {
			continue
		}
		selected := customShuffled[randomIndex+i]

		// 从原位置移除
		customShuffled = append(customShuffled[:randomIndex+i], customShuffled[randomIndex+i+1:]...)
		// 插入到最前面
		head := make([]int, i)
		copy(head, customShuffled[:i])
		head = append(head, selected)
		customShuffled = append(head, customShuffled[i:]...)
	}
	minePosition = append(minePosition, customShuffled[:count]...)

	return minePosition
}

// generateMineRandomPosition
func (m *MineGame) generateMineRandomPosition(hash string, multiple int) []int {
	pos := make([]int, 0, 8)
	// 5e050c2b 9a80f8c1 5fd0c04d b8d401e3 db8ddad7 433a2a5d c1b042d6 e14d15eb
	for i := 0; i < 8; i++ {
		k, _ := strconv.ParseInt(hash[i*8:i*8+8], 16, 64)
		p := math.Floor(float64(k) / (math.MaxUint32 + 1) * float64(multiple-i))
		pos = append(pos, int(p))
	}
	return pos
}

// OpenPosition
func (m *MineGame) OpenPosition(req *entities.MineGameOpenPositionReq) (*entities.MineGameState, error) {
	if err := m.checkOpenPositionReq(req); err != nil {
		return nil, err
	}
	order, err := m.getMineOrder(req.UID)
	if err != nil {
		return nil, err
	}
	if err := m.checkOrderStatus(order, GameStatusPlaying); err != nil {
		return nil, err
	}

	// check open position
	var openPosition []*entities.MineGamePosition
	if order.OpenPosition != "" {
		if err := json.Unmarshal([]byte(order.OpenPosition), &openPosition); err != nil {
			return nil, err
		}
	}
	if err := m.checkPosition(openPosition, req.OpenPosition); err != nil {
		return nil, err
	}
	// check mine position
	var minePosition []int
	if order.MinePosition != "" {
		if err := json.Unmarshal([]byte(order.MinePosition), &minePosition); err != nil {
			return nil, err
		}
	}
	isOpenMine := m.checkIsOpenMine(minePosition, req.OpenPosition)

	// calc multiple
	if !isOpenMine {
		multiple := m.calcMultiple(order.MineCount, len(openPosition)+1, int(m.setting.Rate))
		order.Multiple = math.Round(multiple*100) / 100
	} else {
		order.Multiple = 0
	}
	openPosition = append(openPosition, &entities.MineGamePosition{
		Position: req.OpenPosition,
		Multiple: order.Multiple,
	})
	order.OpenPosition = cjson.StringifyIgnore(openPosition)

	if isOpenMine {
		// settle lose update order
		order.Status = GameStatusGameOver
		order.Settled, order.EndTime = constant.STATUS_SETTLE, time.Now().Unix()
		if err := m.Srv.UpdateMineGameOrder(order); err != nil {
			return nil, err
		}
	} else {
		order.DiamondLeft--
		if order.DiamondLeft == 0 {
			// settle win
			order.Status = GameStatusGameOver
			order.RewardAmount = m.calcRewardAmount(order)
			logger.ZInfo("OpenPosition betting result", zap.Any("order", order))
			if err := m.Srv.SettleOrder(order); err != nil {
				return nil, err
			}
		} else {
			// update order
			if err := m.Srv.UpdateMineGameOrder(order); err != nil {
				return nil, err
			}
		}
	}

	state, err := m.buildMineState(order)
	if err != nil {
		return nil, err
	}
	return state, nil
}

// checkOpenPositionReq
func (m *MineGame) checkOpenPositionReq(req *entities.MineGameOpenPositionReq) error {
	if req.OpenPosition < 0 || req.OpenPosition > 24 {
		return errors.New("open position must in [0,24]")
	}
	return nil
}

// checkPosition
func (m *MineGame) checkPosition(openPosition []*entities.MineGamePosition, position int) error {
	for _, pos := range openPosition {
		if pos.Position == position {
			return errors.New("position has been opened")
		}
	}
	return nil
}

// checkIsOpenMine
func (m *MineGame) checkIsOpenMine(minePosition []int, position int) bool {
	for i := range minePosition {
		if minePosition[i] == position {
			return true
		}
	}
	return false
}

// calcMultiple
func (m *MineGame) calcMultiple(mineCount, openCount, rate int) float64 {
	k := 1.0
	for i := openCount; i > 0; i-- {
		k *= (1.0 - float64(mineCount)/float64(26-i))
	}
	// k最多保留15位数字，和excel一致
	for i := 1; i < 15; i++ {
		if k2 := math.Pow10(i) * k; k2 >= 1 {
			k = math.Floor(k*math.Pow10(14+i)) / math.Pow10(14+i)
			break
		}
	}
	p := decimal.NewFromFloat(1 - float64(rate)/1000).Div(decimal.NewFromFloat(k))
	return p.InexactFloat64()
}

// calcRewardAmount
func (m *MineGame) calcRewardAmount(order *entities.MineGameOrder) float64 {
	return math.Round(order.Delivery*order.Multiple*100) / 100
}

// Cashout
func (m *MineGame) Cashout(uid uint) (*entities.MineGameState, error) {
	order, err := m.getMineOrder(uid)
	if err != nil {
		return nil, err
	}
	if err := m.checkOrderStatus(order, GameStatusPlaying); err != nil {
		return nil, err
	}
	if err := m.checkOrderPlaying(order); err != nil {
		return nil, err
	}

	// settle win
	order.Status = GameStatusGameOver
	order.RewardAmount = m.calcRewardAmount(order)
	logger.ZInfo("Cashout betting result", zap.Any("order", order))
	if err := m.Srv.SettleOrder(order); err != nil {
		return nil, err
	}

	state, err := m.buildMineState(order)
	if err != nil {
		return nil, err
	}
	return state, nil
}

// checkOrderPlaying
func (m *MineGame) checkOrderPlaying(order *entities.MineGameOrder) error {
	if order.Multiple == 1 || order.DiamondLeft == 25-order.MineCount {
		return errors.New("order is not playing")
	}
	return nil
}

// ChangeSeed
func (m *MineGame) ChangeSeed(req *entities.MineGameChangeSeedReq) (*entities.MineGameChangeSeedRsp, error) {
	if err := m.checkChangeSeedReq(req); err != nil {
		return nil, err
	}
	order, err := m.getMineOrder(req.UID)
	if err != nil {
		return nil, err
	}
	if err := m.checkOrderStatus(order, GameStatusPreparing); err != nil {
		return nil, err
	}

	// update order
	order.ClientSeed = req.ClientSeed
	order.ServerSeed, _ = utils.GenerateSecureHex()
	originalHash := fmt.Sprintf("%s%s", order.ClientSeed, order.ServerSeed)
	if err := m.Srv.UpdateMineGameOrder(order); err != nil {
		return nil, err
	}
	return &entities.MineGameChangeSeedRsp{
		ClientSeed: order.ClientSeed,
		OpenHash:   fmt.Sprintf("%x", sha256.Sum256([]byte(originalHash))),
	}, nil
}

// checkChangeSeedReq
func (m *MineGame) checkChangeSeedReq(req *entities.MineGameChangeSeedReq) error {
	if req.ClientSeed == "" {
		return errors.New("client seed is empty")
	}
	return nil
}

func FairCheck(req *entities.FairCheckReq) (*entities.FairCheckRsp, error) {
	m := NewMineGame(nil)
	mineCount := 24
	if v, ok := req.Ext["mine_count"]; ok {
		if count := cast.ToInt(v); count > 0 && count < 25 {
			mineCount = count
		}
	}
	minePosition := m.gererateMinePosition(req.ClientSeed, req.ServerSeed, mineCount)
	minePositionStr := cjson.StringifyIgnore(minePosition)
	return &entities.FairCheckRsp{ResultJson: minePositionStr}, nil
}
