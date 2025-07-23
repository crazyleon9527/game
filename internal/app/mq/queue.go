package mq

import (
	"context"
	"strings"
	"time"

	"rk-api/internal/app/config"
	"rk-api/internal/app/mq/handle"
	"rk-api/internal/app/service/async"

	"rk-api/pkg/logger"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

var MClient *asynq.Client

func Start(service async.IAsyncService) {
	setting := config.Get().RDBSettings

	var opt asynq.RedisConnOpt
	if setting.UseCluster {
		// 集群模式
		opt = asynq.RedisClusterClientOpt{
			Addrs:    setting.ClusterAddrs,
			Password: setting.Password,
		}
	} else {
		// 单机模式
		opt = asynq.RedisClientOpt{
			Addr:     setting.ClusterAddrs[0],
			Password: setting.Password,
			DB:       setting.DB,
			PoolSize: setting.PoolSize,
		}

	}

	initClient(opt)
	go initListen(opt, service)
}

func initClient(redis asynq.RedisConnOpt) {
	MClient = asynq.NewClient(redis)
}

func initListen(redis asynq.RedisConnOpt, service async.IAsyncService) {
	defer func() {
		if r := recover(); r != nil {
			logger.ZError("Recovered in MQ initListen", zap.Any("Error", r))
		}
	}()

	srv := asynq.NewServer(
		redis,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			// 添加重试和错误处理配置
			RetryDelayFunc: func(n int, err error, task *asynq.Task) time.Duration {
				// 如果是集群重定向错误，立即重试
				if isClusterRedirectError(err) {
					logger.ZWarn("Redis cluster redirect error, retrying immediately",
						zap.Error(err),
						zap.String("type", task.Type()),
					)
					return time.Millisecond * 100
				}

				// 其他错误使用指数退避重试
				delay := time.Duration(n*n) * time.Second
				if delay > 30*time.Second {
					delay = 30 * time.Second
				}

				logger.ZWarn("Task retry",
					zap.Error(err),
					zap.Int("attempt", n),
					zap.Duration("delay", delay),
					zap.String("type", task.Type()),
				)

				return delay
			},
			// 错误处理
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				if isClusterRedirectError(err) {
					logger.ZError("Redis cluster redirect error in task",
						zap.Error(err),
						zap.String("type", task.Type()),
						zap.Any("payload", task.Payload()),
					)
				} else {
					logger.ZError("Task processing error",
						zap.Error(err),
						zap.String("type", task.Type()),
						zap.Any("payload", task.Payload()),
					)
				}
			}),
			Logger: logger.GetLogger().Sugar(),
		},
	)
	mux := asynq.NewServeMux()
	mux.HandleFunc(handle.QueueCreateFlow, handle.NewCreateFlowHandler(service))        //用户流水处理
	mux.HandleFunc(handle.QueueUserExpiration, handle.NewUserExpirationHandle(service)) //用户缓存已过期处理
	mux.HandleFunc(handle.QueueInviteRelation, handle.NewInviteRelationHandle(service)) //邀请关系添加处理
	mux.HandleFunc(handle.QueueInvitePinduo, handle.NewInvitePinduoHandle(service))     //邀请pinduo添加处理
	mux.HandleFunc(handle.QueueOptionLog, handle.NewOptionLogHandle(service))           //后台操作日志
	mux.HandleFunc(handle.QueueNotification, handle.NewNotificationHandler(service))    //用户通知信息

	if err := srv.Run(mux); err != nil {
		logger.ZError("MClient RUN", zap.Error(err))
	}

	logger.Error("MQ START")

}

func Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	if MClient != nil {
		info, err := MClient.Enqueue(task, opts...)
		if err != nil && isClusterRedirectError(err) {
			logger.ZWarn("Redis cluster redirect error during enqueue, retrying",
				zap.Error(err),
				zap.String("type", task.Type()),
			)
			// 重试一次
			time.Sleep(100 * time.Millisecond)
			return MClient.Enqueue(task, opts...)
		}
		return info, err
	}
	return nil, nil
}

// 检查是否是集群重定向错误
func isClusterRedirectError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "MOVED") ||
		strings.Contains(err.Error(), "ASK") ||
		strings.Contains(err.Error(), "CROSSSLOT")
}

// func ProvideFlowService(fs *FlowService) *FlowService {
// 	// 返回传入的FlowService实例
// 	return fs
// }

// func InitializeFlowService() *service.FlowService {
// 	wire.Build(
// 		repository.RepoSet,
// 		service.SrvSet,
// 	)
// 	return new(service.FlowService)
// }

// createFlowQueue, _ := handle.NewCreateFlowQueue(flow)
// if _, err := mq.MClient.Enqueue(createFlowQueue); err != nil {
// 	logger.ZInfo("createFlowQueue fail", zap.Error(err))
// }

// func NewUserExpirationQueue(uid string) (*asynq.Task, error) {
//     return asynq.NewTask(
//         QueueUserExpiration,
//         []byte(uid),
//         asynq.Retries(5),                      // 设置最大重试次数为 5
//         asynq.MinRetryInterval(10*time.Second), // 设置最小重试间隔为 10 秒
//     ), nil
// }

// // 回调队列
// orderCallbackQueue, _ := handle.NewOrderCallbackQueue(order)
// mq.MClient.Enqueue(orderCallbackQueue, asynq.MaxRetry(5))   最多重试

// type MessageQueue struct {
// 	service async.IAsyncService
// 	MClient *asynq.Client
// }

// func (m *MessageQueue) Start() {
// 	setting := config.Get().RDBSettings
// 	redis := asynq.RedisClientOpt{
// 		Addr:     setting.DataSource,
// 		Password: setting.Password,
// 		DB:       setting.DB,
// 	}
// 	initClient(redis)
// 	go m.initListen(redis, m.service)
// }

// func (m *MessageQueue) initListen(redis asynq.RedisClientOpt, service async.IAsyncService) {
// 	srv := asynq.NewServer(
// 		redis,
// 		asynq.Config{
// 			Concurrency: 10,
// 			Queues: map[string]int{
// 				"critical": 6,
// 				"default":  3,
// 				"low":      1,
// 			},

// 			Logger: logger.GetLogger().Sugar(),
// 		},
// 	)
// 	mux := asynq.NewServeMux()
// 	mux.HandleFunc(handle.QueueCreateFlow, handle.NewCreateFlowHandler(service))        //用户流水处理
// 	mux.HandleFunc(handle.QueueUserExpiration, handle.NewUserExpirationHandle(service)) //用户缓存已过期处理
// 	mux.HandleFunc(handle.QueueInviteRelation, handle.NewInviteRelationHandle(service)) //邀请关系添加处理
// 	mux.HandleFunc(handle.QueueInvitePinduo, handle.NewInvitePinduoHandle(service))     //邀请pinduo添加处理
// 	mux.HandleFunc(handle.QueueOptionLog, handle.NewOptionLogHandle(service))           //后台操作日志

// 	time.AfterFunc(10*time.Second, func() { //延迟10秒启动
// 		if err := srv.Run(mux); err != nil {
// 			logger.ZError("MClient RUN", zap.Error(err))
// 		}
// 	})
// }
