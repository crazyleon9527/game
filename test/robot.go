package test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/game"
	"rk-api/internal/app/ginx"
	"rk-api/internal/app/middleware"
	"rk-api/pkg/http"
	"rk-api/pkg/logger"
	"rk-api/pkg/structure"
	"strings"
	"sync"
	"time"

	"github.com/looplab/fsm"
	"go.uber.org/zap"
)

// 生产环境中不要频繁改变Seed，这里为示例需要，每次运行都使用不同的Seed

func init() {
	rand.Seed(time.Now().UnixNano())
}

func mapToStruct(m interface{}, s interface{}) error {

	return structure.MapToStruct(m, s)
}

type ApiRobot struct {
	initCount int

	Fsm      *fsm.FSM
	token    string
	apiUrl   string
	username string
	password string
	uid      uint
	room     *entities.WingoRoomResp

	quit    chan struct{}
	started bool
	mu      sync.Mutex // 保护started字段

	logger func(method, log string)
}

const (
	ROBOT_STATE_INIT         = "ROBOT_STATE_INIT"
	ROBOT_STATE_LOGINING     = "ROBOT_STATE_LOGINING"
	ROBOT_STATE_IDLE         = "ROBOT_STATE_IDLE"
	ROBOT_STATE_ROOM         = "ROBOT_STATE_ROOM"
	ROBOT_STATE_ROOM_BETTING = "ROBOT_STATE_ROOM_BETTING"
)

const (
	EVENT_START         = "EVENT_START"
	EVENT_FREE          = "EVENT_FREE"
	EVENT_LOGOUT        = "EVENT_LOGOUT"
	EVENT_ROOM          = "EVENT_ROOM"
	EVENT_START_BETTING = "EVENT_START_BETTING"
)

func NewApiRobot(username string, password string) *ApiRobot {
	r := ApiRobot{
		username: username,
		password: password,
	}
	r.apiUrl = "https://api-dev.cheetahs.asia/api/"
	r.logger = func(method, log string) {
		logger.Info(fmt.Sprintf("user:%s %s %s", r.username, method, log))
	}
	return &r
}

func (r *ApiRobot) SetAPIUrl(apiUrl string) {
	r.apiUrl = apiUrl
}

func (r *ApiRobot) SetLogger(logger func(log string)) {
	r.logger = func(method, log string) {
		logger(fmt.Sprintf("user:%s <span style='color: red;'>%s</span> %s", r.username, method, log))
	}
}

func (r *ApiRobot) Start() {
	r.Fsm = fsm.NewFSM(
		ROBOT_STATE_INIT,
		fsm.Events{
			{Name: EVENT_START, Src: []string{ROBOT_STATE_INIT}, Dst: ROBOT_STATE_LOGINING},
			{Name: EVENT_FREE, Src: []string{ROBOT_STATE_LOGINING, ROBOT_STATE_ROOM}, Dst: ROBOT_STATE_IDLE},
			{Name: EVENT_ROOM, Src: []string{ROBOT_STATE_IDLE}, Dst: ROBOT_STATE_ROOM},
			{Name: EVENT_START_BETTING, Src: []string{ROBOT_STATE_ROOM}, Dst: ROBOT_STATE_ROOM_BETTING},
			{Name: EVENT_LOGOUT, Src: []string{ROBOT_STATE_LOGINING, ROBOT_STATE_IDLE, ROBOT_STATE_ROOM, ROBOT_STATE_ROOM_BETTING}, Dst: ROBOT_STATE_INIT},
		},
		fsm.Callbacks{
			"enter_state": func(_ context.Context, e *fsm.Event) { r.enterState(e) },
		},
	)
	r.Fsm.Event(context.Background(), EVENT_START)
}

func (r *ApiRobot) enterState(e *fsm.Event) {
	// logger.ZInfo("enterSate", zap.String("username", r.username), zap.String("state", e.FSM.Current()))
	if e.Dst == ROBOT_STATE_INIT {
		r.initCount++
		r._init_()
	} else if e.Dst == ROBOT_STATE_LOGINING {
		r._logining_()
	} else if e.Dst == ROBOT_STATE_IDLE {
		r._idle_()
	} else if e.Dst == ROBOT_STATE_ROOM {
		r._room_()
	} else if e.Dst == ROBOT_STATE_ROOM_BETTING {
		r._betting_()
	}
}

func (r *ApiRobot) _init_() {
	if r.initCount == 1 { //第一次直接启动
		r.Fsm.Event(context.Background(), EVENT_START)
	} else {
		time.AfterFunc(10*time.Second, func() {
			r.Fsm.Event(context.Background(), EVENT_START)
		})
	}
}

func (r *ApiRobot) _logining_() {
	// user, err := r.getUser(r.username)
	// if err != nil {
	// 	r.Fsm.Event(context.Background(), EVENT_START)
	// 	return
	// }

	// if user == nil {
	// 	err = r.register()
	// 	if err != nil {
	// 		r.Fsm.Event(context.Background(), EVENT_START)
	// 		return
	// 	}
	// }
	// loginResp, err := r.login(r.username, r.password)
	// if err != nil {
	// 	r.Fsm.Event(context.Background(), EVENT_START)
	// 	return
	// }
	// r.token = loginResp.Token
	// r.uid = loginResp.ID

	// if loginResp.Balance < 500 {
	// 	err = r.recharge(1000000) //充值1W
	// 	if err != nil {
	// 		r.Fsm.Event(context.Background(), EVENT_START)
	// 		return
	// 	}
	// }
	// r.Fsm.Event(context.Background(), EVENT_FREE)
}

func (r *ApiRobot) _idle_() {
	r.Fsm.Event(context.Background(), EVENT_ROOM) //暂时直接 进入 房间状态
}

func (r *ApiRobot) _room_() {
	room, err := r.getRoom(1 + uint(rand.Intn(4)))
	if err != nil {
		r.Fsm.Event(context.Background(), EVENT_FREE)
		return
	}
	r.room = room

	// logger.ZInfo("_room_", zap.Any("room", r.room))

	if err := r.syncOrders(uint(room.Setting.BetType)); err != nil {
		logger.ZInfo("syncOrders", zap.Error(err))
	}
	if err := r.syncPeriod(uint(room.Setting.BetType)); err != nil {
		logger.ZInfo("syncPeriod", zap.Error(err))
	}
	r.Fsm.Event(context.Background(), EVENT_START_BETTING)
}

func (r *ApiRobot) _betting_() {
	go r.StartCycles()

}

func (r *ApiRobot) tryBetting() {
	if r.room.State == game.STATE_BETTING {
		r.createOrder(uint(r.room.Setting.BetType), r.room.PeriodID, float64(rand.Intn(1000)), rand.Intn(9))
	}
}

func (r *ApiRobot) trySyncState() {

	resp, _ := r.syncState(uint(r.room.Setting.BetType), r.room.PeriodID, r.room.State)
	if resp != nil {
		structure.Copy(resp, r.room)
	}

}

// 仅在未启动的情况下启动定时循环
func (r *ApiRobot) StartCycles() {
	r.mu.Lock()
	if r.started {
		r.mu.Unlock()
		return // 如果已经开始就提前返回
	}
	r.started = true
	r.quit = make(chan struct{}) // 重新初始化quit通道
	r.mu.Unlock()

	go r.runCycles() // 启动runCycles goroutine来替代startCycles
}

func (r *ApiRobot) runCycles() {
	stateTicker := time.NewTicker(5 * time.Second)
	betTicker := time.NewTicker(1 * time.Second)
	defer stateTicker.Stop()
	defer betTicker.Stop()

	for {
		select {
		case <-stateTicker.C:
			r.trySyncState()
		case <-betTicker.C:
			r.tryBetting()
		case <-r.quit:
			r.mu.Lock()
			r.started = false
			r.mu.Unlock()
			return
		}
	}
}

// shutdown 安全地关闭 ApiRobot
func (r *ApiRobot) Shutdown() {
	r.mu.Lock()
	if r.started {
		close(r.quit) // 发送退出信号
		r.started = false
	}
	r.mu.Unlock()
}

func (r *ApiRobot) getUser(username string) (user *entities.User, err error) {
	r.logger("getUser", "")
	data := map[string]interface{}{
		"username": username,
	}
	var result interface{}
	result, err = r.request("/user/search-user", data)
	if err != nil {
		return
	}
	if result != nil {
		user = new(entities.User)
		err = mapToStruct(result, user)
	}
	if err != nil {
		return
	}
	return
}

func (r *ApiRobot) getRoom(betType uint) (resp *entities.WingoRoomResp, err error) {
	r.logger("getRoom", fmt.Sprintf("betType:%d", betType))
	data := map[string]interface{}{
		"betType": betType,
	}
	var result interface{}
	result, err = r.request("/wingo/get-room", data)
	if err != nil {
		return
	}
	if result != nil {
		resp = new(entities.WingoRoomResp)
		err = mapToStruct(result, resp)
	}
	if err != nil {
		return
	}
	return
}

func (r *ApiRobot) createOrder(betType uint, periodID string, betAmount float64, ticketNumber int) error {
	r.logger("createOrder", fmt.Sprintf("betType:%d,periodID:%s,betAmount:%0.2f,tickNumber:%d", betType, periodID, betAmount, ticketNumber))

	data := map[string]interface{}{
		"periodID":     periodID,
		"betType":      betType,
		"betAmount":    betAmount,
		"ticketNumber": ticketNumber,
	}
	_, err := r.request("/wingo/create-order", data)
	if err != nil {
		return err
	} else {
		// logger.ZInfo("createOrder", zap.String("username", r.username))
	}
	return nil
}

func (r *ApiRobot) syncState(betType uint, periodID string, state string) (resp *entities.StateResp, err error) {
	r.logger("syncState", fmt.Sprintf("betType:%d,periodID:%s,state:%s", betType, periodID, state))

	data := map[string]interface{}{
		"periodID": periodID,
		"betType":  betType,
		"state":    state,
	}
	var result interface{}
	result, err = r.request("/wingo/state-sync", data)
	if err != nil {
		return
	}

	if result != nil {
		resp = new(entities.StateResp)
		err = mapToStruct(result, resp)
	}

	return
}

func (r *ApiRobot) syncOrders(betType uint) error {
	r.logger("syncOrders", fmt.Sprintf("betType:%d", betType))
	data := map[string]interface{}{
		"page":     1,
		"pageSize": 20,
		"betType":  betType,
	}
	_, err := r.request("/wingo/recent-order-history-list", data)
	if err != nil {
		return err
	} else {

	}
	return nil
}

func (r *ApiRobot) syncPeriod(betType uint) error {
	r.logger("syncPeriod", fmt.Sprintf("betType:%d", betType))
	data := map[string]interface{}{
		"page":     1,
		"pageSize": 10,
		"betType":  betType,
	}
	_, err := r.request("/wingo/recent-period-history-list", data)
	if err != nil {
		return err
	} else {

	}
	return nil
}

func (r *ApiRobot) register() error {
	r.logger("register", fmt.Sprintf("pc:%s ", "1"))
	data := map[string]interface{}{
		"mobile":   "+91" + r.username,
		"username": r.username,
		"password": r.password,
		"verCode":  "9418",
		"pc":       "1",
		"isRobot":  1,
	}
	_, err := r.request("/user/register", data)
	if err != nil {
		return err
	}

	AppendCredentials(r.username, r.password)

	return nil
}

func (r *ApiRobot) recharge(amount float64) error {
	r.logger("recharge", fmt.Sprintf("amount:%0.2f", amount))

	data := entities.EditUserInfoReq{
		BalanceAdd: amount,
		UID:        r.uid,
	}
	timezone := "Asia/Shanghai"
	token := middleware.GenerateMD5Token(fmt.Sprintf("%d", 1), timezone, "julier@landing2023")
	_, err := r.request(fmt.Sprintf("/user/admin/edit-user?token=%s&uid=%d&timezone=%s", token, 1, timezone), data)
	if err != nil {
		return err
	}
	return nil
}

func (r *ApiRobot) login(username string, password string) (resp *entities.LoginResp, err error) {

	r.logger("login", fmt.Sprintf("username:%s login", r.username))

	data := map[string]interface{}{
		"username": username,
		"password": password,
	}
	var result interface{}
	result, err = r.request("/user/login", data)
	if err != nil {
		return
	}
	if result != nil {
		resp = new(entities.LoginResp)
		err = mapToStruct(result, resp)
	}
	if err != nil {
		return
	}
	return
}

// 请求数据
func (r *ApiRobot) request(endPoint string, data interface{}) (interface{}, error) {
	client := http.GetHttpClient()
	if r.token != "" {
		client.SetHeader("Authorization", "bearer "+r.token)
	}
	resp, err := client.SetHeader("Content-Type", "application/json").R().SetBody(data).Post(r.apiUrl + endPoint)
	if err != nil {
		return nil, err
	}

	// logger.ZInfo("enterSate", zap.String("endPoint", r.apiUrl+endPoint), zap.Any("req", data), zap.Any("result", resp), zap.Error(err))

	var result ginx.Resp
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return nil, err
	}
	if result.Code != 0 {
		return nil, errors.New(result.Msg)
	}
	return result.Data, nil
}

// AppendCredentials 追加账号和密码到一个文件
func AppendCredentials(account, password string) error {
	// 使用追加模式打开文件
	file, err := os.OpenFile("credentials.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	// 写入账号和密码到文件
	_, err = file.WriteString(account + " " + password + "\n")
	if err != nil {
		return err
	}

	return nil
}

// ReadCredentials 读取账号和密码列表
func ReadCredentials() ([]string, []string, error) {
	// 读取文件
	data, err := os.ReadFile("credentials.txt")
	if err != nil {
		return nil, nil, err
	}

	// 分割文件内容到行
	lines := strings.Split(string(data), "\n")

	// 创建账号和密码的切片
	accounts := make([]string, 0)
	passwords := make([]string, 0)

	// 遍历行进行处理
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) == 2 {
			accounts = append(accounts, parts[0])
			passwords = append(passwords, parts[1])
		}
	}

	return accounts, passwords, nil
}

// func (r *ApiRobot) genMobile() string {
// 	prefixes := []string{
// 		"130", "131", "132", "133", "134", "135", "136", "137", "138", "139",
// 		"150", "151", "152", "153", "155", "156", "157", "158", "159",
// 		"170", "171", "172", "173", "175", "176", "177", "178",
// 		"180", "181", "182", "183", "185", "186", "187", "188", "189",
// 	}
// 	// 随机选择一个前缀
// 	prefix := prefixes[rand.Intn(len(prefixes))]

// 	// 生成剩余的9位数字并拼接到前缀后面
// 	for i := 0; i < 8; i++ {
// 		prefix += fmt.Sprintf("%d", rand.Intn(10))
// 	}
// 	return prefix
// }
