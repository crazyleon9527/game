package game

import (
	"context"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/errors"
	"rk-api/internal/app/service"
	"rk-api/internal/app/utils"
	"rk-api/pkg/logger"
	"time"

	"github.com/looplab/fsm"
	"go.uber.org/zap"
)

// 提供一个混合多个房间的 Wingo 。 用同一个FSM 控制,房间状态是完全同步。
type MultiWingo struct {
	Fsm     *fsm.FSM //使用一个fsm
	Setting *entities.WingoRoomSetting
	Wingo   //四个房间独立FSM 控制
}

func NewMultiWingo(srv *service.WingoService) *MultiWingo {
	return &MultiWingo{
		Wingo: *NewWingo(srv),
	}
}

func (r *MultiWingo) Init() error {

	r.Srv.SimulateSettleWingo() // 先把旧的未处理的期数处理完毕
	list, err := r.Srv.GetWingoSettingList()
	if err != nil {
		return errors.With(" wingo init fail" + err.Error())
	}

	r.Fsm = fsm.NewFSM( //使用一个 状态机控制所有房间，房间就直接同步
		STATE_INIT,
		fsm.Events{
			{Name: EVENT_START, Src: []string{STATE_INIT, STATE_SETTLE}, Dst: STATE_BETTING},
			{Name: EVENT_STOP_BETTING, Src: []string{STATE_BETTING}, Dst: STATE_WAITING},
			{Name: EVENT_SETTLE, Src: []string{STATE_WAITING}, Dst: STATE_SETTLE},
			{Name: EVENT_STOP, Src: []string{STATE_WAITING, STATE_SETTLE, STATE_BETTING}, Dst: STATE_INIT},
		},
		fsm.Callbacks{
			"enter_state": func(_ context.Context, e *fsm.Event) { r.enterState(e) },
		},
	)

	for _, setting := range list {
		room := NewWingoRoom(setting, r.Srv)

		if err := room.init(); err != nil {
			logger.ZError("wingo room init fail",
				zap.Uint8("bet_type", room.Setting.BetType),
				zap.Error(err),
			)
		} else {
			logger.ZInfo("wingo room init succ ",
				zap.Uint("ID", room.Setting.ID),
				zap.Uint8("bet_type", room.Setting.BetType),
			)
			r.RoomMap.Store(room.Setting.BetType, room)
		}
		r.Setting = setting //拿一个
	}

	return nil
}

func (r *MultiWingo) enterState(e *fsm.Event) {
	// fmt.Printf("wingo enterState %s\n", e.Dst)

	logger.ZInfo("wingo state",
		zap.String("src", e.Src),
		zap.String("Dst", e.Dst),
	)

	r.RoomMap.Range(func(key, value interface{}) bool {
		room, _ := value.(*WingoRoom)
		room.StateSTime = time.Now().Unix()
		room.state = r.Fsm.Current() //设置房间的状态
		return true
	})

	if e.Dst == STATE_BETTING {
		//获取setting
		var err error
		r.RoomMap.Range(func(key, value interface{}) bool {
			room, _ := value.(*WingoRoom)
			room.RoundSTime = time.Now().Unix()
			if err = room.periodNext(); err != nil {
				return false
			}
			return true
		})

		if err != nil {
			r.Fsm.Event(context.Background(), EVENT_STOP) //停止 切换到初始状态
		} else {
			time.AfterFunc(time.Duration(r.Setting.BettingInterval)*time.Second, func() {
				r.Fsm.Event(context.Background(), EVENT_STOP_BETTING)
			})
		}

	} else if e.Dst == STATE_WAITING {

		time.AfterFunc(time.Duration(r.Setting.StopBettingInterval-r.Setting.SettleInterval)*time.Second, func() {
			r.Fsm.Event(context.Background(), EVENT_SETTLE)
		})
	} else if e.Dst == STATE_SETTLE {
		go r.settle()
		time.AfterFunc(time.Duration(r.Setting.SettleInterval)*time.Second, func() {

			r.Fsm.Event(context.Background(), EVENT_START)
		})
	} else if e.Dst == STATE_INIT {
		//  可能因异常 切换到初始化状态
		logger.ZError(" e.Dst == STATE_INIT")
	}

}

func (r *MultiWingo) settle() {

	defer utils.PrintPanicStack()

	r.RoomMap.Range(func(key, value interface{}) bool {
		room, _ := value.(*WingoRoom)
		start := time.Now().Unix()
		go room.settle(room.Period, room.orders)
		logger.ZInfo("MultiWingo room settle take", zap.Int8("betType", int8(room.Period.BetType)), zap.Int64("time", time.Now().Unix()-start))
		return true
	})
}

// 为0启动所有
func (r *MultiWingo) Start() {
	defer utils.PrintPanicStack()

	r.Fsm.Event(context.Background(), EVENT_START) //开启
}
