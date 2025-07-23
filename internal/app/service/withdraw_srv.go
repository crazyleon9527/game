package service

import (
	"fmt"
	"math/rand"
	"rk-api/internal/app/config"
	"rk-api/internal/app/constant"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/errors"
	"rk-api/internal/app/mq"
	"rk-api/internal/app/mq/handle"
	"rk-api/internal/app/pay"
	"rk-api/internal/app/service/repository"
	"rk-api/pkg/cjson"
	"rk-api/pkg/logger"
	"rk-api/pkg/structure"
	"sync"
	"time"

	"github.com/google/wire"
	"github.com/orca-zhang/ecache"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	WITHDRAW_CASH_MIN = 50.00 //最小提现金额

	WITHDRAW_RATE = 5 //抽水 5%
)

// 使用wire.Bind绑定WithdrawService到WithdrawInitializer
var WithdrawServiceSet = wire.NewSet(
	ProvideWithdrawService,
)

type WithdrawService struct {
	Repo      *repository.WithdrawRepository
	UserSrv   *UserService
	AuthSrv   *AuthService
	FlowSrv   *FlowService
	WalletSrv *WalletService
	VerifySrv *VerifyService

	// 在这里添加一个用户ID到Locker的映射, 这样每个用户都可以拥有自己的独立锁
	orderLockersMap     sync.Map
	withdrawReviewMutex sync.Mutex //提现审核

	channelSettingCache *ecache.Cache

	withdrawImplMap map[string]pay.IWithdraw
}

func ProvideWithdrawService(repo *repository.WithdrawRepository, userSrv *UserService, flowSrv *FlowService, fundSrv *WalletService, VerifySrv *VerifyService) *WithdrawService {
	// 初始化你的缓存, 锁和其它实现
	withdrawImplMap := map[string]pay.IWithdraw{
		"kb":   new(pay.KBWithdraw),
		"tk":   new(pay.TKWithdraw),
		"at":   new(pay.ATWithdraw),
		"go":   new(pay.GOWithdraw),
		"cow":  new(pay.COWWithdraw),
		"ant":  new(pay.ANTWithdraw),
		"dy":   new(pay.DYWithdraw),
		"gaga": new(pay.GaGaWithdraw),
	}
	channelSettingCache := ecache.NewLRUCache(1, 6, 10*time.Minute) //初始化缓存
	// 返回你的WithdrawService实例
	return &WithdrawService{
		Repo:                repo,
		UserSrv:             userSrv,
		FlowSrv:             flowSrv,
		VerifySrv:           VerifySrv,
		WalletSrv:           fundSrv,
		withdrawImplMap:     withdrawImplMap,
		channelSettingCache: channelSettingCache,
	}
}

// 生成订单号
func createWithdrawOrderID(uid uint) string {
	// rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("w%d%d%d", uid, time.Now().Unix(), rand.Intn(900000)+100000)
}

// 审核提现
func (s *WithdrawService) ReviewWithdrawal(req *entities.ReviewWithdrawalReq) (err error) {
	if req.OptType == constant.WITHDRAW_REVIEW_OPT_REJECT {
		return s.RejectWithdrawal(req)
	} else if req.OptType == constant.WITHDRAW_REVIEW_OPT_APPROVE {
		return s.ApproveWithdrawal(req)
	} else if req.OptType == constant.WITHDRAW_REVIEW_OPT_RESERVE_FAILED {
		return s.ReverseWithdrawalFailed(req)
	}
	return nil
}

// 拒绝 提现申请,退回提现金额
func (s *WithdrawService) RejectWithdrawal(req *entities.ReviewWithdrawalReq) (err error) {
	s.withdrawReviewMutex.Lock() //互斥锁
	defer s.withdrawReviewMutex.Unlock()

	var record *entities.HallWithdrawRecord
	record, err = s.Repo.GetWithdrawCardRecordByID(req.ID)
	if err != nil {
		return
	}

	if record.Status != constant.WITHDRAW_STATE_WAIT_REVIEW && record.Status != constant.WITHDRAW_STATE_TRADE_FAIL {
		return errors.WithCode(errors.InvalidWithdrawalReview) //The order is not in a review status.
	}

	defer func() {
		log := entities.SystemOptionLog{
			Type:     constant.SYS_OPTION_TYPE_WITHDRAWAL_REVIEW,
			OptionID: req.SysUID,
			IP:       req.IP,
			Content:  cjson.StringifyIgnore(record),
		}
		if err != nil {
			log.Result = "false"
			log.Remark = "运营审核退回失败"
			logger.ZError("ReviewWithdrawal fail", zap.Any("req", req), zap.Error(err))
		} else {
			log.Result = "true"
			log.Remark = "运营审核退回成功"
		}

		optionLogQueue, _ := handle.NewOptionLogQueue(&log)
		if _, err := mq.MClient.Enqueue(optionLogQueue); err != nil {
			logger.ZError("optionLogQueue", zap.Any("log", &log), zap.Error(err))
		}
	}()

	logger.ZInfo("RejectWithdrawal", zap.Uint("id", record.ID),
		zap.String("orderID", record.OrderID),
		zap.Uint8("status", record.Status),
		zap.Float64("cash", record.Cash),
		zap.Uint("uid", record.UID))

	err = s.WalletSrv.HandleWallet(record.UID, func(wallet *entities.UserWallet, tx *gorm.DB) error {
		recordForUpdate := entities.HallWithdrawRecord{
			Status:    constant.WITHDRAW_STATE_REJECTED,
			Reason:    req.Reason,
			CheckUser: req.SysUID,
			CheckTime: time.Now().Unix(),
		}
		recordForUpdate.ID = record.ID

		wallet.SafeAdjustCash(record.Cash)
		// user.AddLockCash(-record.Cash) //锁定金额减少

		if err = s.Repo.UpdateHallWithdrawRecordWithTx(tx, &recordForUpdate); err != nil {
			return err
		}

		logger.ZInfo("RejectWithdrawal UpdateUserWithTx", zap.Uint("uid", wallet.ID), zap.Float64("balance", wallet.Cash))
		if err = s.WalletSrv.UpdateCashWithTx(tx, wallet); err != nil {
			return err
		}
		createFlowQueue, _ := handle.NewCreateFlowQueue(&entities.Flow{
			UID:          wallet.ID,
			FlowType:     constant.FLOW_TYPE_WITHDRAW_LOCK_CASH,
			Number:       record.Cash,
			Balance:      wallet.Cash,
			PromoterCode: wallet.PromoterCode,
		})
		if _, err := mq.MClient.Enqueue(createFlowQueue); err != nil {
			logger.ZError("createFlowQueue", zap.Any("flow", createFlowQueue), zap.Error(err))
		}
		return nil
	})

	return err
}

func (s *WithdrawService) GetRechargeChannelSetting(name string) (*entities.RechargeChannelSetting, error) {
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

// 批准提现
func (s *WithdrawService) ApproveWithdrawal(req *entities.ReviewWithdrawalReq) (err error) {
	s.withdrawReviewMutex.Lock() //互斥锁
	defer s.withdrawReviewMutex.Unlock()
	var record *entities.HallWithdrawRecord
	record, err = s.Repo.GetWithdrawCardRecordByID(req.ID)
	if err != nil {
		return
	}
	if record.Status != constant.WITHDRAW_STATE_WAIT_REVIEW {
		return errors.WithCode(errors.InvalidWithdrawalReview) //The order is not in a review status.
	}

	defer func() {
		log := entities.SystemOptionLog{
			Type:     constant.SYS_OPTION_TYPE_WITHDRAWAL_REVIEW,
			OptionID: req.SysUID,
			IP:       req.IP,
			Content:  cjson.StringifyIgnore(record),
		}
		if err != nil {
			log.Result = "false"
			log.Remark = "运营审核打款提交失败"

			// 打印错误，使用你的日志库方法来记录
			logger.ZError("ApproveWithdrawal fail", zap.Any("req", req), zap.Error(err))
		} else {
			log.Result = "true"
			log.Remark = "运营审核打款提交成功"
			logger.ZError("ApproveWithdrawal succ", zap.Any("req", req))
		}
		optionLogQueue, _ := handle.NewOptionLogQueue(&log)
		if _, err := mq.MClient.Enqueue(optionLogQueue); err != nil {
			logger.ZError("optionLogQueue", zap.Any("log", &log), zap.Error(err))
		}
	}()

	var user *entities.User
	user, err = s.UserSrv.GetUserByUID(record.UID)
	if err != nil {
		return
	}
	var card *entities.WithdrawCard
	card, err = s.Repo.GetWithdrawCard(&entities.WithdrawCard{
		UID:    record.UID,
		Status: 1,
		Active: 1,
	})
	if err != nil {
		return
	}
	if card == nil {
		err = errors.WithCode(errors.WithdrawCardNotExist)
		return
	}

	if card.AccountNumber != record.AccountNumber { //当用户临时更换了可使用卡 则修改提现记录卡的记录
		record.AccountName = card.Name
		record.IFSC = card.IFSC
		record.AccountNumber = card.AccountNumber
	}

	config, err := s.Repo.GetRechargeSetting(&entities.RechargeSetting{WithdrawState: 1, Status: 1})

	if err != nil || config == nil {
		err = errors.WithCode(errors.RechargeConfigNotAvailable)
		return
	}
	var hallPayInfo *entities.RechargeChannelSetting
	hallPayInfo, err = s.GetRechargeChannelSetting(config.Name)
	if err != nil {
		return
	}

	withdraw, ok := s.withdrawImplMap[config.Name]
	if !ok {
		err = errors.With(fmt.Sprintf("not init withdraw impl (%s)", config.Name))
		return
	}
	record.Channel = config.Name                                   //设置渠道
	record.Status = constant.WITHDRAW_STATE_REVIEWED               //审核通过
	if err = s.Repo.UpdateHallWithdrawRecord(record); err != nil { //先把状态设置
		return
	}

	parameters := pay.WithdrawParameters{
		MerNo:          hallPayInfo.AppID,
		AccountNumber:  record.AccountNumber,
		IFSC:           record.IFSC,
		AccountName:    record.AccountName,
		Name:           user.Nickname,
		Email:          user.Email,
		Mobile:         user.Mobile,
		IP:             req.IP, //req.IP
		OrderAmount:    record.RealCash,
		PageURL:        hallPayInfo.WithdrawReturnUrl,
		NotifyURL:      hallPayInfo.WithdrawCallBackUrl,
		MerOrderNo:     record.OrderID,
		PlatformApiUrl: hallPayInfo.WithdrawApiUrl,
		AppKey:         hallPayInfo.WithdrawKey,
		AppSecret:      hallPayInfo.PaySecret,
		Currency:       "INR",
	}
	// logger.Error("-----------------------------------------------", hallPayInfo.WithdrawApiUrl)
	parameters.CheckCompatibility()

	withdrawResp, err := withdraw.RequestWithdraw(&parameters)

	logger.ZInfo("RequestWithdraw", zap.Any("resp", withdrawResp))

	if err != nil {
		record.Status = constant.WITHDRAW_STATE_TRADE_FAIL              //审核通过 ,打款失败
		if err := s.Repo.UpdateHallWithdrawRecord(record); err != nil { //先把状态设置
			return err
		}
		return err
	}

	return nil
}

// 已经成功的，而渠道反馈并不成功 失败冲正
func (s *WithdrawService) ReverseWithdrawalFailed(req *entities.ReviewWithdrawalReq) (err error) {
	s.withdrawReviewMutex.Lock() //互斥锁
	defer s.withdrawReviewMutex.Unlock()

	var record *entities.HallWithdrawRecord
	record, err = s.Repo.GetWithdrawCardRecordByID(req.ID)
	if err != nil {
		return
	}

	if record.Status != constant.WITHDRAW_STATE_TRADE_SUCC {
		return errors.WithCode(errors.InvalidWithdrawalReview) //The order is not in a review status.
	}

	defer func() {
		log := entities.SystemOptionLog{
			Type:     constant.SYS_OPTION_TYPE_WITHDRAWAL_REVIEW,
			OptionID: req.SysUID,
			IP:       req.IP,
			Content:  cjson.StringifyIgnore(record),
		}
		if err != nil {
			log.Result = "false"
			log.Remark = "运营审核失败冲正失败"
			logger.ZError("reverseWithdrawalFailed fail", zap.Any("req", req), zap.Error(err))
		} else {
			log.Result = "true"
			log.Remark = "运营审核失败冲正成功"
		}

		optionLogQueue, _ := handle.NewOptionLogQueue(&log)
		if _, err := mq.MClient.Enqueue(optionLogQueue); err != nil {
			logger.ZError("optionLogQueue", zap.Any("log", &log), zap.Error(err))
		}
	}()
	var fundFreeze *entities.FundFreeze
	fundFreeze, err = s.WalletSrv.GetFundFreeze(&entities.FundFreeze{RecordID: record.OrderID})
	if err != nil {
		return
	}
	err = s.WalletSrv.HandleWallet(record.UID, func(wallet *entities.UserWallet, tx *gorm.DB) error {
		logger.ZInfo("ReverseWithdrawalFailed", zap.Uint("id", record.ID),
			zap.String("orderID", record.OrderID),
			zap.Uint8("status", record.Status),
			zap.Float64("cash", record.Cash),
			zap.Uint("uid", record.UID))

		recordForUpdate := entities.HallWithdrawRecord{
			Status:    constant.WITHDRAW_STATE_TRADE_FAIL,
			Reason:    req.Reason,
			CheckUser: req.SysUID,
			CheckTime: time.Now().Unix(),
		}
		recordForUpdate.ID = record.ID

		if err = s.Repo.UpdateHallWithdrawRecordWithTx(tx, &recordForUpdate); err != nil {
			return err
		}

		fundFreeze.Status = constant.FUND_STATUS_FREEZE //继续冻结
		if err := s.WalletSrv.UpdateFundFreezeWithTx(tx, fundFreeze); err != nil {
			return err
		}
		return nil
	})

	return err
}

func (s *WithdrawService) GetWithdrawDetail(uid uint) (*entities.WithdrawDetail, error) {

	withdrawDetail := new(entities.WithdrawDetail)
	cards, err := s.Repo.GetWithdrawCardListByUID(uid)
	if err != nil {
		return nil, err
	}

	minRecharge, err := s.Repo.GetMinWithdraw()
	if err != nil {
		return nil, err
	}
	WITHDRAW_CASH_MIN = minRecharge

	withdrawDetail.Cards = cards

	withdrawDetail.MinWithdraw = WITHDRAW_CASH_MIN

	wallet, err := s.WalletSrv.GetUserWallet(uid)
	if err != nil {
		return nil, err
	}

	withdrawDetail.MaxWithdraw = wallet.Cash
	withdrawDetail.WithdrawCash = wallet.Cash

	return withdrawDetail, nil
}

func (s *WithdrawService) GetWithdrawCardListByUID(UID uint) ([]*entities.WithdrawCard, error) {
	return s.Repo.GetWithdrawCardListByUID(UID)
}

func (s *WithdrawService) GetHallWithdrawRecordList(param *entities.GetHallWithdrawRecordListReq) error {

	return s.Repo.GetHallWithdrawRecordList(param)
}

func (s *WithdrawService) AddUserWithdrawCard(req *entities.AddUserWithdrawCardReq) (err error) {
	user, err := s.UserSrv.GetUserByUID(req.UID)
	if err != nil {
		logger.Info("----------------------------1--2-----------------", req.UID)
		return err
	}
	card := new(entities.AddWithdrawCardReq)
	structure.Copy(&req, card)

	defer func() {
		log := entities.SystemOptionLog{
			Type:     constant.SYS_OPTION_TYPE_WITHDRAWAL_CARD_ADD,
			OptionID: req.OptionID,
			IP:       req.IP,
			Content:  cjson.StringifyIgnore(card),
		}
		if err != nil {
			log.Result = "false"
			log.Remark = "运营添加用户卡失败"
			logger.ZError("AddUserWithdrawCard fail", zap.Any("req", req), zap.Error(err))
		} else {
			log.Result = "true"
			log.Remark = "运营添加用户卡成功"
		}

		optionLogQueue, _ := handle.NewOptionLogQueue(&log)
		if _, err := mq.MClient.Enqueue(optionLogQueue); err != nil {
			logger.ZError("optionLogQueue", zap.Any("log", &log), zap.Error(err))
		}
	}()
	card.Mobile = user.Mobile
	card.VerCode = config.Get().ServiceSettings.TrustedUserCode
	err = s.AddWithdrawCard(card)
	return
}

func (s *WithdrawService) DelUserWithdrawCard(req *entities.DelUserWithdrawCardReq) (err error) {

	card, err := s.Repo.GetWithdrawCardByID(req.ID)
	if err != nil {
		return err
	}
	if card == nil {
		return errors.WithCode(errors.WithdrawCardNotExist)
	}

	defer func() {
		log := entities.SystemOptionLog{
			Type:     constant.SYS_OPTION_TYPE_WITHDRAWAL_CARD_ADD,
			OptionID: req.OptionID,
			IP:       req.IP,
			Content:  cjson.StringifyIgnore(card),
		}
		if err != nil {
			log.Result = "false"
			log.Remark = "运营添加用户卡失败"
			logger.ZError("AddUserWithdrawCard fail", zap.Any("req", req), zap.Error(err))
		} else {
			log.Result = "true"
			log.Remark = "运营添加用户卡成功"
		}

		optionLogQueue, _ := handle.NewOptionLogQueue(&log)
		if _, err := mq.MClient.Enqueue(optionLogQueue); err != nil {
			logger.ZError("optionLogQueue", zap.Any("log", &log), zap.Error(err))
		}
	}()

	err = s.DelWithdrawCardByID(card.UID, req.ID)

	return
}

func (s *WithdrawService) FixUserWithdrawCard(req *entities.FixUserWithdrawCardReq) (err error) {

	card, err := s.Repo.GetWithdrawCardByID(req.ID)
	if err != nil {
		return err
	}
	if card == nil {
		return errors.WithCode(errors.WithdrawCardNotExist)
	}

	defer func() {
		log := entities.SystemOptionLog{
			Type:     constant.SYS_OPTION_TYPE_WITHDRAWAL_CARD_ADD,
			OptionID: req.OptionID,
			IP:       req.IP,
			Content:  cjson.StringifyIgnore(card),
		}
		if err != nil {
			log.Result = "false"
			log.Remark = "运营修改用户卡失败"
			logger.ZError("AddUserWithdrawCard fail", zap.Any("req", req), zap.Error(err))
		} else {
			log.Result = "true"
			log.Remark = "运营修改用户卡成功"
		}

		optionLogQueue, _ := handle.NewOptionLogQueue(&log)
		if _, err := mq.MClient.Enqueue(optionLogQueue); err != nil {
			logger.ZError("optionLogQueue", zap.Any("log", &log), zap.Error(err))
		}
	}()

	cardForUpdate := new(entities.WithdrawCard)
	cardForUpdate.ID = req.ID
	if req.AccountNumber != nil {
		cardForUpdate.AccountNumber = *req.AccountNumber
	}
	if req.IFSC != nil {
		cardForUpdate.IFSC = *req.IFSC
	}
	if req.Name != nil {
		cardForUpdate.Name = *req.Name
	}

	err = s.Repo.UpdateWithdrawCard(cardForUpdate)

	s.SelectWithdrawCard(card.UID, card.ID) //修改后 自动选为 可用卡

	logger.ZInfo("FixUserWithdrawCard", zap.Any("cardForUpdate", cardForUpdate), zap.Error(err))
	return
}

func (s *WithdrawService) AddWithdrawCard(req *entities.AddWithdrawCardReq) error {

	user, err := s.UserSrv.GetUserByUID(req.UID)
	if err != nil {
		return err
	}
	if user.Mobile != "" && user.Mobile != req.Mobile {
		return errors.With("mobile should be register mobible")
	}

	// if req.VerCode == config.Get().ServiceSettings.TrustedUserCode {
	// 	return errors.With("Please retrieve the verification code again")
	// }

	if err := s.VerifySrv.CheckVerifyCode(req.Mobile, req.VerCode); err != nil { //检测验证码
		return err
	}

	count, err := s.Repo.CountWithdrawCardsByUID(req.UID)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.With("Each user can link only one card,please contact customer service to make changes.")
	}

	card := new(entities.WithdrawCard)
	structure.Copy(&req, card)

	if card.AccountNumber != "" {
		existCard, err := s.Repo.GetWithdrawCard(&entities.WithdrawCard{
			AccountNumber: card.AccountNumber,
		})
		if err != nil {
			return err
		}
		if existCard != nil {
			return errors.WithCode(errors.WithdralCardAccountNumberExist) //提现卡已经存在
		}
	}

	card.Status = 1
	if err := s.Repo.CreateWithdrawCard(card); err != nil {
		return err
	}
	return s.SelectWithdrawCard(card.UID, card.ID) //选择新添加卡为可用卡
}

func (s *WithdrawService) DelWithdrawCardByID(uid uint, id uint) error {
	return s.Repo.DelWithdrawCardByID(uid, id)
}

// 选择 一个用户只有一张使用的卡,所以要将之前的取消
func (s *WithdrawService) SelectWithdrawCard(uid uint, id uint) error {
	return s.Repo.SelectWithdrawCard(uid, id)
}

// 申请提现
func (s *WithdrawService) ApplyForWithdrawal(param *entities.ApplyForWithdrawalReq) error {

	if param.Cash < WITHDRAW_CASH_MIN { //最低提现金额
		return errors.WithCode(errors.MinWithdrawalCashLimit)
	}

	user, err := s.UserSrv.GetUserByUID(param.UID)
	if err != nil {
		return err
	}
	if user.Status == constant.USER_STATE_BLOCKED { //用户被封
		return errors.WithCode(errors.AccountBlocked)
	}

	wallet, err := s.WalletSrv.GetUserWallet(param.UID) //已经处理用户被封情况
	if err != nil {
		return err
	}

	if wallet.Cash < param.Cash { //金额不足
		return errors.WithCode(errors.InsufficientBalance)
	}
	if user.Mobile == "" {
		return errors.WithCode(errors.MobileNotBind) //手机号未绑定
	}

	card, err := s.Repo.GetWithdrawCard(&entities.WithdrawCard{UID: user.ID, Active: 1, Status: 1})
	if err != nil {
		return err
	}
	if card == nil { //提现银行卡不存在
		return errors.WithCode(errors.WithdrawCardNotExist)
	}

	lastRecord, err := s.Repo.GetHallWithdrawRecord(&entities.HallWithdrawRecord{
		UID: user.ID,
	})

	if err != nil {
		return err
	}
	if lastRecord != nil { //最近的一笔提现记录
		if lastRecord.Status == constant.WITHDRAW_STATE_WAIT_REVIEW || lastRecord.Status == constant.WITHDRAW_STATE_REVIEWED || lastRecord.Status == constant.WITHDRAW_STATE_TRADE_FAIL {
			return errors.WithCode(errors.WithdrawalOrderExists)
		}
		// if time.Now().Sub(lastRecord.CreatedAt) < 24*time.Hour { //判断间隔
		// 	return errors.WithCode(errors.WithdrawalIntervalLimit) //两次提现时间间隔应大于24小时
		// }
	}

	todayCount, err := s.Repo.GetTodayWithdrawCount(user.ID)
	if err != nil {
		return err
	}

	if todayCount >= constant.WITHDRAW_DAY_MAX_COUNT { //每日次数限制
		return errors.WithCode(errors.WithdrawalDayCountLimit)
	}

	record := entities.HallWithdrawRecord{
		OrderID:       createWithdrawOrderID(user.ID),
		UID:           user.ID,
		PromoterCode:  user.PromoterCode,
		IFSC:          card.IFSC,
		AccountNumber: card.AccountNumber,
		AccountName:   card.Name,
		Cash:          param.Cash,
		Rate:          uint(WITHDRAW_RATE), //抽水
		StartTime:     time.Now().Unix(),
	}
	record.CalculateFee() //计算抽水

	err = s.WalletSrv.HandleWallet(user.ID, func(wallet *entities.UserWallet, tx *gorm.DB) error {
		if err := s.Repo.CreateHallWithdrawRecordWithTx(tx, &record); err != nil {
			return err
		}
		wallet.SafeAdjustCash(-record.Cash)

		if err := s.WalletSrv.CreateFundFreeze(&entities.FundFreeze{ //创建冻结记录
			UID:          user.ID,
			FreezeAmount: record.Cash,
			RecordID:     record.OrderID,
			Status:       constant.FUND_STATUS_FREEZE,
			Reason:       "提现冻结",
			FreezeType:   constant.FREEZE_TYPE_WITHDRAW,
			Currency:     constant.CURRENCY_CASH,
		}); err != nil {
			return err
		}

		logger.ZInfo("ApplyForWithdrawal UpdateUserWithTx", zap.Uint("uid", wallet.ID), zap.Float64("balance", wallet.Cash))
		if err := s.WalletSrv.UpdateCashWithTx(tx, wallet); err != nil {
			return err
		}
		createFlowQueue, _ := handle.NewCreateFlowQueue(&entities.Flow{
			UID:          user.ID,
			FlowType:     constant.FLOW_TYPE_APPLY_FOR_WITHDRAW_CASH,
			Number:       -record.Cash,
			Balance:      wallet.Cash,
			PromoterCode: user.PromoterCode,
		})
		if _, err := mq.MClient.Enqueue(createFlowQueue); err != nil {
			logger.ZError("createFlowQueue", zap.Any("flow", createFlowQueue), zap.Error(err))
		}
		return nil
	})

	return nil
}

// 充值处理
func (s *WithdrawService) WithdrawCallbackProcess(using pay.IPayBack) (resp string, err error) {

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

	logger.ZInfo("WithdrawCallbackProcess start ---------------------------------------------------------", zap.String("orderID", using.GetTransferOrder().OrderNo))
	defer func() {
		// 如果存在错误，就打印错误
		if err != nil {
			logger.ZError("WithdrawCallbackProcess", zap.Any("transfer", using), zap.Error(err))
		}

		logger.ZInfo("WithdrawCallbackProcess end ---------------------------------------------------------", zap.String("orderID", using.GetTransferOrder().OrderNo))
	}()

	var order *entities.HallWithdrawRecord
	if order, err = s.Repo.GetHallWithdrawRecord(&entities.HallWithdrawRecord{OrderID: transfer.GetMerOrderNo()}); err != nil {
		return
	}

	if order == nil {
		err = errors.With("order not exist")
		return
	}
	order.TradeID = transfer.GetOrderNo()
	logger.ZInfo("WithdrawCallbackProcess", zap.Any("order", order))
	if !using.IsTransactionSucc() {
		// err = errors.With("Transaction failed")
		if order.Status == constant.WITHDRAW_STATE_REVIEWED { //打款中----
			orderForUpdate := &entities.HallWithdrawRecord{
				TradeID:    order.TradeID,
				FinishTime: time.Now().Unix(),
				Status:     constant.WITHDRAW_STATE_TRADE_FAIL, //打款失败
			}
			orderForUpdate.ID = order.ID

			if err := s.Repo.UpdateHallWithdrawRecord(orderForUpdate); err != nil {
				logger.ZError("Transaction failed UpdateHallWithdrawRecord", zap.Any("order", orderForUpdate), zap.Error(err))
			} //更新订单状态
		}
		return
	}

	if order.Status != constant.WITHDRAW_STATE_REVIEWED {
		// err = errors.With("Order status not reviewed")

		return using.GetSuccResp(), nil
	}

	equal := CompareStringFloat(transfer.GetOrderAmount(), order.RealCash)

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
func (s *WithdrawService) HandleBusiAfterTradeSucc(order *entities.HallWithdrawRecord) error {

	orderForUpdate := &entities.HallWithdrawRecord{
		TradeID:    order.TradeID,
		FinishTime: time.Now().Unix(),
		Status:     constant.WITHDRAW_STATE_TRADE_SUCC,
	}
	orderForUpdate.ID = order.ID

	logger.ZInfo("HandleBusiAfterTradeSucc UpdateUserWithTx", zap.Uint("uid", order.UID), zap.String("orderID", order.OrderID))

	var fundFreeze *entities.FundFreeze
	fundFreeze, err := s.WalletSrv.GetFundFreeze(&entities.FundFreeze{RecordID: order.OrderID})
	if err != nil {
		return err
	}

	err = s.WalletSrv.HandleWallet(order.ID, func(wallet *entities.UserWallet, tx *gorm.DB) error {
		logger.ZInfo("HandleBusiAfterTradeSucc UpdateUserWithTx2", zap.Uint("uid", order.UID), zap.String("orderID", order.OrderID))

		if err := s.Repo.UpdateHallWithdrawRecordWithTx(tx, orderForUpdate); err != nil {
			return err
		} //更新订单状态

		fundFreeze.Status = constant.FUND_STATUS_THAW //解冻状态
		if err := s.WalletSrv.UpdateFundFreezeWithTx(tx, fundFreeze); err != nil {
			return err
		}

		logger.ZInfo("withdraw HandleBusiAfterTradeSucc UpdateUserWithTx", zap.Uint("uid", wallet.UID), zap.Float64("FreezeAmount", fundFreeze.FreezeAmount))

		completedWithdraw :=
			entities.CompletedWithdraw{
				UID:          wallet.ID,
				PromoterCode: wallet.PromoterCode,
				OrderID:      order.OrderID,
				TradeID:      order.TradeID,
				Amount:       order.Cash,
				Channel:      order.Channel,
				CreateTime:   time.Now().Unix(),
			}
		currentTime := time.Now()
		midnight := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location())
		completedWithdraw.TodayTime = midnight.Unix()

		if err := s.Repo.CreateCompletedWithdrawWithTx(tx, &completedWithdraw); err != nil { //加入已完成充值表
			return err
		}

		return nil
	})

	return err
}
