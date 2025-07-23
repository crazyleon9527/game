package hash

import (
	"context"
	"rk-api/internal/app/errors"
	"rk-api/internal/app/service"
	"sync"

	"github.com/google/wire"
)

// GameStrategyType 游戏策略
type GameStrategyType uint8

const (
	GameStrategyTypeNone GameStrategyType = iota
	GameStrategyTypeSingleDouble
	GameStrategyTypeSmallBig
	GameStrategyTypeBullBull
	GameStrategyTypeBankerPlayerTie
	GameStrategyTypeLucky
	GameStrategyTypeLimit
)

// RoomType 游戏类型
type RoomType uint8

const (
	RoomTypeNone RoomType = iota
	RoomTypeNormal
	RoomTypeMid
	RoomTypeHigh
	RoomTypeLimit
)

var GameManageSet = wire.NewSet(
	NewGameManage, // 直接提供结构体指针
)

type GameManage struct {
	Srv          *service.HashGameService
	GameMap      sync.Map
	GameRegistry *GameRegistry
}

func NewGameManage(srv *service.HashGameService) *GameManage {
	m := &GameManage{
		Srv:          srv,
		GameRegistry: NewGameRegistry(),
	}
	for gameStrategyType := GameStrategyTypeNone + 1; gameStrategyType < GameStrategyTypeLimit; gameStrategyType++ {
		strategy, sexists := m.GameRegistry.GetStrategy(gameStrategyType)
		rifunc, rexists := m.GameRegistry.GetRifunc(gameStrategyType)
		if sexists && rexists {
			g := NewGame(srv, strategy, rifunc)
			m.GameMap.Store(gameStrategyType, g)
		}
	}
	return m
}

func (m *GameManage) Get(gameStrategyType GameStrategyType, betType RoomType) (IGameRoom, error) {
	if game, ok := m.GameMap.Load(gameStrategyType); ok {
		if g, ok := game.(*Game); ok {
			return g.Get(betType)
		}
	}
	return nil, errors.WithCode(errors.GameStrategyNotExist)
}

type Game struct {
	Srv     *service.HashGameService
	RoomMap sync.Map
}

func NewGame(srv *service.HashGameService, strategy GameStrategy, rifunc func(*service.HashGameService, GameStrategy) IGameRoom) *Game {
	g := &Game{Srv: srv}
	for rtype := RoomTypeNone + 1; rtype < RoomTypeLimit; rtype++ {
		room := rifunc(srv, strategy)
		g.RoomMap.Store(rtype, room)
	}
	g.Start() //暂时不开
	return g
}

func (r *Game) Start() {
	r.RoomMap.Range(func(key, value interface{}) bool {
		if room, ok := value.(IGameRoom); ok {
			room.Start(context.Background()) // 启动监听
		}
		return true
	})
}

func (r *Game) Get(betType RoomType) (IGameRoom, error) {
	if room, ok := r.RoomMap.Load(betType); ok {
		return room.(IGameRoom), nil
	}
	return nil, errors.WithCode(errors.RoomNotExist)
}
