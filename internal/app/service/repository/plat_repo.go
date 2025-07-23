package repository

import (
	"errors"
	"rk-api/internal/app/entities"

	"github.com/google/wire"
	"gorm.io/gorm"
)

var PlatRepositorySet = wire.NewSet(wire.Struct(new(PlatRepository), "*"))

type PlatRepository struct {
	DB *gorm.DB
}

func (r *PlatRepository) GetPlatSetting() (*entities.PlatSetting, error) {
	var entity entities.PlatSetting
	result := r.DB.First(&entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &entity, nil
}
