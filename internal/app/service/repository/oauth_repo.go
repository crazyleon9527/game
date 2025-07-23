package repository

import (
	"github.com/google/wire"
	"gorm.io/gorm"
)

var OauthRepositorySet = wire.NewSet(wire.Struct(new(OauthRepository), "*"))

type OauthRepository struct {
	DB *gorm.DB
}

// func (r *OauthRepository) CreateOauth(Oauth *entities.Oauth) error {
// 	return r.DB.Create(Oauth).Error
// }

// func (r *OauthRepository) CreateRefundOauthOauth(Oauth *entities.RefundOauthOauth) error {
// 	return r.DB.Create(Oauth).Error
// }

// func (r *OauthRepository) CreateRefundLinkOauthOauth(Oauth *entities.RefundLinkOauthOauth) error {
// 	return r.DB.Create(Oauth).Error
// }

// func (r *OauthRepository) GetUnProceedRefundOauthOauth(OauthType, limit int) ([]*entities.RefundOauthOauth, error) {
// 	list := make([]*entities.RefundOauthOauth, 0)
// 	err := r.DB.Clauses(dbresolver.Write).Where("status = ? and type = ? and number < ?", 0, OauthType, 0).Select("id", "uid", "number", "type", "pc").Limit(limit).Find(&list).Error

// 	// err := r.DB.Clauses(dbresolver.Write).Where(" type = ? and status = ?", OauthType, 0).Select("id", "uid", "number", "type", "pc").Limit(limit).Find(&list).Error
// 	return list, err
// }

// func (r *OauthRepository) ProceedRefundOauthOauthListWithTx(tx *gorm.DB, processedUids []uint) error {
// 	if err := tx.Model(&entities.RefundOauthOauth{}).Where("uid IN (?)", processedUids).Update("status", 1).Error; err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (r *OauthRepository) GetOauthList(param *entities.GetOauthListReq) error {
// 	var tx *gorm.DB = r.DB
// 	tx = tx.Where("uid = ?", param.UID)
// 	param.List = make([]*entities.Oauth, 0)
// 	return param.Paginate(tx)
// }
