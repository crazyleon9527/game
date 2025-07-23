package repository

import (
	"rk-api/internal/app/entities"

	"github.com/google/wire"
	"gorm.io/gorm"
)

var RealRepositorySet = wire.NewSet(wire.Struct(new(RealRepository), "*"))

type RealRepository struct {
	DB *gorm.DB
}

func (r *RealRepository) CreateRealAuth(walletAddress *entities.RealAuth) error {
	return r.DB.Create(walletAddress).Error
}

func (r *RealRepository) GetRealAuthByUID(uid uint) (*entities.RealAuth, error) {
	var auth entities.RealAuth
	if err := r.DB.Where("uid = ?", uid).First(&auth).Error; err != nil {
		return nil, err
	}
	return &auth, nil
}

func (r *RealRepository) UpdateRealAuth(auth *entities.RealAuth) error {
	return r.DB.Save(auth).Error
}
