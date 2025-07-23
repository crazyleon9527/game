package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"rk-api/internal/app/constant"
	"rk-api/internal/app/entities"
	"rk-api/pkg/logger"
	"strconv"
	"strings"
	"time"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var StatsRepositorySet = wire.NewSet(wire.Struct(new(StatsRepository), "*"))

type StatsRepository struct {
	DB  *gorm.DB
	RDS redis.UniversalClient
}

func (r *StatsRepository) CreateRankStats(limit int, promoteCode int) error {
	// // 定义两个SQL查询变量
	typeList := []int{constant.FLOW_TYPE_WINGO, constant.FLOW_TYPE_WINGO_REWARD, constant.FLOW_TYPE_NINE, constant.FLOW_TYPE_NINE_REWARD}
	stringTypeList := make([]string, len(typeList))
	for i, val := range typeList {
		stringTypeList[i] = strconv.Itoa(val)
	}
	condition := fmt.Sprintf("type in (%s)", strings.Join(stringTypeList, ","))

	if promoteCode != 0 {
		condition += fmt.Sprintf(" and pc = %d", promoteCode)
	}

	sql := `SELECT SUM(number) AS profit, uid,pc  FROM flow where %s GROUP BY uid,pc ORDER BY profit %s LIMIT %d`

	sqlWin := fmt.Sprintf(sql, condition, "DESC", limit)

	logger.ZInfo("CreateRankStats", zap.String("sql", sqlWin))
	// // 执行赢的SQL查询
	winRanks := make([]entities.RankStats, 0)
	err := r.DB.Raw(sqlWin).Scan(&winRanks).Error
	if err != nil {
		logger.ZError("CreateRankStats", zap.String("sql", sqlWin), zap.Error(err))
		return err
	}

	if len(winRanks) > 0 {
		r.DB.Where("code = ?", promoteCode).Delete(&entities.RankStats{})
	}

	// 同批量导入中赢的排名
	r.DB.CreateInBatches(winRanks, len(winRanks))

	sqlLose := fmt.Sprintf(sql, condition, "ASC", limit)
	logger.ZInfo("CreateRankStats", zap.String("sql", sqlLose))
	// // 执行输的SQL查询
	loseRanks := make([]*entities.RankStats, 0)
	err = r.DB.Raw(sqlLose).Scan(&loseRanks).Error
	if err != nil {
		logger.ZError("CreateRankStats", zap.String("sql", sqlLose), zap.Error(err))
		return err
	}
	if len(loseRanks) > 0 {
		// 删除旧数据
		r.DB.Where("code = ?", promoteCode).Delete(&entities.RankStats{})
	}
	// 同批量导入中输的排名
	// 同批量导入中赢的排名
	r.DB.CreateInBatches(loseRanks, len(loseRanks))

	return nil
}

// 获取总统计（带缓存）
func (r *StatsRepository) GetGameStats(uid uint) (*entities.GameStats, error) {
	cacheKey := fmt.Sprintf("game:stats:%d", uid)

	// 尝试从缓存获取
	if cached, err := r.RDS.Get(context.Background(), cacheKey).Bytes(); err == nil {
		var stats entities.GameStats
		if err := json.Unmarshal(cached, &stats); err == nil {
			return &stats, nil
		}
	}

	// 缓存未命中，查询数据库
	var result struct {
		TotalBetCount  int
		TotalBetAmount float64
		TotalProfit    float64
	}

	r.DB.Model(&entities.GameRecord{}).
		Select("COUNT(*) as total_bet_count, "+
			"SUM(bet_amount) as total_bet_amount, "+
			"SUM(profit) as total_profit").
		Where("uid = ?", uid).
		Scan(&result)

	stats := &entities.GameStats{
		TotalBetCount:  result.TotalBetCount,
		TotalBetAmount: result.TotalBetAmount,
		TotalProfit:    result.TotalProfit,
	}

	// 更新缓存（设置1小时过期）
	if data, err := json.Marshal(stats); err == nil {
		r.RDS.Set(context.Background(), cacheKey, data, time.Hour)
	}

	return stats, nil
}

// 获取游戏记录
func (r *StatsRepository) GetGameRecords(param *entities.GetGameRecordListReq) error {

	var tx *gorm.DB = r.DB
	if param.UID != 0 {
		tx = tx.Where("uid = ?", param.UID).Order("bet_time desc")
	}
	param.List = make([]*entities.GameRecord, 0)
	return param.Paginate(tx)
}

// 获取每日统计
func (r *StatsRepository) GetDailyStats(param *entities.GetGamerDailyStatsListReq) error {
	// 1. 构建基础查询
	query := r.DB.Model(&entities.GamerDailyStats{})

	// 2. 应用UID条件
	if param.UID != 0 {
		query = query.Where("uid = ?", param.UID)
	}

	// 3. 处理日期范围
	var startDate, endDate time.Time
	switch param.DateType {
	case "0d":
		endDate = time.Now()
		startDate = endDate
	case "1d":
		endDate = time.Now()
		startDate = endDate.AddDate(0, 0, -1)
	case "7d":
		endDate = time.Now()
		startDate = endDate.AddDate(0, 0, -7)
	case "30d":
		endDate = time.Now()
		startDate = endDate.AddDate(0, 0, -30)
	default:
		// 如果没有指定或自定义日期，默认最近30天
		endDate = time.Now()
		startDate = endDate.AddDate(0, 0, -30)
	}

	// 格式化成数据库兼容的日期格式
	query = query.Where("date BETWEEN ? AND ?",
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"))

	// 4. 处理货币筛选
	if param.Currency != "" {
		query = query.Where("currency = ?", strings.ToUpper(param.Currency))
	}

	// 5. 排序和分页
	query = query.Order("date DESC")

	// 6. 执行分页查询
	param.List = make([]*entities.GamerDailyStats, 0)
	return param.Paginate(query)
}

func (r *StatsRepository) GetCategoryStatsList(param *entities.GetCategoryStatsListReq) error {
	// 初始化返回列表
	param.List = make([]*entities.GameCategoryStats, 0)

	// 构建基础查询
	query := r.DB.Model(&entities.GameRecord{}).
		Select(`category,
            COUNT(*) AS bet_count,
            SUM(profit) AS profit,SUM(bet_amount) AS bet_amount`).
		Where("uid = ?", param.UID).
		Group("category")

	// 添加分类筛选
	if param.Category != "" {
		query = query.Where("category = ?", param.Category)
	}

	// 获取总数（需要去重统计）
	var total int64
	err := r.DB.Model(&entities.GameRecord{}).
		Select("COUNT(DISTINCT category)").
		Where("uid = ?", param.UID).
		Scan(&total).Error
	if err != nil {
		return err
	}
	param.Count = total

	// 执行分页查询
	offset := (param.Page - 1) * param.PageSize
	err = query.
		Order("bet_count DESC"). // 默认按投注次数降序
		Offset(offset).
		Limit(param.PageSize).
		Find(&param.List).Error

	return err
}

func (r *StatsRepository) GetOrCreateSyncStatus(source string) (*entities.GlobalSyncStatus, error) {
	var syncStatus entities.GlobalSyncStatus
	// 获取或创建记录 //时间窗口为3天
	if err := r.DB.Where(entities.GlobalSyncStatus{Source: source}).
		Assign(entities.GlobalSyncStatus{SyncWindow: 3 * 24 * 60}).
		FirstOrCreate(&syncStatus).Error; err != nil {
		return nil, err
	}

	return &syncStatus, nil
}

func (r *StatsRepository) UpdateSyncStatus(syncStatus *entities.GlobalSyncStatus) error {
	return r.DB.Where("source = ?", syncStatus.Source).Updates(syncStatus).Error
}

func (r *StatsRepository) GetUserProfitLeaderboard(uid uint) ([]*entities.GameProfitStats, error) {
	var results []*entities.GameProfitStats

	// 查询每个游戏的盈利情况，并按盈利降序排序
	// err := r.DB.Model(&entities.GameRecord{}).
	// 	Select("uid,game, SUM(profit) as total_profit").
	// 	Where("uid = ?", uid).
	// 	Group("game").
	// 	Order("total_profit DESC").
	// 	Limit(10).
	// 	Find(&results).Error

	err := r.DB.Model(&entities.GameRecord{}).
		Select("uid, game, profit as total_profit").
		Where("uid = ?", uid).
		Order("profit DESC").
		Limit(10).
		Find(&results).Error

	return results, err
}

func (r *StatsRepository) GetProfitLeaderboard() ([]*entities.GameProfitStats, error) {
	var results []*entities.GameProfitStats

	// err := r.DB.Model(&entities.GameRecord{}).
	// 	Select("uid, game, SUM(profit) as total_profit").
	// 	Group("uid, game").
	// 	Order("total_profit DESC").
	// 	Limit(10).
	// 	Find(&results).Error

	err := r.DB.Model(&entities.GameRecord{}).
		Select("uid,game, profit as total_profit").
		Order("profit DESC").
		Limit(10).
		Find(&results).Error
	return results, err
}
