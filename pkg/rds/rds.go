package rds

import (
	"context"
	"time"

	"github.com/google/martian/log"
	"github.com/redis/go-redis/v9"
	"rk-api/internal/app/config"
)

func InitRDS(setting config.RDBSettings) (redis.UniversalClient, error) {
	var client redis.UniversalClient
	if setting.UseCluster {
		// 集群模式
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:           setting.ClusterAddrs,
			Password:        setting.Password,
			PoolSize:        setting.PoolSize,
			MinIdleConns:    setting.MinIdleConns,
			ConnMaxLifetime: time.Duration(setting.MaxConnAge) * time.Second,
			PoolTimeout:     time.Duration(setting.PoolTimeout) * time.Second,
			ConnMaxIdleTime: time.Duration(setting.IdleTimeout) * time.Second,
		})
	} else {
		// 单机模式
		client = redis.NewClient(&redis.Options{
			Addr:            setting.ClusterAddrs[0],
			Password:        setting.Password,
			DB:              setting.DB,
			PoolSize:        setting.PoolSize,
			MinIdleConns:    setting.MinIdleConns,
			ConnMaxLifetime: time.Duration(setting.MaxConnAge) * time.Second,
			PoolTimeout:     time.Duration(setting.PoolTimeout) * time.Second,
			ConnMaxIdleTime: time.Duration(setting.IdleTimeout) * time.Second,
		})
	}
	pong, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Errorf("redis connect ping failed, err: %v", err)
		return nil, err
	}
	log.Infof("redis connect ping response: pong %s", pong)
	return client, nil
}
