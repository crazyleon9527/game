package service

import (
	"context"
	"fmt"
	"rk-api/internal/app/chat"
	"rk-api/internal/app/constant"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/errors"
	"rk-api/internal/app/mq"
	"rk-api/internal/app/mq/handle"
	"rk-api/internal/app/service/repository"
	"rk-api/internal/app/utils"
	"rk-api/pkg/logger"
	"strings"
	"time"

	"github.com/google/wire"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var CrashGameServiceSet = wire.NewSet(
	ProvideCrashGameService,
)

type CrashGameService struct {
	Repo      *repository.CrashGameRepository
	UserSrv   *UserService
	WalletSrv *WalletService
	hub       *chat.Hub
	// BlockFetcher *chain.BlockFetcher
}

func ProvideCrashGameService(
	repo *repository.CrashGameRepository,
	userSrv *UserService,
	walletSrv *WalletService,
) *CrashGameService {
	// fetcher := chain.NewBlockFetcher([]string{"https://apilist.tronscan.org/api"}, 5) // 5 requests per second
	// go fetcher.StartBackgroundUpdate(context.Background())

	hub := chat.NewHub()
	service := &CrashGameService{
		Repo:      repo,
		UserSrv:   userSrv,
		WalletSrv: walletSrv,
		hub:       hub,
		// BlockFetcher: fetcher,
	}
	go hub.Run()
	service.StartMessageDispatcher()
	return service
}

// ------------------------------------ socket ------------------------------------

func (s *CrashGameService) Connect(uid uint, conn *websocket.Conn) *chat.Client {
	client := &chat.Client{
		UID:       uid,
		Conn:      conn,
		Send:      make(chan []byte, 256),
		Channels:  make(map[string]struct{}),
		Hub:       s.hub,
		Processor: s,
	}
	s.hub.Register <- client
	return client
}

// 加入频道（线程安全）
func (s *CrashGameService) JoinChannel(uid uint, channel string) {
	s.hub.Join <- &chat.Subscribe{
		UID:     uid,
		Channel: channel,
	}
}

/**
 * 从websocket连接中读取消息并处理消息
 * 暂时不需要
 */
func (s *CrashGameService) ProcessMessage(rawMsg []byte) {
	s.SendMessage("all", rawMsg)
}

/**
 * 直接发送消息
 */
func (s *CrashGameService) SendMessage(channel string, content []byte) {
	// 频道消息
	s.Repo.PublishChannelMessage(channel, content)
}

// 统一的消息分发循环
func (s *CrashGameService) StartMessageDispatcher() {
	logger.Info("start message dispatcher")
	go func() {
		defer utils.PrintPanicStack()
		ctx := context.Background()
		pubsub := s.Repo.RDS.PSubscribe(ctx, "crash_channel:*")
		defer pubsub.Close()

		ch := pubsub.Channel()
		for msg := range ch {
			// 解析频道ID
			var channel = strings.TrimPrefix(msg.Channel, "crash_channel:")

			logger.ZInfo("receive message from crash_channel", zap.String("crash_channel", channel))

			s.hub.Broadcast <- &chat.Broadcast{
				Channel: channel,
				Message: []byte(msg.Payload),
			}
		}
	}()
}

// ------------------------------------ CrashGameRound ------------------------------------

func (s *CrashGameService) GetCrashGameRoundList() ([]*entities.CrashGameRound, error) {
	return s.Repo.GetCrashGameRoundList()
}

func (s *CrashGameService) GetCrashGameRound(roundID uint64) (*entities.CrashGameRound, error) {
	return s.Repo.GetCrashGameRound(roundID)
}

func (s *CrashGameService) GetLatestCrashGameRound() (*entities.CrashGameRound, error) {
	return s.Repo.GetLatestCrashGameRound()
}

func (s *CrashGameService) CreateCrashGameRound(round *entities.CrashGameRound) error {
	return s.Repo.CreateCrashGameRound(round)
}

func (s *CrashGameService) UpdateCrashGameRound(round *entities.CrashGameRound) error {
	return s.Repo.UpdateCrashGameRound(round)
}

// ------------------------------------ CrashGameOrder ------------------------------------

func (s *CrashGameService) GetCrashGameOrders(roundIDs []uint64) ([]*entities.CrashGameOrder, error) {
	return s.Repo.GetCrashGameOrders(roundIDs)
}

func (s *CrashGameService) GetTopHeightCrashGameOrder(roundID uint64) (*entities.CrashGameOrder, error) {
	return s.Repo.GetTopHeightCrashGameOrder(roundID)
}

func (s *CrashGameService) GetTopHeightCrashGameOrderList(roundIDs []uint64) ([]*entities.CrashGameOrder, error) {
	return s.Repo.GetTopHeightCrashGameOrderList(roundIDs)
}

func (s *CrashGameService) GetUserCrashGameOrder(uid uint, roundID uint64) ([]*entities.CrashGameOrder, error) {
	return s.Repo.GetUserCrashGameOrder(uid, roundID)
}

func (s *CrashGameService) GetUserCrashGameOrderList(uid uint) ([]*entities.CrashGameOrder, error) {
	return s.Repo.GetUserCrashGameOrderList(uid)
}

func (s *CrashGameService) CreateCrashGameOrder(order *entities.CrashGameOrder) error {
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
		order.Name = user.Nickname
		order.PromoterCode = user.PromoterCode
		if err := s.Repo.CreateCrashGameOrderWithTx(tx, order); err != nil { //创建投注单
			return err
		}

		if err := s.WalletSrv.UpdateCashWithTx(tx, wallet); err != nil {
			return err
		}

		createFlowQueue, _ := handle.NewCreateFlowQueue(&entities.Flow{
			UID:          order.UID,
			FlowType:     constant.FLOW_TYPE_CRASH,
			Number:       -order.BetAmount,
			Balance:      wallet.Cash,
			PromoterCode: user.PromoterCode,
		})
		if _, err := mq.MClient.Enqueue(createFlowQueue); err != nil {
			logger.ZError("createFlowQueue", zap.Any("flow", createFlowQueue), zap.Error(err))
		}
		logger.ZInfo("CreateCrashGameOrder", zap.Any("order", order))
		return nil
	})
	return err
}

func (s *CrashGameService) CancelCrashGameOrder(order *entities.CrashGameOrder) error {
	err := s.WalletSrv.HandleWallet(order.UID, func(wallet *entities.UserWallet, tx *gorm.DB) error {
		wallet.SafeAdjustCash(order.BetAmount)

		if err := tx.Model(&entities.CrashGameOrder{}).
			Where("uid = ? and round_id = ? and bet_index = ?", order.UID, order.RoundID, order.BetIndex).
			Update("status", constant.STATUS_CANCEL).Error; err != nil { //创建投注单
			return err
		}

		if err := s.WalletSrv.UpdateCashWithTx(tx, wallet); err != nil {
			return err
		}

		createFlowQueue, _ := handle.NewCreateFlowQueue(&entities.Flow{
			UID:          order.UID,
			FlowType:     constant.FLOW_TYPE_CRASH_CANCEL,
			Number:       order.BetAmount,
			Balance:      wallet.Cash,
			PromoterCode: order.PromoterCode,
		})
		if _, err := mq.MClient.Enqueue(createFlowQueue); err != nil {
			logger.ZError("createFlowQueue", zap.Any("flow", createFlowQueue), zap.Error(err))
		}
		logger.ZInfo("CancelCrashGameOrder", zap.Any("order", order))
		return nil
	})
	return err
}

func (s *CrashGameService) SettleOrder(order *entities.CrashGameOrder) error {
	if order.Status == constant.STATUS_SETTLE || order.Status == constant.STATUS_CANCEL {
		return nil
	}
	tx := s.Repo.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新订单状态
	order.Status, order.EndTime = constant.STATUS_SETTLE, time.Now().Unix()
	if err := s.Repo.UpdateCrashGameOrderWithTx(tx, order); err != nil {
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
			FlowType:     constant.FlOW_TYPE_CRASH_REWARD,
			Number:       order.RewardAmount,
			Balance:      cash,
			PromoterCode: order.PromoterCode,
		}
	}

	// 创建game_record
	record := &entities.GameRecord{
		Category:     constant.GameCategoryCrash,
		RecordId:     fmt.Sprintf("crash-%d-%d-%d-%d", order.RoundID, time.Now().UnixMilli(), order.UID, order.BetIndex),
		BetTime:      time.Unix(order.BetTime, 0),
		BetAmount:    order.BetAmount,
		Amount:       order.BetAmount,
		Profit:       order.RewardAmount,
		Game:         constant.GameNameCrash,
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

func (s *CrashGameService) SettlePlayerOrders(orders []*entities.CrashGameOrder) error {
	// 过滤已处理的订单
	pendingOrders := make([]*entities.CrashGameOrder, 0, len(orders))
	for _, order := range orders {
		if order.Status != constant.STATUS_SETTLE && order.Status != constant.STATUS_CANCEL {
			order.Status = constant.STATUS_SETTLE
			order.EndTime = time.Now().Unix()
			pendingOrders = append(pendingOrders, order)
		}
	}

	// 按用户分组
	userOrderMap := make(map[uint][]*entities.CrashGameOrder)
	for _, order := range pendingOrders {
		userOrderMap[order.UID] = append(userOrderMap[order.UID], order)
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

func (s *CrashGameService) settleUserBatch(batchUids []uint, userOrderMap map[uint][]*entities.CrashGameOrder) error {
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
	flows := make([]*entities.Flow, 0, len(batchUids))
	records := make([]*entities.GameRecord, 0, len(batchUids))
	for _, uid := range batchUids {
		userOrders := userOrderMap[uid]
		var totalReward float64
		for _, order := range userOrders {
			totalReward += order.RewardAmount

			records = append(records, &entities.GameRecord{
				Category:     constant.GameCategoryCrash,
				RecordId:     fmt.Sprintf("crash-%d-%d-%d-%d", order.RoundID, time.Now().UnixMilli(), uid, order.BetIndex),
				BetTime:      time.Unix(order.BetTime, 0),
				BetAmount:    order.BetAmount,
				Amount:       order.BetAmount,
				Profit:       order.RewardAmount,
				Game:         constant.GameNameCrash,
				Status:       constant.STATUS_SETTLE,
				UID:          uid,
				Currency:     constant.CurrencyCNY, // 假设使用人民币，可以根据实际情况调整
				PromoterCode: order.PromoterCode,
			})

			if err := tx.Model(&entities.CrashGameOrder{}).
				Where("uid = ? and round_id = ? and bet_index = ?", order.UID, order.RoundID, order.BetIndex).
				Updates(map[string]interface{}{
					"status":        constant.STATUS_SETTLE,
					"end_time":      time.Now().Unix(),
					"reward_amount": order.RewardAmount,
					"escape_height": order.EscapeHeight,
				}).Error; err != nil {
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

			// 生成流水记录
			for _, order := range userOrders {
				if order.RewardAmount > 0 {
					flows = append(flows, &entities.Flow{
						UID:          uid,
						FlowType:     constant.FlOW_TYPE_CRASH_REWARD,
						Number:       order.RewardAmount,
						Balance:      cash,
						PromoterCode: walletMap[uid].PromoterCode,
					})
				}
			}
		}
	}

	// 创建game_record
	if err := tx.Model(&entities.GameRecord{}).CreateInBatches(records, len(records)).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	for _, uid := range batchUids {
		// 清除钱包缓存
		s.WalletSrv.ClearWalletCache(uid)
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

// ------------------------------------ CrashAutoBet ------------------------------------

func (s *CrashGameService) GetCrashAutoBetList(status int) ([]*entities.CrashAutoBet, error) {
	return s.Repo.GetCrashAutoBetList(status)
}

func (s *CrashGameService) GetCrashAutoBet(uid uint) (*entities.CrashAutoBet, error) {
	return s.Repo.GetCrashAutoBet(uid)
}

func (s *CrashGameService) CreateCrashAutoBet(autoBet *entities.CrashAutoBet) error {
	return s.Repo.CreateCrashAutoBet(autoBet)
}

func (s *CrashGameService) UpdateCrashAutoBetStatus(uid uint, status uint8) error {
	return s.Repo.UpdateCrashAutoBetStatus(uid, status)
}
