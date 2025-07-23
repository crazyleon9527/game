package service

import (
	"fmt"
	"rk-api/internal/app/constant"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/service/repository"
	"rk-api/pkg/logger"
	"rk-api/pkg/structure"

	"github.com/google/wire"
	"go.uber.org/zap"
)

var FlowServiceSet = wire.NewSet(
	ProvideFlowService,
)

// var FlowServiceSet = wire.NewSet(wire.Struct(new(FlowService), "*"))

var TypeRemarks = map[uint16]string{
	// constant.FLOW_TYPE_WINGO:                   "wingo下注",       //wingo下注
	// constant.FLOW_TYPE_WINGO_REWARD:            "wingo奖励",       //wingo 奖励
	// constant.FLOW_TYPE_NINE:                    "九星下注",          //九星下注
	// constant.FLOW_TYPE_NINE_REWARD:             "九星奖励",          //九星奖励
	// constant.FLOW_TYPE_RETURN_CASH:             "游戏返利",          //游戏返利
	// constant.FLOW_TYPE_RECHARGE_RETURN_CASH:    "充值返利",          //充值返利
	// constant.FLOW_TYPE_GET_RED_ENVELOPE:        "领取红包",          //红包
	// constant.FLOW_TYPE_RECHARGE_CASH:           "充值",            //充值
	// constant.FLOW_TYPE_RECHARGE_ACT_10000:      "充值10000 送2000", //充值10000 送2000
	// constant.FLOW_TYPE_INTEREST:                "每日利息",          //每日利息
	// constant.FLOW_TYPE_PINDUO:                  "拼多多奖励",         //拼多多奖励
	// constant.FLOW_TYPE_APPLY_FOR_WITHDRAW_CASH: "申请提现扣除",        //申请提现
	// constant.FLOW_TYPE_GM_CASH:                 "GM 改动",         //GM 改动
	// constant.FLOW_TYPE_WITHDRAW_LOCK_CASH:      "提现锁定金额返还",      //提现锁定金额返回

	constant.FLOW_TYPE_WINGO:                   "wingo bet",
	constant.FLOW_TYPE_WINGO_REWARD:            "wingo reward",
	constant.FLOW_TYPE_NINE:                    "ninestar bet",
	constant.FLOW_TYPE_NINE_REWARD:             "ninestar reward",
	constant.FLOW_TYPE_RETURN_CASH:             "game return cash",
	constant.FLOW_TYPE_RECHARGE_RETURN_CASH:    "return recharge cash for agent",
	constant.FLOW_TYPE_GET_RED_ENVELOPE:        "red envelope",
	constant.FLOW_TYPE_RECHARGE_CASH:           "recharge cash",
	constant.FLOW_TYPE_RECHARGE_ACT_10000:      "recharge 10000 give 2000",
	constant.FLOW_TYPE_INTEREST:                "daily interest",
	constant.FLOW_TYPE_PINDUO:                  "pinduoduo reward",
	constant.FLOW_TYPE_APPLY_FOR_WITHDRAW_CASH: "apply for withdraw cash",
	constant.FLOW_TYPE_GM_CASH:                 "gm send",
	constant.FLOW_TYPE_WITHDRAW_LOCK_CASH:      "withdraw lock cash return",

	// constant.FLOW_TYPE_R8_WITHDRAW:        "Rich88 提款",
	// constant.FLOW_TYPE_R8_WITHDRAW_POKDEN: "Rich88 POKDEN提走",
	// constant.FLOW_TYPE_R8_DEPOSIT:         "Rich88 退款",
	// constant.FLOW_TYPE_R8_ROLLBACK:        "Rich88 回滚",
	// constant.FLOW_TYPE_R8_ACTIVITY_AWARD:  "Rich88 活动奖励",
	// constant.FLOW_TYPE_ZF_BET:             "智峰投注",
	// constant.FLOW_TYPE_ZF_PAYOUT:          "智峰派彩",
	// constant.FLOW_TYPE_ZF_REFUND:          "智峰退款",
	// constant.FLOW_TYPE_ZF_PAYOUT_FAIL:     "智峰派彩失败",
	// constant.FLOW_TYPE_ZF_CANCEL:          "智峰取消",
}

type FlowService struct {
	Repo         *repository.FlowRepository
	UserSrv      *UserService
	AgentSrv     *AgentService
	StateSrv     *StateService
	InGameReturn bool
}

func ProvideFlowService(
	repo *repository.FlowRepository,
	userSrv *UserService,
	agentSrv *AgentService,
	stateSrv *StateService,
) *FlowService {
	service := &FlowService{
		Repo:     repo,
		UserSrv:  userSrv,
		AgentSrv: agentSrv,
		StateSrv: stateSrv,
	}
	return service
}

func (s *FlowService) CreateFlow(flow *entities.Flow) (err error) {
	if flow.Number == 0 { //为0的不做处理
		return nil
	}

	if remark, ok := TypeRemarks[flow.FlowType]; ok {
		flow.Remark = remark
	}

	defer func() {
		if err != nil {
			logger.ZError("createFlow", zap.Any("flow", flow), zap.Error(err))
		}
	}()

	if err = s.Repo.CreateFlow(flow); err != nil { //创建流水
		logger.ZError("CreateFlowWithTx fail",
			zap.Any("flow", flow),
			zap.Error(err),
		)
		return
	}

	if flow.FlowType > 200 { //表示游戏
		if flow.FlowType < 300 { //内部游戏
			refundFlow := new(entities.RefundGameFlow)
			structure.Copy(flow, refundFlow)
			refundFlow.FlowID = flow.ID
			err = s.Repo.CreateRefundGameFlow(refundFlow)
			return
		} else { //外部链接游戏
			refundFlow := new(entities.RefundLinkGameFlow)
			structure.Copy(flow, refundFlow)
			refundFlow.FlowID = flow.ID
			refundFlow.ID = 0
			err = s.Repo.CreateRefundLinkGameFlow(refundFlow)
			return
		}
	}
	return nil
}

type PlayerRecord struct {
	UID          uint //用户ID
	PromoterCode int
	Number       float64
}

func (b *PlayerRecord) add(number float64) *PlayerRecord {
	if number < 0 {
		b.Number -= number //转正值
	}
	return b
}

// 获取返利列表
func (b *PlayerRecord) getHPRList(relations *[]*entities.HallInviteRelation, rbMap map[uint8]uint, game int) []*entities.GameReturn {
	list := make([]*entities.GameReturn, 0)
	remark := fmt.Sprintf("%d", game)
	for _, relation := range *relations {
		percent := int(rbMap[relation.Level])
		// returnCash := b.Number * 0.05 * float64(percent) / 1000 //基于手续费  0.05 计算
		returnCash := b.Number * float64(percent) / 10000 //基于手续费  万分之5
		if returnCash > constant.PreciseZero {
			relation.ReturnCash = returnCash
			hpr := &entities.GameReturn{
				UID:          b.UID,
				Cash:         b.Number,
				Level:        relation.Level,
				PID:          relation.PID,
				Percent:      percent,
				ReturnCash:   returnCash,
				Remark:       remark,
				PromoterCode: b.PromoterCode,
			}
			list = append(list, hpr)
		}

	}

	return list
}

type PlayerRecordGroup struct {
	players map[uint]*PlayerRecord
}

func NewPlayerFlowGroup() *PlayerRecordGroup {
	return &PlayerRecordGroup{
		players: make(map[uint]*PlayerRecord),
	}
}

func (p *PlayerRecordGroup) get(uid uint) *PlayerRecord {
	if _, ok := p.players[uid]; ok {
		return p.players[uid]
	}
	return nil
}

func (p *PlayerRecordGroup) add(pf *PlayerRecord) {
	p.players[pf.UID] = pf
}

func (p *PlayerRecordGroup) getList() []*PlayerRecord {
	list := make([]*PlayerRecord, 0)
	for _, v := range p.players {
		list = append(list, v)
	}
	return list
}

func (s *FlowService) ProcessRefundGameFlow(limit int) error {

	if s.StateSrv.GetBoolState(constant.StateGameBetAreaLimit) { //处理游戏下注限制
		logger.ZError("ProcessRefundGameFlow", zap.Bool("StateGameBetAreaLimit", s.StateSrv.GetBoolState(constant.StateGameBetAreaLimit)))
		return nil
	}

	if s.StateSrv.GetBoolState(constant.StateMonthBackupAndClean) { //处在月度备份和清理状态
		logger.ZError("ProcessRefundGameFlow", zap.Bool("StateMonthBackupAndClean", s.StateSrv.GetBoolState(constant.StateMonthBackupAndClean)))
		return nil
	}
	if s.StateSrv.GetBoolState(constant.StateChangePC) { //处在更换PC，业务员合并
		logger.ZError("ProcessRefundGameFlow", zap.Bool("StateChangePC", s.StateSrv.GetBoolState(constant.StateChangePC)))
		return nil
	}

	if s.InGameReturn { //必须等待执行完
		logger.ZError("ProcessRefundGameFlow", zap.Bool("InGameReturn", s.InGameReturn))
		return nil
	}

	s.InGameReturn = true
	logger.ZInfo("ProcessRefundGameFlow")
	// processFlowTypeList := []int{201, 211} //wingo,nine
	// for _, flowType := range processFlowTypeList {
	// 	if err := s.ProcessRefundGameFlowByFlowType(limit, flowType); err != nil {
	// 		logger.ZError("ProcessRefundGameFlowByFlowType", zap.Int("flowType", flowType), zap.Error(err))
	// 	}
	// }
	// if err := s.ProcessRefundGameRecord(limit); err != nil {
	// 	logger.ZError("ProcessRefundGameRecord", zap.Int("limit", limit), zap.Error(err))
	// }

	s.InGameReturn = false
	return nil
}

// 获取返利列表
func (s *FlowService) GetFlowList(req *entities.GetFlowListReq) error {
	return s.Repo.GetFlowList(req)
}
