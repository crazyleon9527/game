package service

import (
	"fmt"
	"math"
	"math/rand"
	"rk-api/internal/app/constant"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/errors"
	"rk-api/internal/app/mq"
	"rk-api/internal/app/mq/handle"
	"rk-api/internal/app/pay"
	"rk-api/internal/app/service/repository"
	"rk-api/pkg/logger"
	"sync"
	"time"

	"github.com/google/wire"
	"github.com/orca-zhang/ecache"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	RECHARGE_CASH_MIN = 50.00 //最小提现金额
)

const (
	ACTIVITY_BIG_RECHARGE_CASH = 10000 // 活动大金额 充值
)

// 使用wire.Bind绑定RechargeService到RechargeInitializer
var RechargeServiceSet = wire.NewSet(
	ProvideRechargeService,
)

type RechargeService struct {
	Repo      *repository.RechargeRepository
	UserSrv   *UserService
	walletSrv *WalletService
	FlowSrv   *FlowService
	AgentSrv  *AgentService
	WalletSrv *WalletService
	// 在这里添加一个用户ID到Locker的映射, 这样每个用户都可以拥有自己的独立锁
	orderLockersMap sync.Map

	channelSettingCache *ecache.Cache

	rechargeImplMap map[string]pay.IPay
}

func ProvideRechargeService(repo *repository.RechargeRepository,
	userSrv *UserService,
	flowSrv *FlowService,

	walletSrv *WalletService,
	agentSrv *AgentService) *RechargeService {

	// 初始化你的缓存, 锁和其它实现
	// relationPIDCache := ecache.NewCache()
	rechargeImplMap := map[string]pay.IPay{
		"kb":   new(pay.KBPay),
		"tk":   new(pay.TKPay),
		"at":   new(pay.ATPay),
		"go":   new(pay.GOPay),
		"cow":  new(pay.COWPay),
		"ant":  new(pay.ANTPay),
		"dy":   new(pay.DYPay),
		"gaga": new(pay.GaGaPay),
	}
	channelSettingCache := ecache.NewLRUCache(1, 6, 10*time.Minute) //初始化缓存
	// 返回你的RechargeService实例
	return &RechargeService{
		Repo:                repo,
		UserSrv:             userSrv,
		FlowSrv:             flowSrv,
		AgentSrv:            agentSrv,
		walletSrv:           walletSrv,
		rechargeImplMap:     rechargeImplMap,
		channelSettingCache: channelSettingCache,
	}
}

// 生成订单号
func createRechargeOrderID(uid uint) string {
	// rand.Seed(time.Now().UnixNano()) //p1720561921
	return fmt.Sprintf("p%d%d%d", uid, time.Now().Unix(), rand.Intn(900000)+100000)
}

func (s *RechargeService) QueryAndUpdateRechargeChannelBalance() error {
	list, err := s.Repo.GetAvaliableRechargeSettingList()
	if err != nil {
		return err
	}
	for _, channel := range list {
		cfg, err := s.GetRechargeChannelSetting(channel.Name)
		if err != nil {
			logger.ZError("GetRechargeChannelSetting", zap.Error(err))
		}
		paymentParameters := &pay.PaymentParameters{
			MerNo:          cfg.AppID,
			PlatformApiUrl: cfg.BalanceApiUrl,
			AppKey:         cfg.WithdrawKey,
			AppSecret:      cfg.PaySecret,
		}
		pay, ok := s.rechargeImplMap[channel.Name]
		if ok {
			balanceResp, err := pay.QueryBalance(paymentParameters)
			logger.ZInfo("QueryBalance", zap.String("name", channel.Name), zap.Any("balanceResp", balanceResp))
			if err != nil {
				logger.ZError("QueryBalance", zap.String("name", channel.Name), zap.Error(err))
			} else {
				channel.AvailableAmount = balanceResp.AvailableAmount
				channel.BalanceAmount = balanceResp.BalanceAmount
				channel.FrozenAmount = balanceResp.FrozenAmount
				if channel.AvailableAmount == 0 {
					channel.AvailableAmount = constant.PreciseZero
					// channel.WithdrawState = 0是否关闭 //少于多少就自动关闭
				}
				if channel.BalanceAmount == 0 {
					channel.BalanceAmount = constant.PreciseZero
				}
				if channel.FrozenAmount == 0 {
					channel.FrozenAmount = constant.PreciseZero
				}
				s.Repo.UpdateRechargeSetting(channel)
			}

		}
	}
	return nil
}

func (s *RechargeService) GetRechargeOrderList(param *entities.GetRechargeOrderListReq) error {

	return s.Repo.GetRechargeOrderList(param)
}

func (s *RechargeService) GetRechargeChannelSetting(name string) (*entities.RechargeChannelSetting, error) {
	if val, ok := s.channelSettingCache.Get(name); ok {
		return val.(*entities.RechargeChannelSetting), nil
	}

	setting, err := s.Repo.GetRechargeChannelSetting(&entities.RechargeChannelSetting{Name: name})
	if err != nil || setting == nil {
		return nil, errors.WithCode(errors.RechargeChannelSettingNotExist)
	}
	s.channelSettingCache.Put(name, setting) //丢入缓存中 下次从缓存读取

	return setting, nil
}

func (s *RechargeService) GetRechargeConfig() (*entities.RechargeConfig, error) {
	list, err := s.Repo.GetRechargeSettingList()
	if err != nil {
		return nil, err
	}
	goods, err := s.Repo.GetRechargeGoodList()
	if err != nil {
		return nil, err
	}

	entity := new(entities.RechargeConfig)
	entity.Channels = list
	entity.Goods = goods
	minRecharge, err := s.Repo.GetMinRecharge()
	if err != nil {
		return nil, err
	}
	RECHARGE_CASH_MIN = minRecharge
	entity.MinRecharge = RECHARGE_CASH_MIN
	// logger.ZError("GetRechargeConfig", zap.Any("config", entity))
	return entity, nil
}

func (s *RechargeService) GetRechargeUrlInfo(param *entities.GetRechargeUrlReq) (*entities.RechargeUrlInfo, error) {
	if param.Cash < RECHARGE_CASH_MIN {
		return nil, errors.WithCode(errors.MinRechargeCashLimit)
	}
	config, err := s.Repo.GetRechargeSetting(&entities.RechargeSetting{RechargeState: 1, Name: param.Name})

	if err != nil {
		return nil, err
	}

	if config == nil {
		return nil, errors.WithCode(errors.RechargeConfigNotExist)
	}

	if config.Status != 1 {
		return nil, errors.WithCode(errors.RechargeConfigNotAvailable)
	}

	user, err := s.UserSrv.GetUserByUID(param.UID)
	if err != nil {
		return nil, err
	}
	hallPayInfo, err := s.GetRechargeChannelSetting(param.Name)
	if err != nil {
		return nil, err
	}

	if hallPayInfo == nil {
		return nil, errors.WithCode(errors.RechargeChannelSettingNotExist)
	}

	order := &entities.RechargeOrder{
		UID:          param.UID,
		OrderID:      createRechargeOrderID(param.UID),
		IP:           user.LoginIP,
		PromoterCode: user.PromoterCode,
		Price:        param.Cash,
		Count:        1,
		TotalAmount:  param.Cash, //price * count
		RechargeType: uint8(param.ActType),
		Channel:      config.Name,
		StartTime:    time.Now().Unix(),
	}

	if err := s.Repo.CreateRechargeOrder(order); err != nil {
		return nil, err
	}

	paymentParameters := &pay.PaymentParameters{
		MerNo:          hallPayInfo.AppID,
		Name:           user.Nickname,
		Email:          user.Email,
		Mobile:         user.Mobile,
		OrderAmount:    order.TotalAmount,
		PageURL:        hallPayInfo.PayReturnUrl,
		NotifyURL:      hallPayInfo.PayCallBackUrl,
		MerOrderNo:     order.OrderID,
		PlatformApiUrl: hallPayInfo.RechargeApiUrl,
		AppKey:         hallPayInfo.PayKey,
		AppSecret:      hallPayInfo.PaySecret,
		Currency:       "INR",
	}

	pay, ok := s.rechargeImplMap[hallPayInfo.Name]
	if !ok {
		return nil, errors.With(fmt.Sprintf("not init pay impl (%s)", hallPayInfo.Name))
	}

	paymentUrl, err := pay.RequestPaymentURL(paymentParameters.CheckCompatibility())
	if err != nil {
		return nil, err
	}

	urlInfo := &entities.RechargeUrlInfo{
		Url: paymentUrl.Url,
	}

	return urlInfo, nil
}

func CompareStringFloat(strVal float64, floatVal float64) bool {
	//将字符串转换为float64类型

	//比较两个浮点数
	if math.Abs(strVal-floatVal) < constant.PreciseOne { //根据实际情况选择适当的容差
		return true
	} else {
		return false
	}
}

// 充值处理
func (s *RechargeService) RechargeCallbackProcess(using pay.IPayBack) (resp string, err error) {

	transfer := using.GetTransferOrder()

	// 在此处获取订单的锁，如果不存在，创建一个新的锁。
	val, _ := s.orderLockersMap.LoadOrStore(transfer.GetMerOrderNo(), &sync.Mutex{})
	locker := val.(*sync.Mutex)

	// 锁定当前订单，只有锁定的goroutine可以执行以下操作
	locker.Lock()
	defer func() {
		locker.Unlock()
		s.orderLockersMap.Delete(transfer.GetMerOrderNo()) //删除 订单锁
	}()

	logger.ZInfo("RechargeCallbackProcess start ---------------------------------------------------------", zap.String("orderID", using.GetTransferOrder().OrderNo))
	defer func() {
		// 如果存在错误，就打印错误
		if err != nil {
			logger.ZError("RechargeCallbackProcess", zap.Any("transfer", using), zap.Error(err))
		}

		logger.ZInfo("RechargeCallbackProcess end ---------------------------------------------------------", zap.String("orderID", using.GetTransferOrder().OrderNo))
	}()

	var order *entities.RechargeOrder
	if order, err = s.Repo.GetRechargeOrder(&entities.RechargeOrder{OrderID: transfer.GetMerOrderNo()}); err != nil {
		return
	}
	if order == nil {
		err = errors.With("order not exist")
		return
	}
	order.TradeID = transfer.GetOrderNo()
	logger.ZInfo("RechargeCallbackProcess", zap.Any("order", order))
	if !using.IsTransactionSucc() {
		err = errors.With("Transaction failed")
		if order.Status != constant.RECHARGE_STATE_FAIL { // 事实上 应该是= 0
			orderForUpdate := &entities.RechargeOrder{
				FinishTime: time.Now().Unix(),
				TradeID:    order.TradeID,
				Status:     constant.RECHARGE_STATE_FAIL,
			}
			orderForUpdate.ID = order.ID

			if err := s.Repo.UpdateRechargeOrder(orderForUpdate); err != nil {
				logger.ZError("Transaction failed UpdateRechargeOrder", zap.Any("order", orderForUpdate), zap.Error(err))
			}
		}
		return
	}

	if order.Status != 0 {
		return using.GetSuccResp(), nil
	}

	equal := CompareStringFloat(transfer.GetOrderAmount(), order.TotalAmount)

	if !equal {
		err = errors.With("Amount does not match with the order price")
		return
	}

	if err = s.HandleBusiAfterTradeSucc(order); err != nil { //处理业务
		return
	}

	return using.GetSuccResp(), nil
}

// 交易成功后 处理业务
func (s *RechargeService) HandleBusiAfterTradeSucc(order *entities.RechargeOrder) error {

	logger.ZInfo("HandleBusiAfterTradeSucc UpdateUserWithTx", zap.Uint("uid", order.UID), zap.String("orderID", order.OrderID))

	err := s.WalletSrv.HandleWallet(order.UID, func(wallet *entities.UserWallet, tx *gorm.DB) error {

		orderForUpdate := &entities.RechargeOrder{
			FinishTime: time.Now().Unix(),
			TradeID:    order.TradeID,
			Status:     constant.RECHARGE_STATE_SUCC,
		}
		orderForUpdate.ID = order.ID

		if err := s.Repo.UpdateRechargeOrderWithTx(tx, orderForUpdate); err != nil {
			return err
		}

		wallet.SafeAdjustCash(order.TotalAmount)

		logger.ZInfo("HandleBusiAfterTradeSucc UpdateUserWithTx", zap.Uint("uid", wallet.ID), zap.Float64("cash", wallet.Cash))
		if err := s.WalletSrv.UpdateCashWithTx(tx, wallet); err != nil {
			return err
		}

		completedRecharge :=
			entities.CompletedRecharge{
				UID:          order.UID,
				PromoterCode: wallet.PromoterCode,
				OrderID:      order.OrderID,
				TradeID:      order.TradeID,
				Amount:       order.TotalAmount,
				Channel:      order.Channel,
				CreateTime:   time.Now().Unix(),
			}
		currentTime := time.Now()
		midnight := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location())
		completedRecharge.TodayTime = midnight.Unix()

		if err := s.Repo.CreateCompletedRechargeWithTx(tx, &completedRecharge); err != nil { //加入已完成充值表
			return err
		}

		createFlowQueue, _ := handle.NewCreateFlowQueue(&entities.Flow{
			UID:          order.UID,
			FlowType:     constant.FLOW_TYPE_RECHARGE_CASH,
			Number:       order.TotalAmount,
			Balance:      wallet.Cash,
			PromoterCode: wallet.PromoterCode,
		})

		if _, err := mq.MClient.Enqueue(createFlowQueue); err != nil {
			logger.ZError("createFlowQueue", zap.Any("flow", createFlowQueue), zap.Error(err))
		}

		return nil
	})
	if err != nil {
		return err
	}

	// if order.RechargeType == constant.RECHARGE_ORDER_ACT_TYPE_10000 { //是否是充值活动
	// 	if order.Price == 10000 && user.FirstTen <= 0 { //充值一万送2000

	// 		user.AddBalance(2000)
	// 		userForUpdate := &entities.User{
	// 			Balance:  user.Balance,
	// 			FirstTen: 1,
	// 		}
	// 		userForUpdate.ID = user.ID
	// 		if err := s.UserSrv.UpdateUser(userForUpdate); err != nil {
	// 			logger.ZError("RECHARGE_ORDER_ACT_TYPE_10000 UpdateUser", zap.Any("userForUpdate", userForUpdate), zap.Error(err))
	// 		}
	// 		createFlowQueue, _ := handle.NewCreateFlowQueue(&entities.Flow{
	// 			UID:          order.UID,
	// 			FlowType:     constant.FLOW_TYPE_RECHARGE_ACT_10000,
	// 			Number:       2000,
	// 			Balance:      user.Balance,
	// 			PromoterCode: user.PromoterCode,
	// 		})
	// 		if _, err := mq.MClient.Enqueue(createFlowQueue); err != nil {
	// 			logger.ZError("createFlowQueue", zap.Any("flow", createFlowQueue), zap.Error(err))
	// 		}

	// 	}
	// } else {
	// 	if err := s.AgentSrv.CheckReturnRechargeCash(user.ID, order.TotalAmount, order.OrderID); err != nil { //返利
	// 		logger.ZError("FinalizeReturnRechargeCash",
	// 			zap.Uint("uid", user.ID),
	// 			zap.Float64("cash", order.TotalAmount),
	// 			zap.String("order_id", order.OrderID),
	// 			zap.Error(err),
	// 		)
	// 	}
	// }

	return nil
}
