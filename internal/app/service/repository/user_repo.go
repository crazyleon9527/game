package repository

import (
	"context"
	"errors"
	"fmt"
	"rk-api/internal/app/constant"
	"rk-api/internal/app/entities"
	"strconv"

	"rk-api/pkg/rds"
	"time"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

var UserRepositorySet = wire.NewSet(wire.Struct(new(UserRepository), "*"))

type UserRepository struct {
	DB  *gorm.DB
	RDS redis.UniversalClient
}

const (
	RDS_ACCESS_TOKEN  = "ACCESS_TOKEN:%v"
	RDS_REFRESH_TOKEN = "REFRESH_TOKEN:%v"
)

func (r *UserRepository) GetUserAccessToken(uid uint) (string, error) {
	var key = fmt.Sprintf(constant.REDIS_KEY_USER_ACCESS_TOKEN, uid)
	val, err := r.RDS.Get(context.Background(), key).Result()
	return val, err
}

// CreateAccessToken 写入access token到Redis中
func (r *UserRepository) SetUserAccessToken(uid uint, accessToken string, expiration time.Duration) (err error) {
	var key = fmt.Sprintf(constant.REDIS_KEY_USER_ACCESS_TOKEN, uid)
	return r.RDS.Set(context.Background(), key, accessToken, expiration).Err()
}

func (r *UserRepository) ExpireUserAccessToken(uid uint) (err error) {
	var key = fmt.Sprintf(constant.REDIS_KEY_USER_ACCESS_TOKEN, uid)
	return r.RDS.Del(context.Background(), key).Err()
}

func (r *UserRepository) GetUser(entity *entities.User) (*entities.User, error) {
	result := r.DB.Clauses(dbresolver.Write).First(&entity, entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return entity, nil
}

func (r *UserRepository) CreateUserWithTx(tx *gorm.DB, user *entities.User) error {
	return tx.Create(user).Error
}

func (r *UserRepository) UpdateUserWithTx(tx *gorm.DB, user *entities.User) error {
	return tx.Updates(user).Error
	// 此处添加Debug方法来输出日志
	// return tx.Debug().Updates(user).Error
}

func (r *UserRepository) UpdateUserWithSelects(selects []string, user *entities.User) error {
	return r.DB.Model(user).Select(selects).Updates(user).Error
}

func (r *UserRepository) GetUserByUsername(username string) (*entities.User, error) {
	var user entities.User
	result := r.DB.Where("username = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepository) GetUserByMobile(mobile string) (*entities.User, error) {
	var user entities.User
	result := r.DB.Where("mobile = ?", mobile).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepository) GetUserByInviteCode(inviteCode string) (*entities.User, error) {
	var user entities.User
	result := r.DB.Where("ic = ?", inviteCode).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepository) GetUserIDsByPC(pcValue uint) ([]uint, error) {
	var userIDs []uint

	// 直接查询 pc 字段等于 pcValue 的用户 ID 列表
	if err := r.DB.Model(&entities.User{}).Where("pc = ?", pcValue).Pluck("id", &userIDs).Error; err != nil {
		return nil, err
	}

	return userIDs, nil
}

func (r *UserRepository) UpdatePassword(user *entities.User, newPassword string) error {
	user.Password = newPassword
	return r.DB.Updates(user).Error
}

func (r *UserRepository) CreatorOrGetVerifyCodeByEmailOrMobile(verifyCode *entities.VerifyCode) error {
	result := r.DB.Where("target = ? and type = ?", verifyCode.Target, verifyCode.BusinessType).FirstOrCreate(&verifyCode)
	return result.Error
}

func (r *UserRepository) UpdateVerifyCode(verifyCode *entities.VerifyCode) error {
	return r.DB.Updates(verifyCode).Error
}

func (r *UserRepository) GetVerifyCodeByEmailOrMobile(mobile string) (*entities.VerifyCode, error) {
	var entity entities.VerifyCode
	result := r.DB.Where("target = ?", mobile).Last(&entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &entity, nil
}

func (r *UserRepository) GetVerifyCode(entity *entities.VerifyCode) (*entities.VerifyCode, error) {
	result := r.DB.Last(&entity, entity)
	if result.Error != nil {

		return nil, result.Error
	}
	return entity, nil
}

func (r *UserRepository) UpdateUserTTL(ID uint, expireTime time.Duration) error {
	if expireTime == 0 {
		return r.RDS.Del(context.Background(), fmt.Sprintf("user:base:%d", ID)).Err()
	}

	if err := r.RDS.Expire(context.Background(), fmt.Sprintf("user:base:%d", ID), expireTime).Err(); err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) ClearUserCache(ID uint) error {
	return r.RDS.Del(context.Background(), fmt.Sprintf("user:base:%d", ID)).Err()
}

// 批量删除Redis中的用户缓存
func (r *UserRepository) BatchClearUserCache(ids []uint) error {
	// 创建一个切片来存储要删除的 Redis 键
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = fmt.Sprintf("user:base:%d", id)
	}
	// 使用 Pipeline 批量执行 DEL 命令
	pipe := r.RDS.Pipeline()
	for _, key := range keys {
		pipe.Del(context.Background(), key)
	}

	// 执行所有命令
	_, err := pipe.Exec(context.Background())
	return err
}

func (r *UserRepository) GetUserByID(uid uint) (*entities.User, error) {
	key := fmt.Sprintf("user:base:%d", uid)
	var user entities.User
	err := r.RDS.HGetAll(context.Background(), key).Scan(&user)
	if err != nil {
		return nil, err
	}
	if err := r.DB.First(&user, uid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if keyValues, err := rds.StructToRedisHashOptimized(user); err != nil {
		return nil, err
	} else {
		r.RDS.HSet(context.Background(), key, keyValues)
	}
	return &user, nil
}

func (r *UserRepository) GetAccessTime(uid uint) (int64, error) {
	key := fmt.Sprintf("user:base:%d", uid)
	return r.RDS.HGet(context.Background(), key, "ac").Int64()
}

func (r *UserRepository) SetAccessTime(uid uint, ac int64) error {
	key := fmt.Sprintf("user:base:%d", uid)
	return r.RDS.HSet(context.Background(), key, "ac", ac).Err()
}

func (r *UserRepository) GetUserPCByID(uid uint) (int, error) {
	key := fmt.Sprintf("user:base:%d", uid)

	// 先尝试从 Redis 获取 PC
	pc, err := r.RDS.HGet(context.Background(), key, "pc").Int()
	if err == nil {
		return pc, nil
	}

	// 如果 Redis 没有命中，则从数据库查询
	var user entities.User
	if err := r.DB.Select("promoter_code").First(&user, uid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}

	// 更新 Redis 缓存
	r.RDS.HSet(context.Background(), key, map[string]interface{}{
		"pc": user.PromoterCode,
	})

	return user.PromoterCode, nil
}

func (r *UserRepository) CreateUser(user *entities.User) error {
	if err := r.DB.Create(user).Error; err != nil {
		return err
	}
	key := fmt.Sprintf("user:base:%d", user.ID)
	r.RDS.Del(context.Background(), key) // 删除旧的缓存
	return nil
}

func (r *UserRepository) UpdateUser(user *entities.User) error {

	key := fmt.Sprintf("user:base:%d", user.ID)

	if err := r.DB.Model(&entities.User{}).Where("id = ?", user.ID).Updates(user).Error; err != nil {
		return err
	}
	return r.RDS.Del(context.Background(), key).Err() // 删除旧的缓存
}

func (r *UserRepository) UpdateUserNameAndGender(uid uint, nickname string, gender uint8) error {

	key := fmt.Sprintf("user:base:%d", uid)

	if err := r.DB.Model(&entities.User{}).Where("id = ?", uid).Updates(map[string]interface{}{
		"nickname": nickname,
		"gender":   gender,
	}).Error; err != nil {
		return err
	}
	return r.RDS.Del(context.Background(), key).Err() // 删除旧的缓存
}

// 在Redis中，使用HINCRBY命令对一个不存在的key执行递增操作时，Redis会自动创建这个key并将Hash中的字段初始化为0
func (r *UserRepository) IncrementBetTimes(uid string) error {
	date := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("user:%s:betTimes:%s", uid, date)

	// 递增用户的下注次数
	_, err := r.RDS.HIncrBy(context.Background(), key, "betTimes", 1).Result()
	if err != nil {
		return err
	}
	_, err = r.RDS.Expire(context.Background(), key, 24*time.Hour).Result()
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) GetBetTimes(uid string) (int64, error) {
	date := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("user:%s:betTimes:%s", uid, date)

	result, err := r.RDS.HGet(context.Background(), key, "betTimes").Result()
	if err == redis.Nil {
		return 0, nil
	} else if err != nil {
		// 其他错误
		return 0, err
	}
	betTimes, err := strconv.ParseInt(result, 10, 64)
	if err != nil {
		return 0, err
	}

	return betTimes, nil
}
