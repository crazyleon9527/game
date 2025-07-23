package repository

import (
	"rk-api/internal/app/entities"

	"github.com/google/wire"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

var FlowRepositorySet = wire.NewSet(wire.Struct(new(FlowRepository), "*"))

type FlowRepository struct {
	DB *gorm.DB
}

func (r *FlowRepository) CreateFlow(flow *entities.Flow) error {
	return r.DB.Create(flow).Error
}

func (r *FlowRepository) CreateRefundGameFlow(flow *entities.RefundGameFlow) error {
	return r.DB.Create(flow).Error
}

func (r *FlowRepository) CreateRefundLinkGameFlow(flow *entities.RefundLinkGameFlow) error {
	return r.DB.Create(flow).Error
}

func (r *FlowRepository) GetUnProceedRefundGameFlow(flowType, limit int) ([]*entities.RefundGameFlow, error) {
	list := make([]*entities.RefundGameFlow, 0)
	err := r.DB.Clauses(dbresolver.Write).Where("status = ? and type = ? and number < ?", 0, flowType, 0).Select("id", "uid", "number", "type", "pc").Limit(limit).Find(&list).Error

	// err := r.DB.Clauses(dbresolver.Write).Where(" type = ? and status = ?", flowType, 0).Select("id", "uid", "number", "type", "pc").Limit(limit).Find(&list).Error
	return list, err
}

func (r *FlowRepository) GetUnProceedGameRecords(limit int) ([]*entities.GameRecordRefund, error) {
	list := make([]*entities.GameRecordRefund, 0)
	err := r.DB.Clauses(dbresolver.Write).Where("status = ?", 1).Select("id", "uid", "amount", "pc").Limit(limit).Find(&list).Error
	return list, err
}

func (r *FlowRepository) ProceedRefundGameFlowListWithTx(tx *gorm.DB, processedUids []uint) error {
	if err := tx.Model(&entities.RefundGameFlow{}).Where("uid IN (?)", processedUids).Update("status", 1).Error; err != nil {
		return err
	}
	return nil
}

func (r *FlowRepository) GetFlowList(param *entities.GetFlowListReq) error {
	var tx *gorm.DB = r.DB
	tx = tx.Where("uid = ?", param.UID)
	param.List = make([]*entities.Flow, 0)
	return param.Paginate(tx)
}

// func (r *FlowRepository) GetLastestFlowPeriodByDate(periodDate string, betType uint8) (*entities.FlowPeroid, error) {

// 	var period entities.FlowPeroid
// 	result := r.DB.Where("period_date = ? and bet_type = ?", periodDate, betType).Last(&period)
// 	if result.Error != nil {
// 		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
// 			return nil, nil
// 		}
// 		return nil, result.Error
// 	}
// 	return &period, nil
// }

// func CreatePresetRDSKey(periodDate string) string {
// 	return fmt.Sprintf(constant.REDIS_Flow_PRESET, periodDate)
// }

// func (r *FlowRepository) GetPresetNumberListRDS(periodDate string) ([]int, error) {
// 	res, err := r.RDS.Get(context.Background(), CreatePresetRDSKey(periodDate)).Result()
// 	var list []int
// 	err = json.Unmarshal([]byte(res), &list)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return list, nil
// }

// func (r *FlowRepository) AddPresetNumberListRDS(periodDate string, list []int) error {
// 	jsonArr, err := json.Marshal(list)
// 	if err != nil {
// 		return err
// 	}
// 	err = r.RDS.Set(context.Background(), CreatePresetRDSKey(periodDate), jsonArr, 0).Err()
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (r *FlowRepository) GetFlowSettingList() ([]*entities.FlowRoomSetting, error) {
// 	list := make([]*entities.FlowRoomSetting, 0)
// 	err := r.DB.Where("status = ?", 1).Find(&list).Error
// 	return list, err
// }

// func (r *FlowRepository) GetFlowSetting(betType uint) (*entities.FlowRoomSetting, error) {
// 	var setting entities.FlowRoomSetting
// 	result := r.DB.Where("bet_type = ? and status = ?", betType, 1).Last(&setting)
// 	if result.Error != nil {
// 		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
// 			return nil, nil
// 		}
// 		return nil, result.Error
// 	}
// 	return &setting, nil
// }
