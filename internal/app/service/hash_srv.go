package service

import (
	"context"
	"rk-api/internal/app/constant"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/errors"
	"rk-api/internal/app/mq"
	"rk-api/internal/app/mq/handle"
	"rk-api/internal/app/service/repository"
	"rk-api/pkg/chain"
	"rk-api/pkg/logger"
	"time"

	"github.com/google/wire"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var HashGameServiceSet = wire.NewSet(
	ProvideHashGameService,
)

type HashGameService struct {
	Repo         *repository.HashGameRepository
	UserSrv      *UserService
	BlockFetcher *chain.BlockFetcher

	WalletSrv *WalletService
}

func ProvideHashGameService(repo *repository.HashGameRepository,
	userSrv *UserService,
	walletSrv *WalletService,
) *HashGameService {

	fetcher := chain.NewBlockFetcher([]string{"https://apilist.tronscan.org/api"}, 5) // 5 requests per second
	go fetcher.StartBackgroundUpdate(context.Background())

	return &HashGameService{
		Repo:         repo,
		UserSrv:      userSrv,
		WalletSrv:    walletSrv,
		BlockFetcher: fetcher,
	}
}

func (s *HashGameService) CreateHashGameRound(round entities.IHashGameRound) error {
	return s.Repo.CreateHashGameRound(round)
}

func (s *HashGameService) InsertHashGameRound(round entities.IHashGameRound) error {
	return s.Repo.InitHashGameRound(round)
}

func (s *HashGameService) UpdateHashGameRound(round entities.IHashGameRound) error {
	if err := s.Repo.UpdateHashGameRound(round); err != nil {
		return err
	}
	return nil
}

func (s *HashGameService) CreateHashGameOrder(order entities.IHashGameOrder) error {
	user, err := s.UserSrv.GetUserByUID(order.GetUID())
	if err != nil {
		return err
	}

	wallet, err := s.WalletSrv.GetWallet(order.GetUID())
	if err != nil {
		return err
	}

	if wallet.Cash < order.GetBetAmount() {
		return errors.WithCode(errors.InsufficientBalance)
	}

	err = s.WalletSrv.HandleWallet(user.ID, func(wallet *entities.UserWallet, tx *gorm.DB) error {
		wallet.SafeAdjustCash(-order.GetBetAmount())
		order.CalculateFee() //计算抽水
		order.SetPromoterCode(user.PromoterCode)
		if err := s.Repo.CreateHashGameOrderWithTx(tx, order); err != nil { //创建投注单
			return err
		}

		if err := s.WalletSrv.UpdateCashWithTx(tx, wallet); err != nil {
			return err
		}

		createFlowQueue, _ := handle.NewCreateFlowQueue(&entities.Flow{
			UID:          order.GetUID(),
			FlowType:     constant.FlOW_TYPE_SD,
			Number:       -order.GetBetAmount(),
			Balance:      wallet.Cash,
			PromoterCode: user.PromoterCode,
		})
		if _, err := mq.MClient.Enqueue(createFlowQueue); err != nil {
			logger.ZError("createFlowQueue", zap.Any("flow", createFlowQueue), zap.Error(err))
		}
		logger.ZInfo("CreateWingoOrder", zap.Any("order", order))
		return nil
	})
	return err
}

func (s *HashGameService) SettlePlayerOrders(orders []entities.IHashGameOrder) error {
	// 过滤已处理的订单
	pendingOrders := make([]entities.IHashGameOrder, 0, len(orders))
	for _, order := range orders {
		if order.GetStatus() == constant.STATUS_CREATE {
			order.SetEndTime(time.Now().Unix())
			pendingOrders = append(pendingOrders, order)
		}
	}

	// 按用户分组
	userOrderMap := make(map[uint][]entities.IHashGameOrder)
	for _, order := range pendingOrders {
		userOrderMap[order.GetUID()] = append(userOrderMap[order.GetUID()], order)
	}

	// 分批次处理，每100个订单一个批次
	batchSize := 100
	uids := make([]uint, 0, len(userOrderMap))
	for uid := range userOrderMap {
		uids = append(uids, uid)
	}

	for i := 0; i < len(uids); i += batchSize {
		end := i + batchSize
		if end > len(uids) {
			end = len(uids)
		}
		batchUids := uids[i:end]

		// 处理每个批次
		if err := s.settleUserBatch(batchUids, userOrderMap); err != nil {
			return err
		}
	}

	logger.ZInfo("SettlePlayerOrders processed", zap.Int("order_count", len(pendingOrders)))
	return nil
}

func (s *HashGameService) settleUserBatch(batchUids []uint, userOrderMap map[uint][]entities.IHashGameOrder) error {
	tx := s.Repo.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 预加载本批次用户的钱包信息
	var wallets []*entities.UserWallet
	if err := tx.Model(&entities.UserWallet{}).Where("uid IN (?)", batchUids).Find(&wallets).Error; err != nil {
		tx.Rollback()
		return err
	}

	walletMap := make(map[uint]*entities.UserWallet)
	for _, w := range wallets {
		walletMap[w.UID] = w
	}

	// 处理每个用户的订单
	flows := make([]*entities.Flow, 0)
	for _, uid := range batchUids {
		userOrders := userOrderMap[uid]
		var totalReward float64
		orderIDs := make([]uint, 0, len(userOrders))

		for _, order := range userOrders {
			orderIDs = append(orderIDs, order.GetID())
			totalReward += order.GetRewardAmount()
		}

		// 批量更新订单状态
		if len(orderIDs) > 0 {
			if err := tx.Model(&entities.HashSDGameOrder{}).Where("id IN (?)", orderIDs).
				Update("status", constant.STATUS_SETTLE).Error; err != nil {
				tx.Rollback()
				return err
			}
		}

		if totalReward > 0 {
			// 原子更新钱包现金
			if err := tx.Model(&entities.UserWallet{}).Where("uid = ?", uid).
				Update("cash", gorm.Expr("cash + ?", totalReward)).Error; err != nil {
				tx.Rollback()
				return err
			}

			// 查询更新后的余额
			var cash float64
			if err := tx.Model(&entities.UserWallet{}).Where("uid = ?", uid).
				Pluck("cash", &cash).Error; err != nil {
				tx.Rollback()
				return err
			}
			// 清除钱包缓存
			s.WalletSrv.ClearWalletCache(uid)

			// 生成流水记录
			for _, order := range userOrders {
				if order.GetRewardAmount() > 0 {
					flows = append(flows, &entities.Flow{
						UID:          uid,
						FlowType:     constant.FlOW_TYPE_SD_REWARD,
						Number:       order.GetRewardAmount(),
						Balance:      cash,
						PromoterCode: walletMap[uid].PromoterCode,
					})
				}
			}
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	// 事务提交成功后发送MQ消息
	for _, flow := range flows {
		createFlowQueue, _ := handle.NewCreateFlowQueue(flow)
		if _, err := mq.MClient.Enqueue(createFlowQueue); err != nil {
			logger.ZError("createFlowQueue failed", zap.Any("flow", flow), zap.Error(err))
		}
	}

	return nil
}
