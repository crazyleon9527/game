package service

import (
	"fmt"
	"rk-api/internal/app/constant"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/errors"
	"rk-api/internal/app/mq"
	"rk-api/internal/app/mq/handle"
	"rk-api/internal/app/service/repository"
	"rk-api/internal/app/utils"
	"rk-api/pkg/cjson"
	"rk-api/pkg/logger"
	"sync"
	"time"

	"github.com/google/wire"
	"github.com/orca-zhang/ecache"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var ActivityServiceSet = wire.NewSet(
	ProvideActivityService,
)

const (
	BannerList = "banner_list"
	LogoList   = "logo_list"
)

type ActivityService struct {
	Repo             *repository.ActivityRepository
	UserSrv          *UserService
	WalletSrv        *WalletService
	redEnvelopeMutex sync.Mutex //红包锁
	actCache         *ecache.Cache
}

func ProvideActivityService(repo *repository.ActivityRepository, userSrv *UserService, walletSrv *WalletService) *ActivityService {
	actCache := ecache.NewLRUCache(1, 6, 5*time.Minute) //初始化缓存
	return &ActivityService{
		Repo:             repo,
		UserSrv:          userSrv,
		WalletSrv:        walletSrv,
		redEnvelopeMutex: sync.Mutex{},
		actCache:         actCache,
	}
}

func (s *ActivityService) GetActivityList() ([]*entities.Activity, error) {
	return s.Repo.GetActivityList()
}

func (s *ActivityService) GetBannerList() ([]*entities.Banner, error) {
	if val, ok := s.actCache.Get(BannerList); ok {
		if v, ok := val.([]*entities.Banner); ok {
			return v, nil
		}
	}

	list, err := s.Repo.GetBannerList()
	if err != nil {
		return nil, err
	}
	s.actCache.Put(BannerList, list)
	return list, nil
}

// GetLogoList
func (s *ActivityService) GetLogoList(logoType int) ([]*entities.Logo, error) {
	k := fmt.Sprintf("%s_%d", LogoList, logoType)
	if val, ok := s.actCache.Get(k); ok {
		if v, ok := val.([]*entities.Logo); ok {
			return v, nil
		}
	}

	list, err := s.Repo.GetLogoList(logoType)
	if err != nil {
		return nil, err
	}
	s.actCache.Put(k, list)
	return list, nil
}

func (s *ActivityService) GetRedEnvelopeAmount(redName string) (float64, error) {
	setting, err := s.Repo.GetHongbaoSettingByName(redName)
	if err != nil {
		return 0, err
	}
	if setting == nil {
		return 0, errors.WithCode(errors.RedEnvelopeNotExist)
	}
	return setting.Amount, nil
}

func (s *ActivityService) DelRedEnvelope(req *entities.DelRedEnvelopeReq) (err error) {

	// logger.Info("-=-------------------------------------", req.Name)
	setting, err := s.Repo.GetHongbaoSettingByName(req.Name)
	if err != nil {
		return err
	}
	if setting == nil {
		return errors.WithCode(errors.RedEnvelopeNotExist)
	}

	defer func() {
		log := entities.SystemOptionLog{
			Type:     constant.SYS_OPTION_TYPE_EDIT_USER_INFO,
			OptionID: req.OptionID,
			IP:       req.IP,
			Content:  cjson.StringifyIgnore(req),
		}
		if err != nil {
			log.Result = "false"
			log.Remark = "删除红包失败"
			// 打印错误，使用你的日志库方法来记录
			logger.ZError("DelRedEnvelope", zap.Any("req", req), zap.Error(err))
		} else {
			log.Result = "true"
			log.Remark = "删除红包成功"
			logger.ZInfo("DelRedEnvelope", zap.Any("req", req))
		}
		optionLogQueue, _ := handle.NewOptionLogQueue(&log)
		if _, err := mq.MClient.Enqueue(optionLogQueue); err != nil {
			logger.ZError("optionLogQueue", zap.Error(err))
		}
	}()

	if err = s.Repo.DelHongbaoSettingByName(req.Name); err != nil {
		return err
	}
	return nil

}

func (s *ActivityService) AddRedEnvelope(req *entities.AddRedEnvelopeReq) (err error) {

	defer func() {
		log := entities.SystemOptionLog{
			Type:     constant.SYS_OPTION_TYPE_EDIT_USER_INFO,
			OptionID: req.OptionID,
			IP:       req.IP,
			Content:  cjson.StringifyIgnore(req),
		}
		if err != nil {
			log.Result = "false"
			log.Remark = "添加红包失败"
			// 打印错误，使用你的日志库方法来记录
			logger.ZError("AddRedEnvelope", zap.Any("req", req), zap.Error(err))
		} else {
			log.Result = "true"
			log.Remark = "添加红包成功"
			logger.ZInfo("AddRedEnvelope", zap.Any("req", req))
		}
		optionLogQueue, _ := handle.NewOptionLogQueue(&log)
		if _, err := mq.MClient.Enqueue(optionLogQueue); err != nil {
			logger.ZError("optionLogQueue", zap.Error(err))
		}
	}()

	hong := entities.HongbaoSetting{
		Number: req.Number,
		Amount: req.Amount,
		SYSUID: req.OptionID,
		Type:   req.Type,
		Remark: req.Remark,
		Name:   utils.NewRandomString(8),
	}

	if err = s.Repo.CreateHongbaoSetting(&hong); err != nil {
		return err
	}
	return nil

}

func (s *ActivityService) GetRedEnvelope(req *entities.GetRedEnvelopeReq) error {

	s.redEnvelopeMutex.Lock() //互斥锁
	defer s.redEnvelopeMutex.Unlock()

	if req.UID == 0 {
		return errors.WithCode(errors.AccountNotExist)
	}

	setting, err := s.Repo.GetHongbaoSettingByName(req.RedName)
	if err != nil {
		return err
	}
	if setting == nil {
		return errors.WithCode(errors.RedEnvelopeNotExist)
	}

	if time.Unix(setting.CreatedAt, 0).Add(24 * time.Hour).Before(time.Now()) {
		return errors.WithCode(errors.RedEnvelopeExpire)
	}

	count, err := s.Repo.GetHongbaoCountByHongID(setting.ID)
	if err != nil {
		return err
	}

	if count >= int64(setting.Number) {
		return errors.WithCode(errors.InsufficientRedEnvelope)
	}

	hongbao, err := s.Repo.GetHongbao(&entities.HongbaoRecord{HongID: setting.ID, UID: req.UID})

	if err != nil {
		return err
	}
	if hongbao != nil {
		return errors.WithCode(errors.RedEnvelopeRepeatGet)
	}

	user, err := s.UserSrv.GetUserByUID(req.UID)
	if err != nil {
		return err
	}

	if setting.Type == constant.RED_GET_TYPE_NORMAL {

		err := s.WalletSrv.HandleWallet(req.UID, func(wallet *entities.UserWallet, tx *gorm.DB) error {
			redAmount := setting.Amount //金额 直接等于 设置中的金额
			hongbao = &entities.HongbaoRecord{
				UID:          wallet.UID,
				PromoterCode: wallet.PromoterCode,
				HongID:       setting.ID,
				HongName:     setting.Name,
				Amount:       redAmount,
				Username:     user.Username,
			}
			wallet.SafeAdjustCash(hongbao.Amount) //增加金额
			setting.ReceiveNumber += 1
			if err := s.Repo.UpdatebaoSettingByWithTx(tx, setting); err != nil {
				return err
			}
			if err := s.Repo.CreateHongbaoRecordWithTx(tx, hongbao); err != nil {
				return err
			}

			logger.ZInfo("GetRedEnvelope UpdateUserWithTx", zap.Uint("uid", wallet.ID), zap.Float64("cash", wallet.Cash))
			if err := s.WalletSrv.UpdateCashWithTx(tx, wallet); err != nil {
				return err
			}
			createFlowQueue, _ := handle.NewCreateFlowQueue(&entities.Flow{
				UID:          hongbao.UID,
				FlowType:     constant.FLOW_TYPE_GET_RED_ENVELOPE,
				Number:       hongbao.Amount,
				Balance:      wallet.Cash,
				PromoterCode: wallet.PromoterCode,
			})
			if _, err := mq.MClient.Enqueue(createFlowQueue); err != nil {
				logger.ZError("createFlowQueue", zap.Any("flow", createFlowQueue), zap.Error(err))
			}
			logger.ZInfo("GetRedEnvelope succ",
				zap.Uint("uid", hongbao.UID),
				zap.String("username", hongbao.Username),
				zap.Float64("balance", wallet.Cash),
				zap.Uint("hong_id", hongbao.HongID),
				zap.String("hong_name", hongbao.HongName),
				zap.Float64("amount", hongbao.Amount),
			)
			return nil
		})

		return err
	}

	return errors.With(fmt.Sprintf("not support type %d", setting.Type))
}

func countRatio(base, cbxRatio float64) float64 {
	return base / cbxRatio
}

func amountRatio(base, abxRatio float64) float64 {
	return 1 / (base + abxRatio)
}

func getPinAmount(amount float64, count int, base, abxRatio, cbxRatio float64) float64 {
	ratio := amountRatio(base, abxRatio)
	if base <= 1 {
		return float64(count) / (float64(count) + 1) * amount
	}
	if float64(count) < base {
		return float64(count) * ratio * amount
	}
	stepAmount := base * ratio * amount
	amount -= stepAmount
	count -= int(base)
	base = countRatio(base, cbxRatio)

	return stepAmount + getPinAmount(amount, count, base, abxRatio, cbxRatio)
}

func (s *ActivityService) InvitePinduoOnce(relation *entities.HallInvitePinduo) error {
	// 假定ID为1的设置被认为是有效的设置
	var setting *entities.PinduoSetting

	setting, err := s.Repo.GetPinduoSetting(&entities.PinduoSetting{
		ID: 1,
	})
	if err != nil {
		return err
	}
	if setting == nil {
		return errors.WithCode(errors.PinduoNotExist)
	}

	now := time.Now().Unix()
	if int64(setting.StartTime) > now || int64(setting.EndTime) < now {
		return errors.WithCode(errors.PinduoExpire)
	}

	pinduoRecord, err := s.Repo.GetPinduoRecordByUID(relation.InviteID)
	if err != nil {
		return err
	}
	if pinduoRecord == nil {
		inviteCount := 1
		suggestCash := getPinAmount(setting.Amount, inviteCount, setting.Base, setting.AbxRatio, setting.CbxRatio)

		pinduoRecord = &entities.PinduoRecord{
			UID:         relation.InviteID,
			UserName:    relation.InviteName,
			LockCash:    setting.Amount,
			SuggestCash: suggestCash,
			InviteCount: inviteCount,
			PinID:       setting.ID,
			Created:     now,
		}
		if err := s.Repo.CreatePinduoRecord(pinduoRecord); err != nil {
			return err
		}
	} else {
		if pinduoRecord.Status != 0 {
			return errors.WithCode(errors.PinduoHasGet)
		}

		pinduoRecord.InviteCount++
		pinduoRecord.SuggestCash = getPinAmount(setting.Amount, pinduoRecord.InviteCount, setting.Base, setting.AbxRatio, setting.CbxRatio)
		if err := s.Repo.UpdatePinduoRecord(pinduoRecord); err != nil {
			return err
		}
	}
	return nil
}

func (s *ActivityService) JoinPinDuo(uid uint) (*entities.PinduoInfo, error) {
	s.redEnvelopeMutex.Lock() //互斥锁
	defer s.redEnvelopeMutex.Unlock()

	var setting *entities.PinduoSetting
	setting, err := s.Repo.GetPinduoSetting(&entities.PinduoSetting{
		ID: 1,
	})
	if err != nil {
		return nil, err
	}
	if setting == nil {
		return nil, errors.WithCode(errors.PinduoNotExist)
	}
	user, err := s.UserSrv.GetUserByUID(uid)
	if err != nil {
		return nil, err
	}

	pinduo, err := s.Repo.GetPinduoRecordByUID(uid)
	if err != nil {
		return nil, err
	}

	if pinduo == nil {
		pinduo = &entities.PinduoRecord{
			UID:         user.ID,
			UserName:    user.Username,
			LockCash:    setting.Amount,
			SuggestCash: 0,
			InviteCount: 0,
			Status:      0,
			PinID:       setting.ID,
			Created:     time.Now().Unix(),
		}
		if err := s.Repo.CreatePinduoRecord(pinduo); err != nil {
			return nil, err
		}
	}

	data := entities.PinduoInfo{
		EndTime:     setting.EndTime,
		InviteCount: uint(pinduo.InviteCount),
		Status:      pinduo.Status,
		LockCash:    pinduo.LockCash,
		SuggestCash: pinduo.SuggestCash,
	}

	return &data, nil
}

func (s *ActivityService) ReceivePinduoBonus(uid uint) (*entities.PinduoRecord, error) {
	user, err := s.UserSrv.GetUserByUID(uid)
	if err != nil {
		return nil, err
	}
	pinduo, err := s.Repo.GetPinduoRecordByUID(uid)
	if err != nil {
		return nil, err
	}
	if pinduo == nil {
		return nil, errors.WithCode(errors.PinduoNotExist)
	}
	if pinduo.Status != 0 {
		return nil, errors.With("pinduo cash has get")
	}
	if pinduo.SuggestCash < pinduo.LockCash {
		return nil, errors.With("no condition")
	}
	money := pinduo.LockCash

	err = s.WalletSrv.HandleWallet(uid, func(wallet *entities.UserWallet, tx *gorm.DB) error {
		wallet.SafeAdjustCash(money)

		pinduo.Status = 1
		if err := s.Repo.UpdatePinduoRecordWithTx(tx, pinduo); err != nil {
			return err
		}

		if err := s.WalletSrv.UpdateCashWithTx(tx, wallet); err != nil {
			return err
		}
		createFlowQueue, _ := handle.NewCreateFlowQueue(&entities.Flow{
			UID:          user.ID,
			FlowType:     constant.FLOW_TYPE_PINDUO,
			Number:       money,
			Balance:      wallet.Cash,
			PromoterCode: user.PromoterCode,
		})
		if _, err := mq.MClient.Enqueue(createFlowQueue); err != nil {
			logger.ZError("createFlowQueue", zap.Any("flow", createFlowQueue), zap.Error(err))
		}
		return nil

	})

	return pinduo, err

}

// 结算利息
func (s *ActivityService) SettleDailyInterest() {
	logger.Info("SettleDailyInterest----------------------------------------------")
	s.dailyInterestCalculation(s.Repo.DB, s.Repo.RDS, 50)
}

// dailyInterestCalculation 每日计算利息并更新用户表
func (s *ActivityService) dailyInterestCalculation(db *gorm.DB, rds redis.UniversalClient, batchSize int) {
	// currentDate := time.Now().Format("2006-01-02")
	// var log entities.InterestCalculationLog
	// result := db.FirstOrCreate(&log, entities.InterestCalculationLog{Date: currentDate})
	// if result.Error != nil {
	// 	logger.Error("Failed to create or retrieve the log: %v", result.Error)
	// 	return
	// }
	// if log.IsCalculated {
	// 	// 如果当天的利息已经计算过，则不再计算
	// 	return
	// }

	// var wg sync.WaitGroup
	// sem := make(chan struct{}, 3) // 控制并发量的信号量 //runtime.NumCPU()
	// // offset := 0                   // 查询数据库的偏移

	// // interestRate := decimal.NewFromFloat(0.008)

	// // interestRate := decimal.NewFromFloat(0.03)
	// interestRate := decimal.NewFromFloat(0.1)
	// lastProcessedID := uint(0)
	// for {
	// 	users := []entities.User{}
	// 	// result := db.Where("withdraw_cash > 500").Order("id").Offset(offset).Limit(batchSize).Find(&users)
	// 	//比500 要大
	// 	// 保存上一次处理的最大ID
	// 	// result := db.Select("id,withdraw_cash,interest,pc").Where("id > ?", lastProcessedID).Where("withdraw_cash > 500").Order("id").Offset(offset).Limit(batchSize).Find(&users) //比500 要大
	// 	result := db.Select("id,withdraw_cash,interest,pc").Where("id > ?", lastProcessedID).Where("withdraw_cash > 500").Order("id").Limit(batchSize).Find(&users) //比500 要大
	// 	if result.Error != nil {
	// 		logger.Error("Failed to query users: %v", result.Error)
	// 		break
	// 	}
	// 	if len(users) == 0 {
	// 		break //没有更多用户
	// 	}
	// 	// 更新lastProcessedID为本批次处理的最后一个用户的ID
	// 	lastProcessedID = users[len(users)-1].ID

	// 	wg.Add(1)         // 增加等待组计数
	// 	sem <- struct{}{} // 请求并发量信号量

	// 	go func(users []entities.User) {
	// 		defer func() {
	// 			if r := recover(); r != nil {
	// 				logger.ZError("Recovered in dailyInterestCalculation", zap.Any("Error", r))
	// 			}
	// 		}()

	// 		defer wg.Done()          // 执行完毕后减少等待组计数
	// 		defer func() { <-sem }() // 释放并发量信号量

	// 		userIDKeyList := make([]string, 0, len(users))
	// 		userUpdateList := make([]*entities.User, 0, len(users))
	// 		for i := range users {
	// 			user := users[i]
	// 			interestDecimal := user.CalculateDailyInterest(interestRate)
	// 			interest := interestDecimal.InexactFloat64()
	// 			user.AddInterest(interestDecimal) //增加利息金额，同时增加金额
	// 			user.AddBalance(interest)
	// 			createFlowQueue, _ := handle.NewCreateFlowQueue(&entities.Flow{
	// 				UID:          user.ID,
	// 				FlowType:     constant.FLOW_TYPE_INTEREST,
	// 				Number:       interest,
	// 				Balance:      user.Balance,
	// 				PromoterCode: user.PromoterCode,
	// 			})
	// 			if _, err := mq.MClient.Enqueue(createFlowQueue); err != nil {
	// 				logger.ZError("createFlowQueue", zap.Any("flow", createFlowQueue), zap.Error(err))
	// 			}

	// 			userForUpdate := entities.User{
	// 				Balance:  user.Balance,
	// 				Interest: user.Interest,
	// 			}
	// 			userForUpdate.ID = user.ID
	// 			userUpdateList = append(userUpdateList, &userForUpdate)
	// 			userIDKeyList = append(userIDKeyList, fmt.Sprintf(constant.REDIS_KEY_USER, user.ID))
	// 		}

	// 		err := rds.Del(context.Background(), userIDKeyList...).Err() //先删除缓存
	// 		if err != nil {
	// 			logger.ZError("dailyInterestCalculation expire users fail", zap.Error(err), zap.Any("list", userIDKeyList))
	// 			return
	// 		}

	// 		for _, user := range userUpdateList {
	// 			if err := db.Table("user").Updates(user).Error; err != nil {
	// 				logger.ZError("dailyInterestCalculation updates users fail", zap.Error(err), zap.Any("user", user))
	// 				return
	// 			} // 更新用户表
	// 		}

	// 	}(users)

	// 	// offset += batchSize // 更新偏移量以获取下一批用户
	// }

	// wg.Wait() //等待所有goroutine完成
	// // 标记当天的利息已计算
	// log.IsCalculated = true
	// db.Save(&log)
}
