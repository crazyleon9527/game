package repository

import (
	"github.com/google/wire"
	"gorm.io/gorm"
)

var TransactionRepositorySet = wire.NewSet(wire.Struct(new(TransactionRepository), "*"))

type TransactionRepository struct {
	DB *gorm.DB
}
