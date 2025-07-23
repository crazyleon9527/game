package entities

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ---------------------------------------------------------------基于redis 分布式锁-------------------------------------------------------------------------------------------------------------------
type RedisUserLock struct {
	client redis.UniversalClient
}

func NewRedisUserLock(client redis.UniversalClient) *RedisUserLock {
	return &RedisUserLock{client: client}
}

func (r *RedisUserLock) Lock(userID uint) {
	key := fmt.Sprintf("user_lock:%d", userID)
	r.client.SetNX(context.Background(), key, true, 10*time.Second)
}

func (r *RedisUserLock) Unlock(userID uint) {
	key := fmt.Sprintf("user_lock:%d", userID)
	r.client.Del(context.Background(), key)
}

// -------------------------------------------------------------------本地用户锁，支持清理和超时等待---------------------------------------------------------------------------------------------------------------------
// const (
// 	timeout         = 10 * time.Second
// 	lockExpireAfter = 24 * time.Hour // 一天不活跃的锁将被清除
// )
