package service

import (
	"fmt"
	"rk-api/internal/app/constant"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/service/repository"
	"rk-api/internal/app/utils"
	"rk-api/pkg/logger"
	"strings"
	"time"

	"github.com/google/wire"
	"github.com/spf13/cast"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"
)

var StatsServiceSet = wire.NewSet(
	ProvideStatsService,
)

type IFetchRecords interface {
	// 修改接口定义，移除playerID参数
	FetchRecords(startTime, endTime time.Time) ([]*entities.GameRecord, error)
}

type StatsService struct {
	Repo     *repository.StatsRepository
	adapters map[string]IFetchRecords
	AgentSrv *AgentService
	StateSrv *StateService
	gameSrv  *GameService
}

func ProvideStatsService(
	Repo *repository.StatsRepository,
	JhszSrv *JhszService,
	AgentSrv *AgentService,
	stateSrv *StateService,
	gameSrv *GameService,
) *StatsService {
	adapters := make(map[string]IFetchRecords)
	adapters["jhsz"] = JhszSrv
	return &StatsService{
		Repo:     Repo,
		adapters: adapters,
		AgentSrv: AgentSrv,
		StateSrv: stateSrv,
		gameSrv:  gameSrv,
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// 同步所有第三方游戏数据
func (s *StatsService) SyncThirdPartyData() {
	// _, cancel := context.WithTimeout(context.Background(), 110*time.Second)
	// defer cancel()
	if len(s.adapters) == 0 {
		return
	}

	if s.StateSrv.GetBoolState(constant.StateGameBetAreaLimit) { //处理游戏下注限制
		logger.ZError("SyncThirdPartyData", zap.Bool("StateGameBetAreaLimit", s.StateSrv.GetBoolState(constant.StateGameBetAreaLimit)))
		return
	}

	if s.StateSrv.GetBoolState(constant.StateMonthBackupAndClean) { //处在月度备份和清理状态
		logger.ZError("SyncThirdPartyData", zap.Bool("StateMonthBackupAndClean", s.StateSrv.GetBoolState(constant.StateMonthBackupAndClean)))
		return
	}
	if s.StateSrv.GetBoolState(constant.StateChangePC) { //处在更换PC，业务员合并
		logger.ZError("SyncThirdPartyData", zap.Bool("StateChangePC", s.StateSrv.GetBoolState(constant.StateChangePC)))
		return
	}

	for source, adapter := range s.adapters {

		s.syncThirdParty(source, adapter)
	}

}

func (s *StatsService) MakeProfitRank() error {
	return s.Repo.CreateRankStats(100, 0)
}

// 获取类别统计列表
func (s *StatsService) GetCategoryStatsList(req *entities.GetCategoryStatsListReq) error {
	return s.Repo.GetCategoryStatsList(req)
}

func (s *StatsService) GetGameRecordList(req *entities.GetGameRecordListReq) error {
	return s.Repo.GetGameRecords(req)
}

func (s *StatsService) GetGamerDailyStatsList(req *entities.GetGamerDailyStatsListReq) error {
	return s.Repo.GetDailyStats(req)
}

// 获取返利列表
func (s *StatsService) GetGameStats(uid uint) (*entities.GameStats, error) {
	return s.Repo.GetGameStats(uid)
}

func (s *StatsService) GetUserProfitLeaderboard(uid uint) ([]*entities.GameProfitStats, error) {
	return s.Repo.GetUserProfitLeaderboard(uid)
}

func (s *StatsService) GetProfitLeaderboard() ([]*entities.GameProfitStats, error) {
	return s.Repo.GetProfitLeaderboard()
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (s *StatsService) syncThirdParty(source string, adapter IFetchRecords) error {
	logger.ZInfo("sync third party data", zap.String("source", source))
	// 获取全局同步状态
	syncStatus, err := s.Repo.GetOrCreateSyncStatus(source)
	if err != nil {
		return err
	}

	// 计算时间窗口
	endTime := time.Now().UTC()
	startTime := time.Unix(0, (syncStatus.LastRecordTime)*int64(time.Millisecond)).UTC() // 转换 int64 到 time.Time
	if syncStatus.LastRecordTime == 0 {
		// 如果 LastRecordTime 是零值，使用默认的时间窗口
		startTime = endTime.Add(-time.Duration(syncStatus.SyncWindow) * time.Minute)
	} else {
		startTime = startTime.Add(1 * time.Second) // 避免重复处理同一记录
	}

	retryCount := 0

	// 分页获取数据
	for {

		records, err := adapter.FetchRecords(startTime, endTime)
		// 格式化时间为字符串（例如：YYYY-MM-DD HH:MM:SS）
		startFormatted := startTime.Format("2006-01-02 15:04:05")
		endFormatted := endTime.Format("2006-01-02 15:04:05")

		logger.ZInfo("sync third party data len 0 ",
			zap.Int("retry_count", retryCount),
			zap.Int("records_count", len(records)),
			zap.String("start_time", startFormatted),
			zap.String("end_time", endFormatted),
		)

		// 打印日志，包含 start 和 end 时间

		if err != nil {
			logger.ZError("sync third party data", zap.Error(err))
			return err
		}
		if len(records) == 0 {

			break
		}

		// logger.ZInfo("sync third party data", zap.Int("records_count", len(records)))
		// 处理并保存记录
		lastRecordTime, err := s.processRecords(source, records)
		if err != nil {
			logger.ZError("sync third party data", zap.Error(err))
			break
		}

		// 更新游标（使用最后记录时间 +1ms 避免重复）
		startTime = lastRecordTime.Add(1 * time.Second)

		// 更新同步状态
		syncStatus.LastRecordTime = lastRecordTime.UnixMilli() // 将时间转换为 int64 时间戳

		// 避免超过当前时间+
		if startTime.After(endTime) {
			break
		}

		// 增加循环计数器
		retryCount++

		// 检查循环次数是否超过最大值，避免死循环
		if retryCount >= 10 {
			break
		}
	}

	// 更新全局同步状态
	syncStatus.LastSyncTime = endTime.UnixMilli() // 将时间转换为 int64 时间戳
	if err := s.Repo.UpdateSyncStatus(syncStatus); err != nil {
		// do nothing
	}
	return nil
}

func (s *StatsService) processRecords(source string, records []*entities.GameRecord) (time.Time, error) {
	tx := s.Repo.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var (
		gameRecords    []entities.GameRecord
		dailyStatsMap  = make(map[string]*entities.GamerDailyStats) // key: date+player
		lastRecordTime time.Time
	)

	// 转换记录
	for _, record := range records {
		// 更新最后记录时间
		if record.BetTime.After(lastRecordTime) {
			lastRecordTime = record.BetTime
		}
		identification := s.gameSrv.GetGameIdentification(record.Game)
		if identification == nil {
			continue
		}
		record.Category = identification.Category
		record.Game = identification.Name

		record.RecordId = source + "-" + record.RecordId + "-" + cast.ToString(record.UID) // 记录ID加上来源前缀确保唯一性

		// 保存记录
		gameRecords = append(gameRecords, *record)

		// 生成每日统计key
		dateKey := record.BetTime.Format("2006-01-02")
		mapKey := fmt.Sprintf("%s:%d", dateKey, record.UID)

		// 更新每日统计
		if stat, exists := dailyStatsMap[mapKey]; exists {
			stat.BetCount++
			stat.BetAmount += record.BetAmount
			stat.Profit += record.Profit
		} else {
			dailyStatsMap[mapKey] = &entities.GamerDailyStats{
				Date:      record.BetTime.Truncate(24 * time.Hour),
				UID:       record.UID,
				Source:    source,
				BetCount:  1,
				BetAmount: record.BetAmount,
				Profit:    record.Profit,
				Currency:  record.Currency,
			}
		}
	}

	// 批量写入游戏记录
	if len(gameRecords) > 0 {
		s.ProcessRabateRecords(source, records, func(list []*PlayerRecord) {
			go s.ProcessPlayerRecordGroup(list) //处理返利记录
		})

		result := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "record_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"bet_amount", "amount", "profit"}),
		}).CreateInBatches(gameRecords, 200)

		if result.Error != nil {
			tx.Rollback()
			return time.Time{}, result.Error
		}
		logger.ZInfo("third party data CreateInBatches", zap.Int("RowsAffected", int(result.RowsAffected)))
	}

	// 批量更新每日统计
	if len(dailyStatsMap) > 0 {
		values := make([]string, 0, len(dailyStatsMap))
		for _, stat := range dailyStatsMap {
			values = append(values, fmt.Sprintf("('%s', %d, '%s', %d, %f, %f, '%s')",
				stat.Date.Format(constant.PeriodLayout2), stat.UID, stat.Source,
				stat.BetCount, stat.BetAmount, stat.Profit, stat.Currency))
		}
		isql := fmt.Sprintf(`
		INSERT INTO gamer_daily_stats 
			(date, uid, source, bet_count, bet_amount, profit, currency)
		VALUES %s
		ON DUPLICATE KEY UPDATE
			bet_count = bet_count + VALUES(bet_count),
			bet_amount = bet_amount + VALUES(bet_amount),
			profit = profit + VALUES(profit)`, strings.Join(values, ","))
		err := tx.Exec(isql).Error
		if err != nil {
			tx.Rollback()
			return time.Time{}, err
		}
	}

	return lastRecordTime, tx.Commit().Error
}

func (s *StatsService) ProcessRabateRecords(source string, unproceedRecords []*entities.GameRecord, handle func(list []*PlayerRecord)) {

	playerFlowGroup := NewPlayerFlowGroup()
	for _, record := range unproceedRecords {
		// 如果处理成功，则将记录的ID添加到列表中。

		playerFlow := playerFlowGroup.get(record.UID)
		if playerFlow == nil {
			playerFlow = &PlayerRecord{UID: record.UID, PromoterCode: record.PromoterCode}
			playerFlow.add(record.Amount)
			playerFlowGroup.add(playerFlow)
		} else {
			playerFlow.add(record.Amount)
		}

		record.Status = 2 //已经返利处理  important!
	}
	handle(playerFlowGroup.getList())
}

func (s *StatsService) ProcessPlayerRecordGroup(unproceedRecordList []*PlayerRecord) error {

	defer utils.PrintPanicStack()

	BatchSize := 10                            //每次处理10个用户
	rbMap, err := s.AgentSrv.GetRakeBackMap(0) // game = 0
	if err != nil {
		return err
	}
	for start := 0; start < len(unproceedRecordList); start += BatchSize {
		end := start + BatchSize
		if end > len(unproceedRecordList) {
			end = len(unproceedRecordList)
		}

		batchRecords := unproceedRecordList[start:end]

		tx := s.Repo.DB

		for _, record := range batchRecords {

			relations, err := s.AgentSrv.GetInviteRelationList(record.UID)
			if err != nil {

			} else {
				if len(relations) > 0 { //有邀请关系时候
					hprList := record.getHPRList(&relations, rbMap, 0)
					if err := s.AgentSrv.BatchUpdateRelationReturnCash(tx, &relations); err != nil { //更新邀请关系里的return_cash
						return err
					}
					if err := tx.CreateInBatches(hprList, len(hprList)).Error; err != nil { //插入到领取表
						return err
					}
				}
			}
		}

	}
	return nil
}
