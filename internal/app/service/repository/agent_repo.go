package repository

import (
	"errors"
	"fmt"
	"rk-api/internal/app/entities"
	"rk-api/pkg/logger"
	"strings"
	"time"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

var AgentRepositorySet = wire.NewSet(wire.Struct(new(AgentRepository), "*"))

type AgentRepository struct {
	DB  *gorm.DB
	RDS redis.UniversalClient
}

// 创建
func (r *AgentRepository) CreateGameRebateReceipt(entity *entities.GameRebateReceipt) error {
	return r.DB.Create(entity).Error
}

func (r *AgentRepository) GetGameRebateReceiptList(param *entities.GetGameRebateReceiptListReq) error {
	var tx *gorm.DB = r.DB
	if param.UID != 0 {
		tx = tx.Where("uid = ?", param.UID)
	}
	param.List = make([]*entities.GameRebateReceipt, 0)
	return param.Paginate(tx)
}

func (r *AgentRepository) GetLevel1CountGrouByPID() ([]*entities.LevelCountGroup, error) {
	var list []*entities.LevelCountGroup // 创建一个Result类型的切片，用来存储查询结果
	// 开始构建查询
	if err := r.DB.Model(&entities.HallInviteRelation{}).Select("pid, count(1) as count").Where("level = ?", 1).Group("pid").Scan(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *AgentRepository) DelRelationByUID(uid uint) error {
	return r.DB.Where("uid = ?", uid).Limit(9).Delete(&entities.HallInviteRelation{}).Error
}

func (r *AgentRepository) GetRelationList(ID uint) ([]*entities.HallInviteRelation, error) {
	var tx *gorm.DB = r.DB
	tx = tx.Where("uid = ?", ID).Limit(9)
	list := make([]*entities.HallInviteRelation, 0)
	err := tx.Select("id", "pid", "level").Offset(0).Limit(9).Find(&list).Error
	return list, err
}

func (r *AgentRepository) GetRelation(entity *entities.HallInviteRelation) (*entities.HallInviteRelation, error) {
	result := r.DB.Last(&entity, entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return entity, nil
}

func (r *AgentRepository) AddRelationList(list []*entities.HallInviteRelation) error {
	return r.DB.CreateInBatches(list, len(list)).Error
}

func (r *AgentRepository) GetRakeBackList(game int) ([]*entities.RakeBack, error) {
	list := make([]*entities.RakeBack, 0)
	err := r.DB.Where("status = ? and game = ?", 1, game).Find(&list).Error
	return list, err
}

// 获取今日 PID 邀请人数
func (r *AgentRepository) GetTodayInviteCount(pid uint) (int64, error) {
	var count int64
	err := r.DB.Model(&entities.HallInviteRelation{}).
		Where("pid = ? AND level = ? AND DATE(FROM_UNIXTIME(created_at)) = CURDATE()", pid, 1).
		Count(&count).Error
	return count, err
}

// 获取直属人数
func (r *AgentRepository) GetImmediateInviteCount(pid uint) (int64, error) {
	var count int64
	err := r.DB.Model(&entities.HallInviteRelation{}).
		Where("pid = ?  AND level = ?", pid, 1).
		Count(&count).Error
	return count, err
}

// 获取今日有效投注
func (r *AgentRepository) GetTodayValidBet(PID uint) (float64, error) {
	var totalBet float64
	err := r.DB.Model(&entities.GameReturn{}).
		Where("pid = ? AND DATE(FROM_UNIXTIME(created_at)) = CURDATE()", PID).
		Select("IFNULL(SUM(Cash), 0)").Scan(&totalBet).Error // 如果为空，使用 0
	return totalBet, err
}

// 获取总有效投注
func (r *AgentRepository) GetTotalValidBet(PID uint) (float64, error) {
	var totalBet float64
	err := r.DB.Model(&entities.GameReturn{}).
		Where("pid = ? ", PID). // status 0 表示有效投注
		Select("IFNULL(SUM(Cash), 0)").Scan(&totalBet).Error
	return totalBet, err
}

func (r *AgentRepository) GetGameReturnCash(UID uint) (float64, error) {
	// 创建一个用于存放结果的结构体
	querySQL := "SELECT SUM(return_cash) AS num FROM game_return WHERE pid = ? AND status = 0"
	type Result struct {
		Num float64
	}
	var result Result
	if err := r.DB.Raw(querySQL, UID).Scan(&result).Error; err != nil {
		return 0, err
	}
	return float64(result.Num), nil
}

func (r *AgentRepository) GetGameReturnCashGroup(limit int) ([]*entities.ReturnCash, error) {
	// 创建一个用于存放结果的结构体
	querySQL := "SELECT SUM(return_cash) AS num,pid FROM game_return WHERE  status = 0 group by pid limit %d"
	// // 执行赢的SQL查询
	results := make([]*entities.ReturnCash, 0)

	if err := r.DB.Raw(fmt.Sprintf(querySQL, limit)).Scan(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

func (r *AgentRepository) GetGameCashAlreadyReturn(UID uint) (float64, error) {
	// 创建一个用于存放结果的结构体
	querySQL := "SELECT SUM(return_cash) AS num FROM game_return WHERE pid = ? AND status = 1"
	type Result struct {
		Num float64
	}
	var result Result
	if err := r.DB.Raw(querySQL, UID).Scan(&result).Error; err != nil {
		return 0, err
	}
	return float64(result.Num), nil
}

// 更新
func (r *AgentRepository) UpdateGameReturnStatusWithTx(tx *gorm.DB, UID uint) error {
	updateSQL := "UPDATE game_return SET status = 1, get_time = ? WHERE pid = ? AND status = 0"
	err := tx.Exec(updateSQL, time.Now().Unix(), UID).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *AgentRepository) GetRechargeReturn(uid uint) (*entities.RechargeReturn, error) {
	var entity entities.RechargeReturn
	result := r.DB.Where("uid = ?", uid).First(&entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &entity, nil
}

// func (r *AgentRepository) GetHallProfitReturn(entity *entities.HallProfitReturn) (*entities.HallProfitReturn, error) {
// 	result := r.DB.Last(&entity, entity)
// 	if result.Error != nil {
// 		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
// 			return nil, nil
// 		}
// 		return nil, result.Error
// 	}
// 	return entity, nil
// }

func (r *AgentRepository) GetRechargeCashAlreadyReturn(UID uint) (float64, error) {
	// 创建一个用于存放结果的结构体
	querySQL := "SELECT SUM(return_cash) AS num FROM recharge_return WHERE pid = ? AND status = 1"
	type Result struct {
		Num float64
	}
	var result Result
	if err := r.DB.Raw(querySQL, UID).Scan(&result).Error; err != nil {
		return 0, err
	}
	return float64(result.Num), nil
}

func (r *AgentRepository) GetMonthRechargeCashAlreadyReturn(UID uint) (float64, error) {
	// 创建一个用于存放结果的结构体
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	querySQL := "SELECT SUM(return_cash) AS num FROM recharge_return WHERE pid = ? AND status = 1 AND created_at > ?"
	type Result struct {
		Num float64
	}
	var result Result
	if err := r.DB.Raw(querySQL, UID, startOfMonth.Unix()).Scan(&result).Error; err != nil {
		return 0, err
	}
	return float64(result.Num), nil
}

func (r *AgentRepository) GetRechargeReturnByID(id uint) (*entities.RechargeReturn, error) {
	var entity entities.RechargeReturn
	// result := r.DB.Where("id = ?", id).First(&entity)

	result := r.DB.Clauses(dbresolver.Write).Where("id = ?", id).First(&entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &entity, nil
}

func (r *AgentRepository) CreateGameReturnWithTx(tx *gorm.DB, entity *entities.GameReturn) error {
	err := tx.Create(entity).Error
	return err
}

func (r *AgentRepository) UpdateGameReturnWithTx(tx *gorm.DB, entity *entities.GameReturn) error {
	err := tx.Updates(entity).Error
	return err
}

func (r *AgentRepository) CreateRechargeReturnWithTx(tx *gorm.DB, entity *entities.RechargeReturn) error {
	err := tx.Create(entity).Error
	return err
}

func (r *AgentRepository) UpdateRechargeReturnWithTx(tx *gorm.DB, entity *entities.RechargeReturn) error {
	err := tx.Updates(entity).Error
	return err
}

func (r *AgentRepository) GetLevelRelationList(uid uint) ([]*entities.LevelRelationInfo, error) {
	var results []*entities.LevelRelationInfo
	err := r.DB.Table("hall_invite_relation").Select("count(1) as num, level").Where("pid = ?", uid).Group("level").Scan(&results).Error
	return results, err
}

func (r *AgentRepository) GetPromotionRelationList(param *entities.GetPromotionListReq) error {
	var tx *gorm.DB = r.DB
	tx = tx.Table("hall_invite_relation").Select("return_cash, mobile,uid").Where("pid = ? AND level = ?", param.UID, param.Level).Order("return_cash desc")
	param.List = make([]*entities.HallInviteRelation, 0)
	err := param.Paginate(tx)
	if err != nil {
		return err
	}

	// var uids []uint
	// if relationList, ok := param.List.([]*entities.HallInviteRelation); ok {
	// 	for _, relation := range relationList {
	// 		uids = append(uids, relation.UID)
	// 	}
	// }

	// type Result struct {
	// 	Cash float64 `gorm:"column:cash"`
	// 	UID  uint    `gorm:"column:uid"`
	// }
	// var results []*Result

	// now := time.Now()
	// startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()) //当前月的

	// err = r.DB.Table("game_return").Select("sum(return_cash) as cash, uid").
	// 	Where("uid IN (?) and created_at > ?", uids, startOfMonth.Unix()).
	// 	Group("uid").
	// 	Find(&results).Error

	// if err != nil {
	// 	return err
	// }

	// if relationList, ok := param.List.([]*entities.HallInviteRelation); ok {
	// 	for _, relation := range relationList {
	// 		for _, result := range results {
	// 			if relation.UID == result.UID {
	// 				relation.ReturnCash = result.Cash
	// 				break
	// 			}
	// 		}
	// 	}
	// }

	return nil

}

func (r *AgentRepository) BatchUpdateRelationReturnCash(tx *gorm.DB, relations *[]*entities.HallInviteRelation) error {
	// // 使用事务批量更新所有用户
	// err := r.DB.Transaction(func(tx *gorm.DB) error {
	// 	for _, relation := range *relations {
	// 		if err := tx.Table("hall_invite_relation").Where("id = ?", relation.ID).Update("return_cash", relation.ReturnCash).Error; err != nil {
	// 			// 返回任何错误会回滚事务
	// 			return err
	// 		}
	// 	}
	// 	// 返回 nil 提交事务
	// 	return nil
	// })
	// return err

	sql := GenerateUpdateSQL("hall_invite_relation", "return_cash", relations)

	// logger.ZError("BatchUpdateRelationReturnCash", zap.Any("relation", relations))

	logger.Info("-------------BatchUpdateRelationReturnCash------------------", sql)
	return tx.Exec(sql).Error
}

// 直接等于
func GenerateUpdateSQL2(tableName string, keyName string, relations *[]*entities.HallInviteRelation) string {
	// 存储 CASE 中的 WHEN-THEN 语句部分
	var caseParts []string
	// 存储所有用户的 UID
	var uids []string

	// 遍历 users 列表构造 SQL 语句
	for _, relation := range *relations {
		caseParts = append(caseParts, fmt.Sprintf("WHEN %d THEN %f", relation.ID, relation.ReturnCash))
		uids = append(uids, fmt.Sprintf("%d", relation.ID))
	}
	// JOIN 当所有的 WHEN-THEN 语句并构建完整的 CASE 语句
	caseStatement := strings.Join(caseParts, " ")
	// JOIN UID 列表用于 WHERE IN 语句
	uidList := strings.Join(uids, ", ")

	// 构造并返回完整的 SQL 语句
	return fmt.Sprintf(`UPDATE %s SET %s = CASE id %s ELSE cash END WHERE id IN (%s);`, tableName, keyName, caseStatement, uidList)
}

// 在原先基础上加
func GenerateUpdateSQL(tableName string, keyName string, relations *[]*entities.HallInviteRelation) string {
	var caseParts []string
	var uids []string

	for _, relation := range *relations {
		casePart := fmt.Sprintf("WHEN %d THEN %s + %f", relation.ID, keyName, relation.ReturnCash)
		caseParts = append(caseParts, casePart)
		uids = append(uids, fmt.Sprintf("%d", relation.ID))
	}

	caseStatement := strings.Join(caseParts, " ")
	uidList := strings.Join(uids, ", ")

	return fmt.Sprintf("UPDATE %s SET %s = CASE id %s ELSE %s END WHERE id IN (%s);", tableName, keyName, caseStatement, keyName, uidList)
}
