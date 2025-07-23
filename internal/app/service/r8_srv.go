package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
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
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var R8ServiceSet = wire.NewSet(
	ProvideR8Service,
)

type R8Service struct {
	Repo      *repository.R8Repository
	UserSrv   *UserService
	WalletSrv *WalletService
	appID     string
	apiUrl    string
	appKey    string
}

func ProvideR8Service(repo *repository.R8Repository,
	userSrv *UserService,
	walletSrv *WalletService,
) *R8Service {
	setting := config.Get().R8Setting
	logger.ZInfo("ProvideR8Service", zap.Any("setting", setting))
	// 返回你的RechargeService实例
	return &R8Service{
		Repo:      repo,
		UserSrv:   userSrv,
		WalletSrv: walletSrv,
		appID:     setting.AppID,
		appKey:    setting.AppKey,
		apiUrl:    setting.ApiUrl,
	}
}

func (s *R8Service) fetchBetRecords(serverURL, apiKey, pfId, fromDateTime, toDateTime string) ([]*entities.R8BetRecord, error) {
	client := resty.GetHttpClient()

	resp, err := client.R().
		SetHeader("api_key", apiKey).
		SetHeader("pf_id", pfId).
		SetHeader("timestamp", fmt.Sprintf("%d", time.Now().UTC().Unix())).
		SetQueryParams(map[string]string{
			"from": fromDateTime,
			"to":   toDateTime,
		}).
		Get(serverURL + "/bet/records")

	// 错误处理
	if err != nil {
		log.Fatalf("Error on response.\n[ERROR] - %s", err)
		return nil, err
	}

	type ResponseData struct {
		Code int                     `json:"code"`
		Msg  string                  `json:"msg"`
		Data []*entities.R8BetRecord `json:"data"`
	}

	var responseData ResponseData
	err = json.Unmarshal(resp.Body(), &responseData)
	if err != nil {
		return nil, err
	}
	if err != nil {
		logger.Error("Error while decoding JSON response.\n[ERROR] - %s", err)
		return nil, err
	}
	return responseData.Data, nil
}

// 查询投注，用于返利
func (s *R8Service) QueryBetRecords() error {
	record, err := s.Repo.GetLastestR8BetRecord()
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
	list, err := s.fetchBetRecords(s.apiUrl, s.appKey, s.appID, from.Format(constant.DateLayout), from.Add(time.Hour).Format(constant.DateLayout))
	if err != nil {
		return err
	}
	return s.Repo.BatchCreateR8BetRecord(list)
}

func (s *R8Service) Login(req *entities.R8GameLoginReq) (interface{}, error) {
	user, err := s.UserSrv.GetUserByUID(req.UID)
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{
		"account": s.toAccount(user.ID), // 这里可能是 name 或 uid，视具体情况而定
		// "game_code": "UpDownV2",
		"game_code": req.GameCode,
		"lang":      "en-US",
	}
	if req.GameCode != "" {
		data["game_code"] = req.GameCode
	}

	return s.sendApiPost("/login", data)
}

func (s *R8Service) Kick(uid uint) (interface{}, error) {
	accountID := s.toAccount(uid)
	data := map[string]interface{}{
		"account": accountID, //
	}
	return s.sendApiPost(fmt.Sprintf("/logout/%s", accountID), data)
}

func (s *R8Service) FetchBalance(account string) gin.H {
	uid, err := s.fromAccount(account)
	if err != nil {
		return gin.H{
			"code": 22006,
			"msg":  "无效的用户ID",
		}
	}
	wallet, err := s.WalletSrv.GetWallet(uid)
	if wallet == nil || err != nil {
		return gin.H{
			"code": 22006,
			"msg":  "單一錢包玩家不存在",
		}
	}
	return gin.H{
		"code": 0,
		"msg":  "Success",
		"data": gin.H{
			"balance": wallet.Cash,
		},
	}
}

func (s *R8Service) GetSessionIDToken(apiKey, pfid, timestamp string) gin.H {
	// 生成API key的哈希值
	hasher := sha256.New()
	hasher.Write([]byte(pfid + s.appKey + timestamp))
	apiKeyHashed := hex.EncodeToString(hasher.Sum(nil))

	if apiKey != apiKeyHashed {
		return gin.H{
			"code": 11003,
			"msg":  "API key 驗證錯誤",
		}
	}

	return gin.H{
		"code": 0,
		"msg":  "Success",
		"data": gin.H{
			"sid": constant.AuthorizationFixKey,
		},
	}
}

func (s *R8Service) AwardActivity(req *entities.R8Activity) gin.H {
	uid, err := s.fromAccount(req.Account)
	if err != nil {
		return gin.H{"code": 22007, "msg": "Invalid username"}
	}

	wallet, err := s.WalletSrv.GetWallet(uid)
	if wallet == nil || err != nil {
		return gin.H{"code": 22006, "msg": "单一钱包玩家不存在"}
	}
	order, err := s.Repo.GetActivityOrderByAwardID(req.AwardID)
	if err != nil {
		return gin.H{"code": 22008, "msg": "單⼀錢包平台發⽣錯誤"}
	}
	if order != nil {
		return gin.H{"code": 22008, "msg": "奖励单已经存在"}
	}

	orderForUpdate := new(entities.R8ActivityOrder)
	structure.Copy(req, orderForUpdate)

	err = s.WalletSrv.HandleWallet(uid, func(wallet *entities.UserWallet, tx *gorm.DB) error {
		if err := s.Repo.CreateActivityOrderWithTx(tx, orderForUpdate); err != nil {
			return err
		}
		wallet.SafeAdjustCash(req.Money)

		if err := s.WalletSrv.UpdateCashWithTx(tx, wallet); err != nil {
			tx.Rollback()
			return err
		}

		remark := fmt.Sprintf("Rich88 %s %s", req.ActivityType, req.Action)

		createFlowQueue, _ := handle.NewCreateFlowQueue(&entities.Flow{
			UID:          wallet.ID,
			FlowType:     constant.FLOW_TYPE_R8_ACTIVITY_AWARD,
			Number:       req.Money,
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
		return gin.H{"code": 22008, "msg": "單⼀錢包平台發⽣錯誤"}
	}

	return gin.H{"code": 0, "msg": "Success"}

}

func (s *R8Service) Transfer(req *entities.R8Transfer) gin.H {

	uid, err := s.fromAccount(req.Account)
	if err != nil {
		return gin.H{"code": 22007, "msg": "Invalid username"}
	}

	wallet, err := s.WalletSrv.GetUserWallet(uid)
	if wallet == nil || err != nil {
		return gin.H{"code": 22006, "msg": "单一钱包玩家不存在"}
	}

	logger.ZInfo("r8 transfer", zap.Any("req", req))

	order, err := s.Repo.GetOrderByTransferNo(req.TransferNo, req.Action)
	if err != nil {
		logger.Error("----------------------------------err-----------------")
		return gin.H{"code": 22008, "msg": "單⼀錢包平台發⽣錯誤"}
	}
	if order != nil {
		logger.Error("----------------------------------exist-----------------")
		return gin.H{"code": 22008, "msg": "transferNo+action exist"}
	}

	if req.Action == "withdraw" {
		req.FlowType = constant.FLOW_TYPE_R8_WITHDRAW
		if req.GameCode == "PokDeng" {
			req.FlowType = constant.FLOW_TYPE_R8_WITHDRAW_POKDEN
		}
		if wallet.Cash-math.Abs(req.Money) < 0 {
			return gin.H{"code": 22007, "msg": "單⼀錢包玩家⾦錢不⾜"}
		}
		req.Money = -req.Money
	} else if req.Action == "rollback" {
		req.FlowType = constant.FLOW_TYPE_R8_ROLLBACK
	} else if req.Action == "deposit" {
		req.FlowType = constant.FLOW_TYPE_R8_DEPOSIT
	}

	orderForUpdate := entities.R8TransferOrder{
		UID:        wallet.UID,
		Amount:     req.Money,
		RoundID:    req.RoundID,
		RecordID:   req.RecordID,
		TransferNo: req.TransferNo,
		GameCode:   req.GameCode,
		Action:     req.Action,
	}

	err = s.WalletSrv.HandleWallet(uid, func(wallet *entities.UserWallet, tx *gorm.DB) error {

		if err := s.Repo.CreateOrderWithTx(tx, &orderForUpdate); err != nil {
			return err
		}

		wallet.SafeAdjustCash(req.Money)

		if err := s.WalletSrv.UpdateCashWithTx(tx, wallet); err != nil {
			return err
		}

		remark := fmt.Sprintf("Rich88 %s %s", req.GameCode, req.Action)

		createFlowQueue, _ := handle.NewCreateFlowQueue(&entities.Flow{
			UID:          wallet.ID,
			FlowType:     uint16(req.FlowType),
			Number:       req.Money,
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
		return gin.H{"code": 22008, "msg": "單⼀錢包平台發⽣錯誤"}
	}

	return gin.H{"code": 0, "msg": "Success"}
}

func (s *R8Service) toAccount(uid uint) string {
	return fmt.Sprintf("%d", uid)
}

func (s *R8Service) fromAccount(account string) (uint, error) {
	num, err := strconv.ParseUint(account, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(num), err
}

func (s *R8Service) sendApiPost(urlPath string, data map[string]interface{}) (interface{}, error) {
	client := resty.GetHttpClient()

	timestamp := time.Now().Unix()
	signStr := fmt.Sprintf("%s%s%d", s.appID, s.appKey, timestamp)
	h := sha256.New()
	h.Write([]byte(signStr))
	apiKey := hex.EncodeToString(h.Sum(nil))

	resp, err := client.R().
		SetHeaders(map[string]string{
			"Content-Type": "application/json",
			"Accept":       "application/json",
			"api_key":      apiKey,
			"pf_id":        s.appID,
			"timestamp":    fmt.Sprintf("%d", timestamp),
		}).
		SetBody(data).
		Post(s.apiUrl + urlPath)

	if err != nil {
		return nil, err
	}

	var apiResponse map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &apiResponse); err != nil {
		return nil, err
	}

	// 假设 logger.ZError 是您定义的日志记录函数
	logger.ZInfo("sendApiPost", zap.Any("req", data), zap.String("url", s.apiUrl+urlPath), zap.Any("apiResponse", apiResponse))

	if code, ok := apiResponse["code"].(int); ok && code != 0 {
		return nil, errors.With(apiResponse["msg"].(string)) // 修正了错误处理
	} else {
		return apiResponse["data"], nil
	}
}
