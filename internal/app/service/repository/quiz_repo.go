package repository

import (
	"rk-api/internal/app/entities"

	"github.com/google/wire"
	"gorm.io/gorm"
)

var QuizRepositorySet = wire.NewSet(wire.Struct(new(QuizRepository), "*"))

type QuizRepository struct {
	DB *gorm.DB
}

func (r *QuizRepository) GetQuizEvent() (*entities.QuizEvent, error) {
	var quizEvent *entities.QuizEvent
	err := r.DB.Where("is_closed = ? and status = ? and is_fetch = ?", 0, 1, 1).
		Order("priority desc").First(&quizEvent).Error
	return quizEvent, err
}

func (r *QuizRepository) GetQuizEventByID(eventID uint) (*entities.QuizEvent, error) {
	var quizEvent *entities.QuizEvent
	err := r.DB.Where("event_id = ?", eventID).First(&quizEvent).Error
	return quizEvent, err
}

func (r *QuizRepository) GetQuizEventList(param *entities.QuizListReq) error {
	tx := r.DB.Where("is_closed = ? and status = ? and is_fetch = ?", 0, 1, 1).
		Order("priority desc")
	param.List = make([]*entities.QuizEvent, 0)
	return param.Paginate(tx)
}

func (r *QuizRepository) GetQuizEventsMarkets(eventIDs []uint) ([]*entities.QuizMarket, error) {
	var quizMarkets []*entities.QuizMarket
	err := r.DB.Where("event_id in ?", eventIDs).Find(&quizMarkets).Error
	return quizMarkets, err
}

func (r *QuizRepository) GetQuizMarkets(eventID uint) ([]*entities.QuizMarket, error) {
	var quizMarkets []*entities.QuizMarket
	err := r.DB.Where("event_id = ?", eventID).Find(&quizMarkets).Error
	return quizMarkets, err
}

func (r *QuizRepository) GetQuizMarket(eventID, marketID uint) (*entities.QuizMarket, error) {
	var quizMarkets entities.QuizMarket
	err := r.DB.Where("event_id = ? and market_id = ?", eventID, marketID).First(&quizMarkets).Error
	return &quizMarkets, err
}

// UpdateQuizMarket 更新竞猜市场
func (r *QuizRepository) UpdateQuizMarket(quizMarket *entities.QuizMarket) error {
	return r.DB.Save(quizMarket).Error
}

// CreateQuizEvent 创建竞猜信息
func (r *QuizRepository) CreateQuizEvent(quizEvent *entities.QuizEvent) error {
	return r.DB.Create(quizEvent).Error
}

// CreateQuizMarkets 创建竞猜市场
func (r *QuizRepository) CreateQuizMarkets(quizMarkets []*entities.QuizMarket) error {
	return r.DB.Create(quizMarkets).Error
}

// GetQuizBuyRecord 获取竞猜购买记录
func (r *QuizRepository) GetQuizBuyRecord(param *entities.QuizBuyRecordReq) error {
	tx := r.DB.Where("uid = ?", param.UID).Order("updated_at desc")
	param.List = make([]*entities.QuizBuyRecord, 0)
	return param.Paginate(tx)
}

// CreateQuizBuyRecord 创建竞猜购买记录
func (r *QuizRepository) CreateQuizBuyRecord(record *entities.QuizBuyRecord) error {
	return r.DB.Create(record).Error
}

// CreateQuizBuyRecordWithTx 创建竞猜购买记录
func (r *QuizRepository) CreateQuizBuyRecordWithTx(tx *gorm.DB, order *entities.QuizBuyRecord) error {
	return tx.Create(order).Error
}
