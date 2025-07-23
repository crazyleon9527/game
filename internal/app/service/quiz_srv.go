package service

import (
	"encoding/json"
	"math"
	"rk-api/internal/app/config"
	"rk-api/internal/app/constant"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/errors"
	"rk-api/internal/app/mq"
	"rk-api/internal/app/mq/handle"
	"rk-api/internal/app/service/repository"
	resty "rk-api/pkg/http"
	"rk-api/pkg/logger"
	"sort"
	"time"

	"github.com/google/wire"
	"github.com/orca-zhang/ecache"
	"github.com/spf13/cast"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	QuizPriorityKey = "quiz_priority"
	QuizListKey     = "quiz_list"
)

var QuizServiceSet = wire.NewSet(
	ProvideQuizService,
)

type QuizService struct {
	Repo              *repository.QuizRepository
	UserSrv           *UserService
	WalletSrv         *WalletService
	quizCache         *ecache.Cache
	eventLimit        uint
	eventStartOffset  uint
	eventEndOffsetMin uint
	eventEndOffsetMax uint
	clobEndpoint      string
	gammaEndpoint     string
}

func ProvideQuizService(repo *repository.QuizRepository, userSrv *UserService, walletSrv *WalletService) *QuizService {
	setting := config.Get().QuizSetting
	logger.ZInfo("ProvideQuizService", zap.Any("setting", setting))
	service := &QuizService{
		Repo:              repo,
		UserSrv:           userSrv,
		WalletSrv:         walletSrv,
		quizCache:         ecache.NewLRUCache(1, 1, 10*time.Minute), //初始化缓存
		eventLimit:        setting.EventLimit,
		eventStartOffset:  setting.EventStartOffset,
		eventEndOffsetMin: setting.EventEndOffsetMin,
		eventEndOffsetMax: setting.EventEndOffsetMax,
		clobEndpoint:      setting.ClobEndpoint,
		gammaEndpoint:     setting.GammaEndpoint,
	}
	return service
}

func (s *QuizService) GetQuizInfo() (*entities.QuizInfoRsp, error) {
	if val, ok := s.quizCache.Get(QuizPriorityKey); ok {
		if v, ok := val.(*entities.QuizInfoRsp); ok {
			if !v.IsClosed {
				logger.ZInfo("GetQuizInfo.quizCache")
				return v, nil
			} else {
				// logger.ZInfo("GetQuizInfo.quizCache closed GetOriginQuizInfo")
				// return s.GetOriginQuizInfo()
			}
		}
	}
	info, err := s.getDBQuizInfo()
	if err != nil || info == nil {
		// logger.Errorf("GetQuizInfo.GetDBQuizInfo Error GetOriginQuizInfo, err - %s", err)
		// return s.GetOriginQuizInfo()
		logger.Errorf("GetQuizInfo.GetDBQuizInfo Error, err - %s, data - %v", err, info)
		return nil, err
	}
	logger.ZInfo("GetQuizInfo.GetDBQuizInfo")
	s.quizCache.Put(QuizPriorityKey, info)
	return info, nil
}

func (s *QuizService) GetQuizList(req *entities.QuizListReq) error {
	err := s.getDBQuizList(req)
	if err != nil {
		logger.Errorf("GetQuizList.GetDBQuizInfo Error, err - %s", err)
		return err
	}
	logger.ZInfo("GetQuizList.GetDBQuizInfo")
	return nil
}

func (s *QuizService) QuizBuy(req *entities.QuizBuyReq) (*entities.QuizBuyRecord, error) {
	// Price
	event, err := s.Repo.GetQuizEventByID(req.EventID)
	if err != nil {
		return nil, err
	}
	market, err := s.Repo.GetQuizMarket(req.EventID, req.MarketID)
	if err != nil {
		return nil, err
	}
	if market == nil {
		return nil, errors.With("market not found")
	}
	var tokenID string
	if req.IsYes == 1 {
		tokenID = market.YesToken
	} else {
		tokenID = market.NoToken
	}
	price, err := s.fetchQuizMarketPrice(tokenID)
	if err != nil {
		return nil, err
	}
	fprice := cast.ToFloat64(price)
	// TODO: 钱包 扣钱 Price
	// Record
	order := &entities.QuizBuyRecord{
		UID:            req.UID,
		EventID:        req.EventID,
		MarketID:       req.MarketID,
		Title:          event.Title,
		Icon:           event.Icon,
		GroupItemTitle: market.GroupItemTitle,
		IsYes:          req.IsYes,
		PayMoney:       req.PayMoney,
		Price:          fprice,
		Rate:           0,
		StartTime:      uint(time.Now().Unix()),
	}
	s.CreateCrashGameOrder(order)

	// market
	if req.IsYes == 1 {
		market.YesPrice = fprice
		market.Ratio = uint(math.Round(fprice * 100))
	} else {
		market.NoPrice = fprice
	}
	err = s.Repo.UpdateQuizMarket(market)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (s *QuizService) CreateCrashGameOrder(order *entities.QuizBuyRecord) error {
	user, err := s.UserSrv.GetUserByUID(order.UID)
	if err != nil {
		return err
	}

	wallet, err := s.WalletSrv.GetWallet(order.UID)
	if err != nil {
		return err
	}

	if wallet.Cash < order.PayMoney {
		return errors.WithCode(errors.InsufficientBalance)
	}

	err = s.WalletSrv.HandleWallet(user.ID, func(wallet *entities.UserWallet, tx *gorm.DB) error {
		wallet.SafeAdjustCash(-order.PayMoney)
		order.CalculateFee() //计算抽水
		order.PromoterCode = user.PromoterCode

		err = s.Repo.CreateQuizBuyRecordWithTx(tx, order)
		if err != nil {
			return err
		}

		if err := s.WalletSrv.UpdateCashWithTx(tx, wallet); err != nil {
			return err
		}

		createFlowQueue, _ := handle.NewCreateFlowQueue(&entities.Flow{
			UID:          order.UID,
			FlowType:     constant.FLOW_TYPE_CRASH,
			Number:       -order.PayMoney,
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

func (s *QuizService) GetQuizBuyRecord(req *entities.QuizBuyRecordReq) error {
	return s.Repo.GetQuizBuyRecord(req)
}

// GetQuizPricesHistory
func (s *QuizService) GetQuizPricesHistory(eventID uint) (*entities.QuizPricesHistoryRsp, error) {
	markets, err := s.Repo.GetQuizMarkets(eventID)
	if err != nil {
		logger.Errorf("GetDBQuizInfo.GetQuizMarkets Error, err - %s, data - %v", err, markets)
		return nil, err
	}
	if len(markets) == 0 {
		return nil, errors.With("markets not found")
	}
	sort.Slice(markets, func(i, j int) bool {
		return markets[i].Ratio > markets[j].Ratio
	})

	maxMarketCount := 4
	rsp := &entities.QuizPricesHistoryRsp{List: make([]*entities.QuizPricesHistoryRspItem, 0, maxMarketCount)}
	for _, market := range markets {
		history, err := s.fetchQuizMarketPricesHistory(market.YesToken)
		if err != nil {
			logger.Errorf("GetDBQuizInfo.GetQuizMarkets Error, err - %s, data - %v", err, markets)
			continue
		}
		rsp.List = append(rsp.List, &entities.QuizPricesHistoryRspItem{
			EventID:        eventID,
			MarketID:       market.MarketID,
			GroupItemTitle: market.GroupItemTitle,
			History:        history,
		})
		maxMarketCount--
		if maxMarketCount == 0 {
			break
		}
	}
	return rsp, nil
}

// GetQuizMarketPricesHistory
func (s *QuizService) GetQuizMarketPricesHistory(eventID, marketID uint) (*entities.QuizPricesHistoryRspItem, error) {
	market, err := s.Repo.GetQuizMarket(eventID, marketID)
	if err != nil {
		return nil, err
	}
	if market == nil {
		return nil, errors.With("market not found")
	}

	history, err := s.fetchQuizMarketPricesHistory(market.YesToken)
	if err != nil {
		return nil, err
	}
	rsp := &entities.QuizPricesHistoryRspItem{
		EventID:        eventID,
		MarketID:       market.MarketID,
		GroupItemTitle: market.GroupItemTitle,
		History:        history,
	}
	return rsp, nil
}

func (s *QuizService) getDBQuizList(req *entities.QuizListReq) error {
	err := s.Repo.GetQuizEventList(req)
	if err != nil {
		logger.Errorf("GetDBQuizInfo.GetQuizEventList Error, err - %s", err)
		return err
	}

	elist, _ := req.List.([]*entities.QuizEvent)
	eventIDs := make([]uint, 0, len(elist))
	for _, v := range elist {
		eventIDs = append(eventIDs, v.EventID)
	}

	ms, err := s.Repo.GetQuizEventsMarkets(eventIDs)
	if err != nil {
		logger.Errorf("GetDBQuizInfo.GetQuizEventsMarkets Error, err - %s", err)
		return err
	}

	mmarket := make(map[uint][]*entities.QuizMarket, len(ms))
	for _, v := range ms {
		mmarket[v.EventID] = append(mmarket[v.EventID], v)
	}

	list := make([]*entities.QuizInfoRsp, 0, len(elist))
	for _, v := range elist {
		if len(mmarket[v.EventID]) > 0 {
			list = append(list, s.buildQuizInfoRsp(v, mmarket[v.EventID]))
		}
	}
	req.List = list
	return nil
}

func (s *QuizService) getDBQuizInfo() (*entities.QuizInfoRsp, error) {
	quizEvent, err := s.Repo.GetQuizEvent()
	if err != nil || quizEvent == nil {
		logger.Errorf("GetDBQuizInfo.GetQuizEvent Error, err - %s, data - %v", err, quizEvent)
		return nil, err
	}
	quizMarkets, err := s.Repo.GetQuizMarkets(quizEvent.EventID)
	if err != nil || len(quizMarkets) == 0 {
		logger.Errorf("GetDBQuizInfo.GetQuizMarkets Error, err - %s, data - %v", err, quizMarkets)
		return nil, err
	}
	return s.buildQuizInfoRsp(quizEvent, quizMarkets), nil
}

func (s *QuizService) buildQuizInfoRsp(quizEvent *entities.QuizEvent, quizMarkets []*entities.QuizMarket) *entities.QuizInfoRsp {
	info := &entities.QuizInfoRsp{
		EventID:  quizEvent.EventID,
		Title:    quizEvent.Title,
		Icon:     quizEvent.Icon,
		Volume:   quizEvent.Volume,
		CloseAt:  quizEvent.CloseAt,
		IsClosed: quizEvent.IsClosed == 1,
		Markets:  make([]*entities.QuizInfoItem, 0, len(quizMarkets)),
	}
	for _, market := range quizMarkets {
		if market.GroupItemTitle == "" && len(quizMarkets) == 1 {
			market.GroupItemTitle = quizEvent.Title
		}
		info.Markets = append(info.Markets, &entities.QuizInfoItem{
			MarketID:       market.MarketID,
			GroupItemTitle: market.GroupItemTitle,
			Ratio:          market.Ratio,
			YesPrice:       market.YesPrice,
			NoPrice:        market.NoPrice,
			IsYesWinner:    market.IsYesWinner == 1,
		})
	}
	sort.Slice(info.Markets, func(i, j int) bool {
		return info.Markets[i].Ratio > info.Markets[j].Ratio
	})
	return info
}

func (s *QuizService) GetOriginQuizInfo() (*entities.QuizInfoRsp, error) {
	quizEventData, err := s.fetchQuizEventData()
	if err != nil || quizEventData == nil {
		return nil, err
	}

	quizEvent := s.buildQuizEvent(quizEventData)
	err = s.Repo.CreateQuizEvent(quizEvent)
	if err != nil {
		return nil, err
	}
	quizMarkets := s.buildQuizMarkets(quizEventData)
	err = s.Repo.CreateQuizMarkets(quizMarkets)
	if err != nil {
		return nil, err
	}

	info := s.buildQuizInfoRsp(quizEvent, quizMarkets)
	s.quizCache.Put(QuizPriorityKey, info)
	return info, nil
}

func (s *QuizService) buildQuizEvent(quizEventData *entities.QuizEventData) *entities.QuizEvent {
	endTime, _ := time.Parse(time.RFC3339, quizEventData.EndDate)
	return &entities.QuizEvent{
		EventID: cast.ToUint(quizEventData.ID),
		Slug:    quizEventData.Slug,
		Title:   quizEventData.Title,
		Icon:    quizEventData.Icon,
		Volume:  quizEventData.Volume,
		CloseAt: uint(endTime.Unix()),
	}
}

func (s *QuizService) buildQuizMarkets(quizEventData *entities.QuizEventData) []*entities.QuizMarket {
	quizMarkets := make([]*entities.QuizMarket, 0, len(quizEventData.Markets))
	for _, market := range quizEventData.Markets {
		var tokenList, priceList []string
		err := json.Unmarshal([]byte(market.ClobTokenIds), &tokenList)
		if err != nil || len(tokenList) < 2 {
			continue
		}
		err = json.Unmarshal([]byte(market.OutcomePrices), &priceList)
		if err != nil || len(priceList) < 2 {
			continue
		}
		yesPrice := cast.ToFloat64(priceList[0])
		ratio := uint(math.Round(yesPrice * 100))
		quizMarkets = append(quizMarkets, &entities.QuizMarket{
			EventID:        cast.ToUint(quizEventData.ID),
			MarketID:       cast.ToUint(market.ID),
			GroupItemTitle: market.GroupItemTitle,
			QuestionID:     market.QuestionID,
			ConditionID:    market.ConditionID,
			Ratio:          ratio,
			YesToken:       tokenList[0],
			YesPrice:       yesPrice,
			NoToken:        tokenList[1],
			NoPrice:        cast.ToFloat64(priceList[1]),
		})
	}
	return quizMarkets
}

func (s *QuizService) fetchQuizMarketPrice(tokenID string) (string, error) {
	client := resty.GetHttpClient()
	resp, err := client.R().SetQueryParams(map[string]string{
		"side":     "buy",
		"token_id": tokenID,
	}).Get(s.clobEndpoint + "/price")
	// 错误处理
	if err != nil {
		logger.Errorf("fetchQuizMarketPrice Error on response.\n[ERROR] - %s", err)
		return "", err
	}

	var responseData entities.QuizPriceData
	err = json.Unmarshal(resp.Body(), &responseData)
	if err != nil {
		logger.Errorf("fetchQuizMarketPrice decode [ERROR] - %v, responseData: %s, body: %s",
			err, responseData, string(resp.Body()))
		return "", err
	}
	return responseData.Price, nil
}

func (s *QuizService) fetchQuizEventData() (*entities.QuizEventData, error) {
	client := resty.GetHttpClient()
	startDateMin := time.Now().Add(time.Duration(s.eventStartOffset) * -24 * time.Hour)
	endDateMin := time.Now().Add(time.Duration(s.eventEndOffsetMin) * 24 * time.Hour)
	endDateMax := time.Now().Add(time.Duration(s.eventEndOffsetMax) * 24 * time.Hour)
	resp, err := client.R().SetQueryParams(map[string]string{
		"active":         "true",
		"closed":         "false",
		"archived":       "false",
		"limit":          cast.ToString(s.eventLimit),
		"start_date_min": startDateMin.Format(time.DateOnly), // 2025-01-01
		"start_date_max": time.Now().Format(time.DateOnly),   // 2025-01-01
		"end_date_min":   endDateMin.Format(time.DateOnly),   // 2025-01-01
		"end_date_max":   endDateMax.Format(time.DateOnly),   // 2025-01-01
	}).Get(s.gammaEndpoint + "/events")
	// 错误处理
	if err != nil {
		logger.Errorf("fetchQuizEventData Error on response.\n[ERROR] - %s", err)
		return nil, err
	}

	var responseData []*entities.QuizEventData
	err = json.Unmarshal(resp.Body(), &responseData)
	if err != nil || len(responseData) == 0 {
		logger.Errorf("fetchQuizEventData decode [ERROR] - %v, responseData: %s, body: %s",
			err, responseData, string(resp.Body()))
		return nil, err
	}
	return responseData[0], nil
}

func (s *QuizService) fetchQuizMarketPricesHistory(tokenID string) ([]*entities.QuizPricesHistoryDataItem, error) {
	client := resty.GetHttpClient()
	resp, err := client.R().SetQueryParams(map[string]string{
		"interval": "1d",
		"market":   tokenID,
	}).Get(s.clobEndpoint + "/prices-history")
	// 错误处理
	if err != nil {
		logger.Errorf("fetchQuizMarketPricesHistory Error on response.\n[ERROR] - %s", err)
		return nil, err
	}

	var responseData entities.QuizPricesHistoryData
	err = json.Unmarshal(resp.Body(), &responseData)
	if err != nil {
		logger.Errorf("fetchQuizMarketPricesHistory decode [ERROR] - %v, responseData: %s, body: %s",
			err, responseData, string(resp.Body()))
		return nil, err
	}
	return responseData.History, nil
}
