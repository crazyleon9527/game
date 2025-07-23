package service

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"math"
	"rk-api/internal/app/config"
	"rk-api/internal/app/constant"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/errors"
	"rk-api/internal/app/mq"
	"rk-api/internal/app/mq/handle"
	"rk-api/internal/app/service/repository"
	resty "rk-api/pkg/http"
	"rk-api/pkg/logger"
	"rk-api/pkg/structure"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var ZfServiceSet = wire.NewSet(
	ProvideZfService,
)

type ZfService struct {
	Repo         *repository.ZfRepository
	UserSrv      *UserService
	WalletSrv    *WalletService
	apiUrl       string
	appID        string
	appSecret    string
	signSecret   string
	token        *string
	tokenTimeout time.Time
}

func ProvideZfService(repo *repository.ZfRepository,
	userSrv *UserService,
	walletSrv *WalletService,
) *ZfService {
	setting := config.Get().ZfSetting
	logger.ZInfo("ProvideZfService", zap.Any("setting", setting))
	return &ZfService{
		Repo:       repo,
		UserSrv:    userSrv,
		WalletSrv:  walletSrv,
		apiUrl:     setting.ApiUrl,
		appID:      setting.AppID,
		appSecret:  setting.AppSecret,
		signSecret: setting.SignSecret,
	}
}

func (s *ZfService) fetchBetRecords(fromDateTime, toDateTime string) ([]*entities.ZfBetRecord, error) {

	data := map[string]interface{}{
		"merchant_code": s.appID,
		"from":          fromDateTime,
		"to":            toDateTime,
	}

	resp, err := s.sendApi("/chain/query_game_history", data, false)
	if err != nil {
		return nil, err
	}

	type Detail struct {
		TotalPages           int                     `json:"total_pages"`
		CurrentPage          int                     `json:"current_page"`
		TotalRowsCurrentPage int                     `json:"total_rows_current_page"`
		GameHistory          []*entities.ZfBetRecord `json:"game_history"`
	}
	var detail Detail
	structure.MapToStruct(resp, &detail)

	return detail.GameHistory, nil
}

// 查询投注，用于返利
func (s *ZfService) QueryBetRecords() error {
	record, err := s.Repo.GetLastestZfBetRecord()
	if err != nil {
		return err
	}
	if record == nil {
		return errors.With("need insert a tag createdAt time record")
	}

	from := time.Unix(record.CreatedAt, 0) //需要插入一条数据启动开始
	to := from.Add(time.Hour)              //每隔一个小时
	if to.After(time.Now()) {
		return nil //时间未到
	}
	list, err := s.fetchBetRecords(from.Format(constant.DateLayout), from.Add(time.Hour).Format(constant.DateLayout))
	if err != nil {
		return err
	}
	return s.Repo.BatchCreateZfBetRecord(list)
}

const InvalidUsername = 6

func (s *ZfService) Launch(req *entities.ZfGameLoginReq) (interface{}, error) {
	_, err := s.QueryPlayer(req.UID) // 查询uid
	if err != nil {
		if appError, ok := err.(*errors.Error); ok { //调用错误
			if appError.Code != InvalidUsername { //不存在此用户
				return nil, err
			}
		}
		_, err := s.Register(req.UID) //没有玩家就注册
		if err != nil {
			return nil, err
		}
	}

	return s.Login(req)
}

func (s *ZfService) Login(req *entities.ZfGameLoginReq) (interface{}, error) {
	user, err := s.UserSrv.GetUserByUID(req.UID)
	if err != nil {
		return nil, err
	}
	username := s.toAccount(user.ID)
	data := map[string]interface{}{
		"merchant_code": s.appID,
		"game_code":     req.GameCode,
		"username":      username,
		"language":      "en-US",
		// "home_link":   "/index.html", // 如果还需要该参数，取消注释添加到data中
	}
	result, err := s.sendApi("/chain/query_game_launcher", data, false)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *ZfService) Register(uid uint) (interface{}, error) {

	user, err := s.UserSrv.GetUserByUID(uid)
	if err != nil {
		return nil, err
	}
	username := s.toAccount(user.ID)
	data := map[string]interface{}{
		"merchant_code": s.appID,
		"username":      username,
	}
	result, err := s.sendApi("/create_player", data, true)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type PlayerResp struct {
	ID       int    `json:"id"`
	Role     int    `json:"role"`
	Username string `json:"username"`
	Company  int    `json:"company"`
}

func (s *ZfService) QueryPlayer(uid uint) (*PlayerResp, error) {
	account := s.toAccount(uid)
	data := map[string]interface{}{
		"merchant_code": s.appID,
		"username":      account,
	}
	result, err := s.sendApi("/query_player", data, false)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	var resp PlayerResp
	structure.Copy(result, &resp)
	return &resp, nil
}

func (s *ZfService) Kick(req *entities.ZfKickReq) (interface{}, error) {
	account := s.toAccount(req.UID)

	data := map[string]interface{}{
		"merchant_code": s.appID,
		"game_code":     req.GameCode,
		"username":      account,
	}

	result, err := s.sendApi("/kick_player", data, true)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *ZfService) toAccount(uid uint) string {
	return fmt.Sprintf("%d", uid)
}

func (s *ZfService) fromAccount(account string) (uint, error) {
	num, err := strconv.ParseUint(account, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(num), err
}

type TokenResp struct {
	AuthToken string `json:"auth_token"`
	Timeout   int    `json:"timeout"`
}

func (s *ZfService) requestToken() (*TokenResp, error) {
	data := map[string]interface{}{
		"merchant_code": s.appID,
		"secure_key":    s.appSecret,
	}
	result, err := s.request("/generate_token", data, true)
	if err != nil {
		return nil, err
	}
	var resp TokenResp
	structure.MapToStruct(result, &resp)
	return &resp, nil
}

func (s *ZfService) sendApi(endPoint string, data map[string]interface{}, post bool) (interface{}, error) {
	if s.token == nil || time.Now().After(s.tokenTimeout) { //已经过期
		resp, err := s.requestToken()
		if err != nil {
			return nil, errors.With("request token fail")
		}
		s.token = &resp.AuthToken
		s.tokenTimeout = time.Now().Add(time.Duration(resp.Timeout-600) * time.Second) //小于600 会刷新
	}
	data["auth_token"] = *s.token
	return s.request(endPoint, data, post)
}

// 请求数据

type ApiResponse struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
	Detail  interface{} `json:"detail"`
	Message string      `json:"message"`
}

func (s *ZfService) request(endPoint string, data map[string]interface{}, post bool) (interface{}, error) {
	data["sign"] = s.sign(data) //sign
	var result map[string]interface{}
	var err error
	if post {

		result, err = resty.SendPost(resty.GetHttpClient(), s.apiUrl+endPoint, data, true)

		logger.ZInfo("request post", zap.String("endPoint", s.apiUrl+endPoint), zap.Any("data", data), zap.Any("result", result), zap.Error(err))
	} else {

		result, err = resty.SendGet(resty.GetHttpClient(), s.apiUrl+endPoint, data)

		logger.ZInfo("request get", zap.String("endPoint", s.apiUrl+endPoint), zap.Any("data", data), zap.Any("result", result), zap.Error(err))
	}
	if err != nil {
		return nil, err
	}

	// logger.ZInfo("recv request", zap.String("endPoint", s.apiUrl+endPoint), zap.Any("result", result))

	if result["success"].(bool) {
		return result["detail"], nil
	} else {
		return nil, &errors.Error{Message: result["message"].(string), Code: int(result["code"].(float64))}
	}
}

func (s *ZfService) ValidateSignature(data interface{}) bool {
	jsonMap, _ := structure.StructToMap(data)
	receivedSign, exsit := jsonMap["sign"]
	// logger.ZInfo("post ValidateSignature", zap.Any("data", data))
	// logger.Error("----------ValidateSignature----------", receivedSign, "|", s.sign(params))
	if !exsit {
		return false
	}
	return receivedSign == s.sign(jsonMap)
}

func (s *ZfService) sign(params map[string]interface{}) string {
	if len(params) == 0 {
		return ""
	}
	// 删除键为"sign"的参数
	delete(params, "sign")
	// 按键名排序
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接参数值
	var signStr string
	for _, k := range keys {
		val := params[k]
		switch v := val.(type) {
		case string:
			signStr += v
		case bool:
			signStr += strconv.FormatBool(v)
		case int, int64, int32, int16, int8, uint, uint64, uint32, uint16, uint8:
			signStr += fmt.Sprintf("%d", v)
		case float32:
			signStr += strconv.FormatFloat(float64(v), 'f', -1, 32)
		case float64:
			signStr += strconv.FormatFloat(v, 'f', -1, 64)
		default:
			// 对于其他非预期类型的处理
			signStr += fmt.Sprintf("%v", v)
		}
	}
	signStr += s.signSecret

	// SHA1哈希处理
	h := sha1.New()
	h.Write([]byte(signStr))
	bs := h.Sum(nil)
	// 转换成16进制字符串
	return hex.EncodeToString(bs)
}

func (s *ZfService) Bet(req *entities.ZfBetReq) gin.H {
	if !s.ValidateSignature(req) {
		return gin.H{"is_success": false, "err_msg": "sign error"}
	}

	uid, err := s.fromAccount(req.Username)
	if err != nil {
		return gin.H{"is_success": false, "msg": "Invalid username"}
	}
	wallet, err := s.WalletSrv.GetWallet(uid)
	if wallet == nil || err != nil {
		return gin.H{"is_success": false, "msg": "user not exist"}
	}

	if wallet.Cash < req.Amount {
		return gin.H{"is_success": false, "msg": "insufficient balance"}
	}

	order, err := s.Repo.GetOrderBy(wallet.UID, req.GameCode, req.RoundID, req.BetID)
	if err != nil {
		return gin.H{"is_success": false, "msg": err.Error()}
	}
	if order != nil {
		return gin.H{"is_success": false, "msg": fmt.Sprintf(" order exist uid =%d,game_code =%s,round_id =%d,bet_id= %d", uid, req.GameCode, req.RoundID, req.BetID)}
	}

	order = &entities.ZfTransferOrder{
		UID:      wallet.UID,
		RoundID:  req.RoundID,
		BetID:    req.BetID,
		GameCode: req.GameCode,
		Amount:   req.Amount,
	}

	err = s.WalletSrv.HandleWallet(uid, func(wallet *entities.UserWallet, tx *gorm.DB) error {
		if err := s.Repo.CreateOrderWithTx(tx, order); err != nil {
			return err
		}
		wallet.SafeAdjustCash(-req.Amount)

		if err := s.WalletSrv.UpdateCashWithTx(tx, wallet); err != nil {
			return err
		}

		remark := fmt.Sprintf("zfgame bet code:%s", req.GameCode)

		createFlowQueue, _ := handle.NewCreateFlowQueue(&entities.Flow{
			UID:          wallet.ID,
			FlowType:     constant.FLOW_TYPE_ZF_BET,
			Number:       -req.Amount,
			Balance:      wallet.Cash,
			Remark:       remark,
			PromoterCode: wallet.PromoterCode,
		})
		if _, err := mq.MClient.Enqueue(createFlowQueue); err != nil {
			logger.ZError("createFlowQueue", zap.Any("flow", createFlowQueue), zap.Error(err))
		}
		return nil
	})
	if err != nil {
		return gin.H{"is_success": false, "msg": err.Error()}
	}
	return gin.H{"is_success": true, "err_msg": "", "currency": "USD"}
}

func (s *ZfService) Payout(req *entities.ZfPayoutReq) gin.H {

	if !s.ValidateSignature(req) {
		return gin.H{"is_success": false, "err_msg": "sign error"}
	}

	uid, err := s.fromAccount(req.Username)
	if err != nil {
		return gin.H{"is_success": false, "msg": "Invalid username"}
	}
	wallet, err := s.WalletSrv.GetWallet(uid)
	if wallet == nil || err != nil {
		return gin.H{"is_success": false, "msg": "user not exist"}
	}

	order, err := s.Repo.GetOrderBy(wallet.UID, req.GameCode, req.RoundID, req.BetID)
	if err != nil {
		return gin.H{"is_success": false, "msg": err.Error()}
	}
	if order == nil {
		return gin.H{"is_success": false, "msg": fmt.Sprintf("not found order with uid =%d,game_code =%s,round_id =%d,bet_id= %d", uid, req.GameCode, req.RoundID, req.BetID)}
	}

	orderForUpdate := entities.ZfTransferOrder{
		RewardAmount: req.Amount,
		Status:       1,
	}
	orderForUpdate.ID = order.ID

	err = s.WalletSrv.HandleWallet(uid, func(wallet *entities.UserWallet, tx *gorm.DB) error {
		if err := s.Repo.UpdateOrderWithTx(tx, &orderForUpdate); err != nil {
			return err
		}
		wallet.SafeAdjustCash(req.Amount)

		if err := s.WalletSrv.UpdateCashWithTx(tx, wallet); err != nil {
			return err
		}
		tx.Commit()
		remark := fmt.Sprintf("zfgame reward code:%s", req.GameCode)
		createFlowQueue, _ := handle.NewCreateFlowQueue(&entities.Flow{
			UID:          wallet.ID,
			FlowType:     constant.FLOW_TYPE_ZF_PAYOUT,
			Number:       req.Amount,
			Balance:      wallet.Cash,
			Remark:       remark,
			PromoterCode: wallet.PromoterCode,
		})
		if _, err := mq.MClient.Enqueue(createFlowQueue); err != nil {
			logger.ZError("createFlowQueue", zap.Any("flow", createFlowQueue), zap.Error(err))
		}
		return nil

	})
	if err != nil {
		return gin.H{"is_success": false, "msg": err.Error()}
	}

	return gin.H{"is_success": true, "err_msg": "", "currency": "USD"}
}

func (s *ZfService) Refund(req *entities.ZfRefund) gin.H {

	if !s.ValidateSignature(req) {
		return gin.H{"is_success": false, "err_msg": "sign error"}
	}

	uid, err := s.fromAccount(req.Username)
	if err != nil {
		return gin.H{"is_success": false, "msg": "Invalid username"}
	}
	wallet, err := s.WalletSrv.GetUserWallet(uid)
	if wallet == nil || err != nil {
		return gin.H{"is_success": false, "msg": "user not exist"}
	}

	order, err := s.Repo.GetOrderBy(wallet.UID, req.GameCode, req.RoundID, req.BetID)
	if err != nil {
		return gin.H{"is_success": false, "msg": err.Error()}
	}
	if order == nil {
		return gin.H{"is_success": false, "msg": fmt.Sprintf("not found order with uid =%d,game_code =%s,round_id =%d,bet_id= %d", wallet.ID, req.GameCode, req.RoundID, req.BetID)}
	}
	if math.Abs(order.Amount-req.Amount) > constant.PreciseZero {
		return gin.H{"is_success": false, "msg": fmt.Sprintf("amount not equal record with uid =%d,game_code =%s,round_id =%d,bet_id= %d", wallet.ID, req.GameCode, req.RoundID, req.BetID)}
	}

	orderForUpdate := entities.ZfTransferOrder{
		Type:   req.Type,
		Status: 1,
	}
	orderForUpdate.ID = order.ID

	err = s.WalletSrv.HandleWallet(uid, func(wallet *entities.UserWallet, tx *gorm.DB) error {
		if err := s.Repo.UpdateOrderWithTx(tx, &orderForUpdate); err != nil {
			return err
		}
		wallet.SafeAdjustCash(req.Amount)

		if err := s.WalletSrv.UpdateCashWithTx(tx, wallet); err != nil {
			return err
		}

		var remark string
		if req.Type == 1 { // 1: refund 2: payout failed 3: issue cancel; 1:退回 2:派彩失败 3:取消
			req.FlowType = constant.FLOW_TYPE_ZF_REFUND
			remark = fmt.Sprintf("zfgame refund code:%s", req.GameCode)

		} else if req.Type == 2 {
			req.FlowType = constant.FLOW_TYPE_ZF_PAYOUT_FAIL
			remark = fmt.Sprintf("zfgame payout failed code:%s", req.GameCode)

		} else if req.Type == 3 {
			req.FlowType = constant.FLOW_TYPE_ZF_CANCEL
			remark = fmt.Sprintf("zfgame cancel code:%s", req.GameCode)
		}

		createFlowQueue, _ := handle.NewCreateFlowQueue(&entities.Flow{
			UID:          wallet.UID,
			FlowType:     uint16(req.FlowType),
			Number:       req.Amount,
			Balance:      wallet.Cash,
			Remark:       remark,
			PromoterCode: wallet.PromoterCode,
		})
		if _, err := mq.MClient.Enqueue(createFlowQueue); err != nil {
			logger.ZError("createFlowQueue", zap.Any("flow", createFlowQueue), zap.Error(err))
		}
		return nil
	})
	if err != nil {
		return gin.H{"is_success": false, "msg": err.Error()}
	}

	return gin.H{"is_success": true, "err_msg": "", "currency": "USD"}
}

func (s *ZfService) FetchBalance(req *entities.ZfBalanceReq) gin.H {

	if !s.ValidateSignature(req) {
		return gin.H{"is_success": false, "err_msg": "sign error"}
	}

	uid, err := s.fromAccount(req.Username)
	if err != nil {
		return gin.H{"is_success": false, "msg": "Invalid username"}
	}

	wallet, err := s.WalletSrv.GetWallet(uid)

	if wallet == nil || err != nil {
		return gin.H{"is_success": false, "msg": "user not exist"}
	}

	return gin.H{
		"is_success": true,
		"err_msg":    "",
		"currency":   "USD",
		"username":   req.Username,
		"balance":    wallet.Cash,
	}
}

func (s *ZfService) Settle(req *entities.ZfSettleReq) gin.H {
	if !s.ValidateSignature(req) {
		return gin.H{"is_success": false, "err_msg": "sign error"}
	}
	return gin.H{
		"is_success": true,
		"err_msg":    "",
	}
}
