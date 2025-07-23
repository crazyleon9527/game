package repository

import (
	"errors"
	"rk-api/internal/app/entities"
	"rk-api/pkg/logger"

	"github.com/google/wire"
	"gorm.io/gorm"
)

var AdminRepositorySet = wire.NewSet(wire.Struct(new(AdminRepository), "*"))

type AdminRepository struct {
	DB *gorm.DB
}

func (r *AdminRepository) CreateSystemOptionLog(entity *entities.SystemOptionLog) error {
	return r.DB.Create(entity).Error
}

func (r *AdminRepository) GetSysUserByUsername(account string) (*entities.SysUser, error) {
	var entity entities.SysUser
	result := r.DB.Where("user_name = ?", account).Last(&entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &entity, nil
}

func (r *AdminRepository) GetSysUser(entity *entities.SysUser) (*entities.SysUser, error) {
	result := r.DB.First(&entity, entity)
	if result.Error != nil {
		// if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// 	return nil, nil
		// }
		return nil, result.Error
	}
	return entity, nil
}

func (r *AdminRepository) UpdateSysUser(user *entities.SysUser) error {
	return r.DB.Where("user_name = ?", user.Username).Updates(user).Error
}

func (r *AdminRepository) GetSysUserAreaList() ([]*entities.SysUserArea, error) {
	list := make([]*entities.SysUserArea, 0)
	err := r.DB.Table("sys_user_admin").Select("uid", "room").Find(&list).Error
	return list, err
}

func (r *AdminRepository) GetSysUserAdmin(id uint) (*entities.SysUserAdmin, error) {
	var entity entities.SysUserAdmin
	result := r.DB.Where("uid = ?", id).Last(&entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &entity, nil
}

func (r *AdminRepository) CallMonthBackupAndCleanV3(backupMonth string, tableNames string) error {
	// 创建一个结构体来存储可能的输出
	type Result struct {
		Message string
	}
	var results []Result

	// 执行存储过程
	err := r.DB.Raw("CALL MonthBackupAndCleanV3(?, ?)", backupMonth, tableNames).Scan(&results).Error
	if err != nil {
		return err
	}

	// 处理结果（如果有的话）
	for _, result := range results {
		logger.ZInfo(result.Message)
	}

	return nil
}

func (r *AdminRepository) CallMonthBackupAndCleanV4(tableNames string) error {
	// 创建一个结构体来存储可能的输出
	type Result struct {
		Message string
	}
	var results []Result

	// 执行存储过程
	err := r.DB.Raw("CALL MonthBackupAndCleanV4(?)", tableNames).Scan(&results).Error
	if err != nil {
		return err
	}

	// 处理结果（如果有的话）
	for _, result := range results {
		logger.ZInfo(result.Message)
	}

	return nil
}
func (r *AdminRepository) CallChangePC(pcSRC uint, pcDST uint) error {
	// 创建一个结构体来存储可能的输出
	type Result struct {
		Message string
	}
	var results []Result

	// 执行存储过程
	err := r.DB.Raw("CALL ChangePCV3(?,?)", pcSRC, pcDST).Scan(&results).Error
	if err != nil {
		return err
	}
	// 处理结果（如果有的话）
	for _, result := range results {
		logger.ZInfo(result.Message)
	}

	return nil
}
