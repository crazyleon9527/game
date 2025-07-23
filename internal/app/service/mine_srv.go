package service

import (
	"fmt"
	"rk-api/internal/app/constant"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/errors"
	"rk-api/internal/app/mq"
	"rk-api/internal/app/mq/handle"
	"rk-api/internal/app/service/repository"
	"rk-api/pkg/logger"
	"time"

	"github.com/google/wire"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var MineGameServiceSet = wire.NewSet(
	ProvideMineGameService,
)

type MineGameService struct {
	Repo      *repository.MineGameRepository
	UserSrv   *UserService
	WalletSrv *WalletService
}

func ProvideMineGameService(
	repo *repository.MineGameRepository,
	userSrv *UserService,
	walletSrv *WalletService,
) *MineGameService {
	service := &MineGameService{
		Repo:      repo,
		UserSrv:   userSrv,
		WalletSrv: walletSrv,
	}
	return service
}

// ------------------------------------ MineGameOrder ------------------------------------

func (s *MineGameService) GetUserMineGameOrder(uid uint) (*entities.MineGameOrder, error) {
	return s.Repo.GetUserMineGameOrder(uid)
}

func (s *MineGameService) GetUserMineGameOrderList(uid uint) ([]*entities.MineGameOrder, error) {
	return s.Repo.GetUserMineGameOrderList(uid)
}

func (s *MineGameService) UpdateMineGameOrderStatus(uid uint, roundID uint64, status int) error {
	return s.Repo.UpdateMineGameOrderStatus(uid, roundID, status)
}

func (s *MineGameService) UpdateMineGameOrder(order *entities.MineGameOrder) error {
	return s.Repo.UpdateMineGameOrder(order)
}

func (s *MineGameService) CreateMineGameOrder(order *entities.MineGameOrder) error {
	return s.Repo.CreateMineGameOrder(order)
}

func (s *MineGameService) PlaceOrder(order *entities.MineGameOrder) error {
	user, err := s.UserSrv.GetUserByUID(order.UID)
	if err != nil {
		return err
	}

	wallet, err := s.WalletSrv.GetWallet(order.UID)
	if err != nil {
		return err
	}

	if wallet.Cash < order.BetAmount {
		return errors.WithCode(errors.InsufficientBalance)
	}

	err = s.WalletSrv.HandleWallet(user.ID, func(wallet *entities.UserWallet, tx *gorm.DB) error {
		wallet.SafeAdjustCash(-order.BetAmount)
		order.CalculateFee() //计算抽水
		order.PromoterCode = user.PromoterCode
		if err := s.Repo.UpdateMineGameOrderWithTx(tx, order); err != nil { //创建投注单
			return err
		}

		if err := s.WalletSrv.UpdateCashWithTx(tx, wallet); err != nil {
			return err
		}

		createFlowQueue, _ := handle.NewCreateFlowQueue(&entities.Flow{
			UID:          order.UID,
			FlowType:     constant.FLOW_TYPE_MINE,
			Number:       -order.BetAmount,
			Balance:      wallet.Cash,
			PromoterCode: user.PromoterCode,
		})
		if _, err := mq.MClient.Enqueue(createFlowQueue); err != nil {
			logger.ZError("createFlowQueue", zap.Any("flow", createFlowQueue), zap.Error(err))
		}
		logger.ZInfo("MineGameService.PlaceOrder", zap.Any("order", order))
		return nil
	})
	return err
}

func (s *MineGameService) SettleOrder(order *entities.MineGameOrder) error {
	if order.Settled == constant.STATUS_SETTLE {
		return nil
	}
	order.Settled, order.EndTime = constant.STATUS_SETTLE, time.Now().Unix()

	tx := s.Repo.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新订单状态
	if err := s.Repo.UpdateMineGameOrderWithTx(tx, order); err != nil {
		tx.Rollback()
		return err
	}

	var flow *entities.Flow
	if order.RewardAmount > 0 {
		// 原子更新钱包现金
		if err := tx.Model(&entities.UserWallet{}).Where("uid = ?", order.UID).
			Update("cash", gorm.Expr("cash + ?", order.RewardAmount)).Error; err != nil {
			tx.Rollback()
			return err
		}

		// 查询更新后的余额
		var cash float64
		if err := tx.Model(&entities.UserWallet{}).Where("uid = ?", order.UID).
			Pluck("cash", &cash).Error; err != nil {
			tx.Rollback()
			return err
		}

		flow = &entities.Flow{
			UID:          order.UID,
			FlowType:     constant.FlOW_TYPE_MINE_REWARD,
			Number:       order.RewardAmount,
			Balance:      cash,
			PromoterCode: order.PromoterCode,
		}
	}

	// 创建game_record
	record := &entities.GameRecord{
		Category:     constant.GameCategoryMine,
		RecordId:     fmt.Sprintf("mine-%d-%d-%d", order.RoundID, time.Now().UnixMilli(), order.UID),
		BetTime:      time.Unix(order.BetTime, 0),
		BetAmount:    order.BetAmount,
		Amount:       order.BetAmount,
		Profit:       order.RewardAmount,
		Game:         constant.GameNameMine,
		Status:       constant.STATUS_SETTLE,
		UID:          order.UID,
		Currency:     constant.CurrencyCNY, // 假设使用人民币，可以根据实际情况调整
		PromoterCode: order.PromoterCode,
	}
	if err := tx.Model(&entities.GameRecord{}).Create(record).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	// 清除钱包缓存
	s.WalletSrv.ClearWalletCache(order.UID)

	// 事务提交成功后发送MQ消息 生成流水记录
	if flow != nil {
		createFlowQueue, _ := handle.NewCreateFlowQueue(flow)
		if _, err := mq.MClient.Enqueue(createFlowQueue); err != nil {
			logger.ZError("createFlowQueue failed", zap.Any("flow", flow), zap.Error(err))
		}
	}

	return nil
}
