package service

import (
	"fmt"
	"rk-api/internal/app/constant"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/errors"
	"rk-api/internal/app/mq"
	"rk-api/internal/app/mq/handle"
	"rk-api/internal/app/utils"

	"rk-api/internal/app/service/repository"
	"rk-api/pkg/cjson"
	"rk-api/pkg/logger"
	"rk-api/pkg/structure"
	"strconv"
	"strings"
	"time"

	"github.com/google/wire"
	"github.com/orca-zhang/ecache"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var AgentServiceSet = wire.NewSet(
	ProvideAgentService,
)

type AgentService struct {
	Repo             *repository.AgentRepository
	UserSrv          *UserService
	AdminSrv         *AdminService
	relationPIDCache *ecache.Cache
	rakeCache        *ecache.Cache
	WalletSrv        *WalletService
}

func ProvideAgentService(repo *repository.AgentRepository,
	userSrv *UserService,
	adminSrv *AdminService,
	fundSrv *WalletService) *AgentService {

	// 返回你的RechargeService实例
	return &AgentService{
		Repo:             repo,
		UserSrv:          userSrv,
		AdminSrv:         adminSrv,
		WalletSrv:        fundSrv,
		rakeCache:        ecache.NewLRUCache(2, 8, 12*time.Hour),    //初始化缓存
		relationPIDCache: ecache.NewLRUCache(30, 100, 12*time.Hour), //初始化缓存
	}

}

func (s *AgentService) GetGameRebateReceiptList(param *entities.GetGameRebateReceiptListReq) error {
	return s.Repo.GetGameRebateReceiptList(param)
}

// 修复一级邀请数目
func (s *AgentService) FixLevel1InviteCount() ([]*entities.LevelCountGroup, error) {
	list, err := s.Repo.GetLevel1CountGrouByPID()
	if err != nil {
		return nil, err
	}
	for _, group := range list {
		if group.Count > 1 { //对邀请数目大于1的情形处理，修复之前bug产生的问题
			invite, _ := s.UserSrv.GetUserByUID(group.PID)
			if invite != nil {
				inviteForUpdate := new(entities.User)
				inviteForUpdate.ID = invite.ID
				inviteForUpdate.InviteCount = uint(group.Count) //邀请的一级增加一个
				if err := s.UserSrv.UpdateUser(inviteForUpdate); err != nil {
					return nil, err
				}
			}
		}
	}
	return list, err
}

func (s *AgentService) GetInviteRelationList(uid uint) ([]*entities.HallInviteRelation, error) {

	cacheKey := fmt.Sprintf("%d", uid)
	if val, ok := s.relationPIDCache.Get(cacheKey); ok {
		return val.([]*entities.HallInviteRelation), nil
	}

	list, err := s.Repo.GetRelationList(uid)
	if err != nil {
		return nil, err
	}
	s.relationPIDCache.Put(cacheKey, list) //丢入缓存中 下次从缓存读取

	return list, nil
}

// 修改代理关系
func (s *AgentService) FixInviteRelation(req *entities.FixInviteRelationReq) (err error) {

	user, err := s.UserSrv.GetUserByUID(req.UID)
	if err != nil {
		return err
	}
	var parent *entities.User
	if req.PID != 0 {
		parent, err = s.UserSrv.GetUserByUID(req.PID)
		if err != nil {
			return err
		}
		if user.Inviter == parent.Mobile {
			return errors.WithCode(errors.InviteRelationExist)
		}
	}
	userHistory := new(entities.User)
	structure.Copy(user, userHistory)

	defer func() {
		log := entities.SystemOptionLog{
			Type:     constant.SYS_OPTION_TYPE_INVITE_RELATION_FIX,
			OptionID: req.OptionID,
			IP:       req.IP,
			Content:  cjson.StringifyIgnore(userHistory),
		}
		if err != nil {
			log.Result = "false"
			log.Remark = "修改用户邀请关系失败"
			logger.ZError("ReviewWithdrawal fail", zap.Any("req", req), zap.Error(err))
		} else {
			log.Result = "true"
			log.Remark = "修改用户邀请关系成功"
		}

		optionLogQueue, _ := handle.NewOptionLogQueue(&log)
		if _, err := mq.MClient.Enqueue(optionLogQueue); err != nil {
			logger.ZError("optionLogQueue fail", zap.Error(err))
		}
	}()
	if user.Inviter != "" { //处理之前邀请
		invite, _ := s.UserSrv.GetUserByname(user.Inviter)
		if invite != nil {
			inviteForUpdate := new(entities.User)
			inviteForUpdate.ID = invite.ID
			inviteForUpdate.InviteCount -= 1 //原先邀请的一级减少一个
			if inviteForUpdate.InviteCount < 0 {
				inviteForUpdate.InviteCount = 0
			}

			if err = s.UserSrv.UpdateUser(inviteForUpdate); err != nil {
				return err
			}
		}

		if err = s.Repo.DelRelationByUID(user.ID); err != nil {
			return err
		}
	}

	if req.PromotionCode != 0 { //业务员邀请
		sysUser, err := s.AdminSrv.GetSysUser(&entities.SysUser{UID: int(req.PromotionCode)}) //用户的分销人处理
		if err != nil {
			return errors.WithCode(errors.InvalidPromotionCode) //
		}
		user.Promoter = sysUser.Username //
		user.PromoterCode = int(req.PromotionCode)
		user.Inviter = "" //置空
	}

	if parent != nil { //玩家邀请
		user.Promoter = parent.Promoter // 设置为父亲的推广信息
		user.PromoterCode = parent.PromoterCode
		user.Inviter = parent.Mobile
	}

	userForUpdate := entities.User{
		Promoter:     user.Promoter,
		PromoterCode: user.PromoterCode,
		Inviter:      user.Inviter, //邀请者手机号
	}
	userForUpdate.ID = user.ID

	if err = s.UserSrv.UpdateUser(&userForUpdate); err != nil {
		return err
	}

	if parent != nil {
		inviteForUpdate := new(entities.User)
		inviteForUpdate.ID = parent.ID
		inviteForUpdate.InviteCount += 1 //邀请的一级增加一个

		if err = s.UserSrv.UpdateUser(inviteForUpdate); err != nil {
			return err
		}
		relation := &entities.HallInviteRelation{
			UID:    user.ID,
			Mobile: user.Mobile,
			PID:    parent.ID,
			Level:  1, //1级
		}
		if err = s.FillInviteRelation(relation); err != nil {
			return err
		}
	}

	return nil
}

func (s *AgentService) FillInviteRelation(relation *entities.HallInviteRelation) error {

	list, err := s.Repo.GetRelationList(relation.PID)
	if err != nil {
		return err
	}
	relation.Level = 1 //一级父亲
	newList := []*entities.HallInviteRelation{relation}
	for _, pRelation := range list { //再把当前父亲加进去
		pRelation.ID = 0 //重置为0 ，方便作为新的关系添加
		pRelation.Level++
		pRelation.Mobile = relation.Mobile
		pRelation.UID = relation.UID
		pRelation.ReturnCash = 0
		newList = append(newList, pRelation)
		if pRelation.Level >= constant.AGENT_LEVEL_MAX { //最多9级
			break
		}
	}

	if err = s.Repo.AddRelationList(newList); err != nil {
		return err
	}

	logger.ZInfo("invite relation",
		zap.Uint("pid", relation.PID),
		zap.Uint("uid", relation.UID),
		zap.String("mobile", relation.Mobile),
	)

	return nil

}

func (s *AgentService) GetRakeBackMap(game int) (map[uint8]uint, error) {

	cacheKey := fmt.Sprintf("%d", game)
	if val, ok := s.rakeCache.Get(cacheKey); ok {
		return val.(map[uint8]uint), nil
	}

	list, err := s.Repo.GetRakeBackList(game)
	if err != nil {
		return nil, err
	}
	rbMap := make(map[uint8]uint, 0)
	for _, rb := range list {
		rbMap[rb.Level] = rb.Value
	}

	s.rakeCache.Put(cacheKey, rbMap) //丢入缓存中 下次从缓存读取

	return rbMap, nil
}

// 每次处理Limit 名玩家的游戏返现
func (s *AgentService) BatchFinalizeGameCashReturn(limit int) error {
	list, err := s.Repo.GetGameReturnCashGroup(limit)
	if err != nil {
		return err
	}
	for _, rc := range list {
		user, err := s.UserSrv.GetUserByUID(uint(rc.PID))
		if err != nil {
			return err
		}
		err = s.dealGameReturnCashForUser(user, rc.Cash)
		if err != nil {
			return err
		}
	}
	return nil
}

// 领取return cash
func (s *AgentService) FinalizeGameCashReturn(uid uint) error {

	//判断用户是否存在
	returnCash, err := s.Repo.GetGameReturnCash(uid)
	if err != nil {
		return err
	}

	if returnCash <= 0 { //没有返利金额
		return nil
	}

	user, err := s.UserSrv.GetUserByUID(uid)
	if err != nil {
		return err
	}

	err = s.dealGameReturnCashForUser(user, returnCash)
	if err != nil {
		return err
	}
	go func() {
		defer utils.PrintPanicStack()
		s.Repo.CreateGameRebateReceipt(&entities.GameRebateReceipt{
			UID:         user.ID,
			Cash:        returnCash,
			Status:      1, //已经处理
			ReceiveTime: time.Now().Unix(),
		})
	}()

	return nil
}

func (s *AgentService) dealGameReturnCashForUser(user *entities.User, returnCash float64) error {
	if returnCash <= 0 { //没有返利金额
		return nil
	}

	err := s.WalletSrv.HandleWallet(user.ID, func(wallet *entities.UserWallet, tx *gorm.DB) error {
		wallet.SafeAdjustCash(returnCash) //增加金额
		if err := s.Repo.UpdateGameReturnStatusWithTx(tx, user.ID); err != nil {
			return err
		}

		logger.ZInfo("dealGameReturnCashForUser UpdateUserWithTx", zap.Uint("uid", wallet.UID), zap.Float64("returnCash", returnCash), zap.Float64("balance", wallet.Cash))
		if err := s.WalletSrv.UpdateCashWithTx(tx, wallet); err != nil {
			return err
		}

		createFlowQueue, _ := handle.NewCreateFlowQueue(&entities.Flow{
			UID:          user.ID,
			FlowType:     constant.FLOW_TYPE_RETURN_CASH,
			Number:       returnCash,
			Balance:      wallet.Cash,
			PromoterCode: user.PromoterCode,
		})
		if _, err := mq.MClient.Enqueue(createFlowQueue); err != nil {
			logger.ZError("createFlowQueue", zap.Any("flow", createFlowQueue), zap.Error(err))
		}
		return nil
	})

	return err
}

// 按充值返回对应的
func getLevelReturnCash(cash float64) float64 {
	if cash >= 100000 {
		return 20000
	} else if cash >= 50000 {
		return 5500
	} else if cash >= 10000 {
		return 1500
	} else if cash >= 5000 {
		return 1000
	} else if cash >= 3000 {
		return 600
	} else if cash >= 1000 {
		return 300
	} else if cash >= 500 {
		return 150
	} else if cash >= 200 {
		return 20
	}
	return 0
}

func (s *AgentService) GetMonthRechargeCashAlreadyReturn(uid uint) (float64, error) {
	return s.Repo.GetMonthRechargeCashAlreadyReturn(uid)
}

// 领取return cash
func (s *AgentService) FinalizeRechargeCashReturn(req *entities.FinalizeRechargeReturnReq) error {

	profitReturn, err := s.Repo.GetRechargeReturnByID(req.ID)
	if err != nil {
		return err
	}
	if profitReturn == nil { //没有返利金额
		return errors.WithCode(errors.RechargeReturnNotExist)
	}
	if profitReturn.Status != 0 {
		return errors.WithCode(errors.InvalidRechargeReturn)
	}

	err = s.WalletSrv.HandleWallet(profitReturn.PID, func(wallet *entities.UserWallet, tx *gorm.DB) error {
		wallet.SafeAdjustCash(profitReturn.ReturnCash) //增加金额
		profitReturn.Status = 1                        //已经处理

		returnForUpdate := entities.RechargeReturn{
			Status:  1,
			GetTime: time.Now().Unix(),
		}
		returnForUpdate.ID = profitReturn.ID
		returnForUpdate.GetTime = time.Now().Unix()
		if err := s.Repo.UpdateRechargeReturnWithTx(tx, &returnForUpdate); err != nil {
			return err
		}

		logger.ZInfo("FinalizeRechargeCashReturn UpdateUserWithTx", zap.Any("user", wallet))
		if err := s.WalletSrv.UpdateCashWithTx(tx, wallet); err != nil {
			return err
		}
		createFlowQueue, _ := handle.NewCreateFlowQueue(&entities.Flow{
			UID:          wallet.UID,
			FlowType:     constant.FLOW_TYPE_RECHARGE_RETURN_CASH,
			Number:       profitReturn.ReturnCash,
			Balance:      wallet.Cash,
			PromoterCode: wallet.PromoterCode,
		})
		if _, err := mq.MClient.Enqueue(createFlowQueue); err != nil {
			logger.ZError("createFlowQueue", zap.Any("flow", createFlowQueue), zap.Error(err))
		}
		return nil
	})

	return nil
}

func (s *AgentService) AutoFinalizeRechargeCashReturn(id uint) error {

	return s.FinalizeRechargeCashReturn(&entities.FinalizeRechargeReturnReq{
		ID:       id,
		OptionID: 1,
		IP:       "0.0.0.0",
	})
}

// 检测和处理充值返利
func (s *AgentService) CheckReturnRechargeCash(uid uint, cash float64, orderID string) error {
	user, err := s.UserSrv.GetUserByUID(uid)
	if err != nil {
		return err
	}
	hpr, err := s.Repo.GetRechargeReturn(uid)
	if err != nil {
		return err
	}
	if hpr != nil {
		return errors.With("already exit recharge cash") //只允许一次
	}

	levelCash := getLevelReturnCash(cash)
	// if levelCash <= 0 {
	// 	return nil
	// }

	relation, err := s.Repo.GetRelation(&entities.HallInviteRelation{
		Level: 1,
		UID:   user.ID,
	})
	if err != nil {
		return err
	}
	if relation == nil {
		return nil
	}

	hpr = &entities.RechargeReturn{
		UID:          user.ID,
		PromoterCode: user.PromoterCode,
		Cash:         cash,
		Level:        relation.Level,
		PID:          relation.PID,
		Percent:      0,
		ReturnCash:   levelCash,
		Remark:       orderID,
		Status:       0,
	}

	if hpr.ReturnCash == 0 {
		hpr.Status = 1
	}

	tx := s.Repo.DB.Begin()
	if err := s.Repo.CreateRechargeReturnWithTx(tx, hpr); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()

	logger.ZInfo("CreateReturnRechargeCash", zap.Any("hpr", hpr))

	go func() {
		defer utils.PrintPanicStack()

		if err := s.AutoFinalizeRechargeCashReturn(hpr.ID); err != nil {
			logger.ZError("AutoFinalizeRechargeCashReturn", zap.Uint("id", hpr.ID), zap.Error(err))
		} else {
			logger.ZInfo("AutoFinalizeRechargeCashReturn", zap.Uint("id", hpr.ID))
		}
	}()

	return nil
}

func (s *AgentService) GetPromotionProfit(uid uint) (*entities.PromotionProfit, error) {
	user, err := s.UserSrv.GetUserByUID(uid)
	if err != nil {
		return nil, err
	}

	if user.InviteCode == "" { //没有邀请码则生成邀请码
		user.InviteCode = s.createInviteCode(user.GetUserID())
		userForUpdate := new(entities.User)
		userForUpdate.ID = user.ID
		userForUpdate.InviteCode = user.InviteCode
		s.UserSrv.UpdateUser(userForUpdate)
	}

	forReturn, err := s.Repo.GetGameReturnCash(user.ID) //查询还未领取的
	if err != nil {
		return nil, err
	}

	// alreadyReturn, err := s.Repo.GetGameCashAlreadyReturn(user.ID) //查询已经领取的
	// if err != nil {
	// 	return nil, err
	// }

	promotionProfit := new(entities.PromotionProfit)
	promotionProfit.Code = user.InviteCode
	// promotionProfit.HasGet = alreadyReturn
	promotionProfit.NotGet = forReturn
	promotionProfit.LevelMap = make(map[string]int)

	todayValidBet, err := s.Repo.GetTodayValidBet(user.ID)
	if err != nil {
		return nil, err
	}
	totalValidBet, err := s.Repo.GetTotalValidBet(user.ID)
	if err != nil {
		return nil, err
	}
	promotionProfit.TodayValidBet = todayValidBet
	promotionProfit.TotalValidBet = totalValidBet
	promotionProfit.InviteCount = int(user.InviteCount) //暂时不用

	todayInvite, err := s.Repo.GetTodayInviteCount(user.ID)
	if err != nil {
		return nil, err
	}
	promotionProfit.TodayLink = int(todayInvite)
	// list, err := s.Repo.GetLevelRelationList(uid)
	// if err != nil {
	// 	return nil, err
	// }

	// for _, rl := range list {
	// 	promotionProfit.InviteCount += rl.Num
	// 	// promotionProfit.LevelMap[fmt.Sprintf("level%d", rl.Level)] = rl.Num
	// }

	return promotionProfit, nil
}

// 创建邀请码
func (s *AgentService) createInviteCode(logicId string) string {
	if len(logicId) < 4 {
		return ""
	}
	code := ""
	config := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}

	end := logicId[3:]
	pre := logicId[0:3]
	preArr := strings.Split(pre, "")

	for _, v := range preArr {
		intValue, err := strconv.Atoi(v)
		if err != nil {
			return ""
		}
		code += config[intValue]
	}

	code += end
	return code
}

// 获取推广列表
func (s *AgentService) GetPromotionList(param *entities.GetPromotionListReq) error {
	err := s.Repo.GetPromotionRelationList(param)
	return err
}

// 获取推广链接
func (s *AgentService) GetPromotionLink(uid uint) (*entities.PromotionLink, error) {
	user, err := s.UserSrv.GetUserByUID(uid)
	if err != nil {
		return nil, err
	}
	if user.InviteCode == "" { //没有邀请码则生成邀请码
		user.InviteCode = s.createInviteCode(user.GetUserID())
		userForUpdate := new(entities.User)
		userForUpdate.ID = user.ID
		userForUpdate.InviteCode = user.InviteCode
		if err := s.UserSrv.UpdateUser(userForUpdate); err != nil {
			return nil, err
		}
	}
	pl := &entities.PromotionLink{
		// PromotionCode: user.P,
		InviteCode: user.InviteCode,
	}
	return pl, nil
}

func (s *AgentService) BatchUpdateRelationReturnCash(tx *gorm.DB, relations *[]*entities.HallInviteRelation) error {
	return s.Repo.BatchUpdateRelationReturnCash(tx, relations)
}
