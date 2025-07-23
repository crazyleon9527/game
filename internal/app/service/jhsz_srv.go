package service

import (
	"fmt"
	"log"
	"rk-api/internal/app/config"
	"rk-api/internal/app/constant"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/errors"
	"rk-api/internal/app/mq"
	"rk-api/internal/app/mq/handle"
	"rk-api/internal/app/service/repository"
	"rk-api/internal/app/utils"
	resty "rk-api/pkg/http"
	"rk-api/pkg/logger"
	"strconv"
	"strings"
	"time"

	"github.com/google/wire"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var JhszServiceSet = wire.NewSet(
	ProvideJhszService,
)

type JhszService struct {
	Repo       *repository.JhszRepository
	UserSrv    *UserService
	apiUrl     string
	appID      string
	appSecret  string
	signSecret string

	walletSrv *WalletService
	authSrv   *AuthService
}

func ProvideJhszService(repo *repository.JhszRepository,
	userSrv *UserService,
	walletSrv *WalletService,
	authSrv *AuthService,
) *JhszService {
	setting := config.Get().JhszSetting
	logger.ZInfo("ProvideJhszService", zap.Any("setting", setting))
	return &JhszService{
		Repo:       repo,
		UserSrv:    userSrv,
		apiUrl:     setting.ApiUrl,
		appID:      setting.AppID,
		appSecret:  setting.AppSecret,
		signSecret: setting.SignSecret,
		walletSrv:  walletSrv,
		authSrv:    authSrv,
	}
}

const (
	ErrorCode_Success       = 100001
	ErrorCode_VersionError  = 100004
	ErrorCode_ERR_Exception = 100000
)

func (s *JhszService) Login(req *entities.PlatLoginReq) (*entities.PlatLoginResp, error) {
	var signature = utils.GenerateSign(req.GetSignMap(), s.signSecret)
	if signature != req.Sign {
		return nil, errors.With("sign error")
	}

	user, err := s.authSrv.Login(&entities.LoginCredentials{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	loginResp := &entities.PlatLoginResp{
		Account:      s.toAccount(user.ID),
		Name:         user.Nickname,
		MerchantCode: s.appID,
		Sex:          fmt.Sprintf("%d", user.Gender),
	}
	return loginResp, nil
}

func (s *JhszService) Launch(req *entities.JhszGameLoginReq) (*entities.JhszGameLoginResp, error) {
	user, err := s.UserSrv.GetUserByUID(req.UID)
	if err != nil {
		return nil, err
	}
	account := s.toAccount(user.ID)

	var chainLaunchGameReq = &entities.ChainLaunchReq{
		Account:      account,
		Name:         user.Nickname,
		GameCode:     req.GameCode,
		Language:     req.Language,
		HomeLink:     req.HomeLink,
		Ip:           req.Ip,
		DeviceOS:     req.DeviceOS,
		DeviceId:     req.DeviceId,
		Sex:          fmt.Sprintf("%d", user.Gender),
		RoomId:       req.RoomId,
		MerchantCode: s.appID,
		Nonce:        time.Now().Format("20060102150405"),
	}

	chainLaunchGameReq.Signature = utils.GenerateSign(chainLaunchGameReq.ToMap(), s.signSecret)

	var chainLaunchResp = &entities.ChainLaunchResp{}

	client := resty.GetHttpClient()
	_, err = client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(chainLaunchGameReq).
		SetResult(chainLaunchResp).
		Post(s.apiUrl + "/ChainLaunch")

	logger.ZInfo("Launch", zap.Any("req", chainLaunchGameReq), zap.Any("resp", chainLaunchResp), zap.Error(err))

	if err != nil {
		return nil, errors.With("launch game error")
	}
	if chainLaunchResp.Error != ErrorCode_Success {
		return nil, errors.With(chainLaunchResp.Message)
	}
	return &entities.JhszGameLoginResp{GameLink: chainLaunchResp.GameUrl}, nil
}

func (s *JhszService) toAccount(uid uint) string {
	return fmt.Sprintf("JH_%d", uid)
}

func (s *JhszService) fromAccount(account string) (uint, error) {

	uidStr := strings.TrimPrefix(account, "JH_")

	// 将提取的 uid 字符串转换为 uint
	num, err := strconv.ParseUint(uidStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return uint(num), err
}

func (s *JhszService) SendNotification(req *entities.JhszNotificationReq) error {
	logger.Info("SendNotification", zap.Any("req", req))
	uid, err := s.fromAccount(req.Account)
	if err != nil {
		return errors.With("Invalid account")
	}
	user, err := s.UserSrv.GetUserByUID(uid)
	if user == nil || err != nil {
		return errors.With("plat user not exist")
	}
	SendNotification(user.ID, req.Message, req.Title)
	return nil
}

func (s *JhszService) GetAvailableFreeCard(req *entities.GetAvailableFreeCardReq) (*entities.JhszTransferResp, error) {

	var signature = utils.GenerateSign(req.GetSignMap(), s.signSecret)
	if signature != req.Sign {
		return nil, errors.With("sign error")
	}

	uid, err := s.fromAccount(req.Account)
	if err != nil {
		return nil, errors.With("Invalid account")
	}
	user, err := s.UserSrv.GetUserByUID(uid)
	if user == nil || err != nil {
		return nil, errors.With("user not exist")
	}
	return nil, err
}

func (s *JhszService) UseFreeCard(req *entities.UseFreeCardReq) (*entities.JhszTransferResp, error) {

	var signature = utils.GenerateSign(req.GetSignMap(), s.signSecret)
	if signature != req.Sign {
		return nil, errors.With("sign error")
	}

	uid, err := s.fromAccount(req.Account)
	if err != nil {
		return nil, errors.With("Invalid account")
	}
	user, err := s.UserSrv.GetUserByUID(uid)
	if user == nil || err != nil {
		return nil, errors.With("user not exist")
	}
	return nil, nil
	// return &entities.JhszTransferResp{Balance: wallet.Cash}, err
}

func (s *JhszService) Transfer(req *entities.JhszTransferReq) (*entities.JhszTransferResp, error) {

	var signature = utils.GenerateSign(req.GetSignMap(), s.signSecret)
	if signature != req.Sign {
		return nil, errors.With("sign error")
	}

	uid, err := s.fromAccount(req.Account)
	if err != nil {
		return nil, errors.With("Invalid account")
	}
	user, err := s.UserSrv.GetUserByUID(uid)
	if user == nil || err != nil {
		return nil, errors.With("user not exist")
	}

	wallet, err := s.walletSrv.GetUserWallet(user.ID)
	if err != nil {
		return nil, err
	}
	if wallet == nil {
		return nil, errors.With("user wallet not exist")
	}
	logger.ZInfo("JhszService-Transfer", zap.Any("req", req))

	err = s.walletSrv.HandleWallet(uid, func(wallet *entities.UserWallet, tx *gorm.DB) error {
		var action = req.Action

		switch action {
		case "withdraw":
			req.FlowType = constant.FLOW_TYPE_JHSZ_WITHDRAW
			transferAmount, err := s.Withdraw(wallet, req.Currency, req.Amount)
			if err != nil {
				return err
			}
			req.Amount = -transferAmount //提现-金额
		case "rollback":

		case "deposit":
			req.FlowType = constant.FLOW_TYPE_JHSZ_DEPOSIT

			transferAmount, err := s.Deposit(wallet, req.Currency, req.Amount)
			if err != nil {
				return err
			}
			req.Amount = transferAmount
		case "freeze":
			req.FlowType = constant.FLOW_TYPE_JHSZ_FREEZE
			transferAmount, err := s.Freeze(wallet, req.Currency, req.Amount)
			if err != nil {
				return err
			}
			req.Amount = -transferAmount

		case "unfreeze":
			req.FlowType = constant.FLOW_TYPE_JHSZ_UNFREEZE
			transferAmount, err := s.Unfreeze(wallet, req.Currency, req.Amount)
			if err != nil {
				return err
			}
			req.Amount = transferAmount
		default:
			return errors.With("Invalid action")
		}

		if err := s.walletSrv.UpdateCashWithTx(tx, wallet); err != nil {
			return nil
		}

		remark := fmt.Sprintf("JHSZ %s %s %s %s", req.Action, req.GameCode, req.RecordId, req.RoundId)

		createFlowQueue, _ := handle.NewCreateFlowQueue(&entities.Flow{
			UID:          user.ID,
			FlowType:     uint16(req.FlowType),
			Number:       req.Amount,
			Balance:      wallet.Cash,
			Remark:       remark,
			PromoterCode: user.PromoterCode,
		})
		if _, err := mq.MClient.Enqueue(createFlowQueue); err != nil {
			logger.ZError("createFlowQueue", zap.Any("flow", createFlowQueue), zap.Error(err))
		}

		return nil
	})

	return &entities.JhszTransferResp{Balance: wallet.Cash, TransferAmount: -req.Amount}, err
}

func (s *JhszService) Freeze(wallet *entities.UserWallet, currency string, amount float64) (transferAmount float64, err error) {
	logger.ZInfo("Freeze", zap.Float64("amount", amount), zap.Float64("wallet.cash", wallet.Cash))
	switch currency {
	case constant.CURRENCY_CASH:
		if wallet.Cash <= 0 {
			err = errors.With("insufficient cash")
			return
		}
		if wallet.Cash < amount { //不够就所有
			amount = wallet.Cash
		}
		wallet.SafeAdjustCash(-amount)
		transferAmount = amount
	default:
		err = errors.With("Invalid currency")
		return
	}
	return
}

func (s *JhszService) Unfreeze(wallet *entities.UserWallet, currency string, amount float64) (transferAmount float64, err error) {
	logger.ZInfo("Unfreeze", zap.Float64("amount", amount))
	switch currency {
	case constant.CURRENCY_CASH:
		wallet.SafeAdjustCash(amount)
		transferAmount = amount
	default:
		err = errors.With("Invalid currency")
		return
	}
	return
}

func (s *JhszService) Withdraw(wallet *entities.UserWallet, currency string, amount float64) (transferAmount float64, err error) {
	logger.ZInfo("Withdraw", zap.Float64("amount", amount))
	switch currency {
	case constant.CURRENCY_CASH:
		if wallet.Cash < amount {
			err = errors.With("insufficient cash")
			return
		}
		wallet.SafeAdjustCash(-amount)

		transferAmount = amount
	default:
		err = errors.With("Invalid currency")
	}
	return
}

func (s *JhszService) Deposit(wallet *entities.UserWallet, currency string, amount float64) (transferAmount float64, err error) {
	logger.ZInfo("Deposit", zap.Float64("amount", amount))
	switch currency {
	case constant.CURRENCY_CASH:
		wallet.SafeAdjustCash(amount)
		transferAmount = amount
	default:
		err = errors.With("Invalid currency")
		return
	}
	return
}

func (s *JhszService) FetchWallet(req *entities.JhszBalanceReq) (*entities.JhszTransferResp, error) {

	// var signature = utils.GenerateSign(req.GetSignMap(), s.signSecret)
	// if signature != req.Sign {

	// 	return nil, errors.With("sign error")
	// }

	uid, err := s.fromAccount(req.Account)
	if err != nil {
		return nil, errors.With("Invalid username")
	}

	wallet, err := s.walletSrv.GetUserWallet(uid)

	if wallet == nil || err != nil {
		return nil, errors.With("user not exist")
	}
	logger.ZInfo("fetchbalance", zap.Float64("balance", wallet.Cash))
	return &entities.JhszTransferResp{Balance: wallet.Cash}, nil
}

// //////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// 实现 FetchRecords 方法
func (s *JhszService) FetchRecords(startTime, endTime time.Time) ([]*entities.GameRecord, error) {

	logger.ZInfo("JhszService-FetchRecords", zap.Any("startTime", startTime), zap.Any("endTime", endTime))

	client := resty.GetHttpClient()

	// 时间格式化为 Unix 时间戳（毫秒）
	startTimestamp := startTime.UnixMilli()
	endTimestamp := endTime.UnixMilli()

	// 解析 JSON 数据
	var result struct {
		Code  int    `json:"code"`
		Msg   string `json:"msg"`
		Count int    `json:"count"`
		Data  []struct {
			RecordId string  `json:"RecordId"`
			Account  string  `json:"Account"`
			GameName string  `json:"GameName"`
			Score    float64 `json:"Score"`
			Tax      float64 `json:"Tax"`
			Time     string  `json:"Time"`
		} `json:"data"`
	}

	// 请求接口获取游戏记录
	_, err := client.R().
		SetQueryParam("limit", fmt.Sprintf("%d", 200)).
		SetQueryParam("merchantCode", s.appID).
		SetQueryParam("startDate", fmt.Sprintf("%d", startTimestamp)).
		SetQueryParam("endDate", fmt.Sprintf("%d", endTimestamp)).
		SetResult(&result).
		Get(s.apiUrl + "/LogPlayerGameTable")

	// 构造查询参数
	url := fmt.Sprintf("%s/LogPlayerGameTable?merchantCode=%s&startDate=%d&endDate=%d&limit=%d",
		s.apiUrl, s.appID, startTimestamp, endTimestamp, 200)

	// 打印请求 URL

	if err != nil {
		return nil, err
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("error fetching data: %s,%d", result.Msg, result.Code)
	}

	logger.ZInfo("JhszService-FetchRecords", zap.String("url", url), zap.Int("count", len(result.Data)))

	// 转换为 GameRecord 结构体
	var gameRecords []*entities.GameRecord
	for _, record := range result.Data {
		// 转换时间格式
		betTime, err := time.Parse("2006-01-02 15:04:05", record.Time)
		if err != nil {
			return nil, err
		}

		uid, err := s.fromAccount(record.Account)
		if err != nil {
			// logger.ZError("JhszService-FetchRecords", zap.String("account", record.Acount), zap.Error(err))
			continue
		}
		pc, err := s.UserSrv.GetUserPC(uid)
		if err != nil {
			continue
		}
		// var uid uint = 1 // 假设用户 ID 为 1

		gameRecord := &entities.GameRecord{
			BetTime:      betTime,
			BetAmount:    record.Score + record.Tax, // 假设投注金额是包含税费的金额
			Amount:       record.Score + record.Tax, // 假设有效流水是扣除税费后的金额
			Profit:       record.Score,              // 假设盈亏是扣除税费后的金额
			Game:         record.GameName,
			RecordId:     record.RecordId,
			Status:       1, // 假设已结算
			UID:          uid,
			Currency:     "CNY", // 假设使用人民币，可以根据实际情况调整
			PromoterCode: pc,
		}
		gameRecords = append(gameRecords, gameRecord)
	}

	// logger.ZInfo("JhszService-FetchRecords", zap.Int("count", len(gameRecords)))

	return gameRecords, nil
}

func (s *JhszService) FetchOnlineCount() ([]map[string]interface{}, error) {
	// 使用 resty 发送 HTTP 请求
	client := resty.GetHttpClient()

	// 构造请求 URL
	url := fmt.Sprintf("%s/GetOnlineCount?merchantCode=%s", s.apiUrl, s.appID)

	// 解析 JSON 数据
	var result struct {
		Code  int                      `json:"code"`
		Msg   string                   `json:"msg"`
		Count int                      `json:"count"`
		Data  []map[string]interface{} `json:"data"`
	}

	// 发送 GET 请求
	_, err := client.R().
		SetResult(&result).
		Get(url)

	if err != nil {
		return nil, fmt.Errorf("error fetching online count: %v", err)
	}

	// 判断返回代码
	if result.Code != 200 {
		return nil, fmt.Errorf("error fetching data: %s, %d", result.Msg, result.Code)
	}

	// 打印返回的 URL 和在线人数统计
	log.Printf("Fetched online count data: URL=%s, Count=%d", url, len(result.Data))

	// 返回在线人数统计
	return result.Data, nil
}

// 1740731328000
// 1740738428016
