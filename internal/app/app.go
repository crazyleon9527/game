package internal

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"rk-api/internal/app/config"
	"rk-api/internal/app/mq"
	"rk-api/internal/app/task"
	"rk-api/internal/app/utils"
	"rk-api/pkg/logger"
	"syscall"
	"time"

	"go.uber.org/zap"

	_ "rk-api/internal/app/swagger" // 引入swagger
)

type options struct {
	ConfigFile string
	Version    string
}

type Option func(*options)

func SetConfigFile(s string) Option {
	return func(o *options) {
		o.ConfigFile = s
	}
}

func SetVersion(s string) Option {
	return func(o *options) {
		o.Version = s
	}
}

func Init(ctx context.Context, opts ...Option) (func(), error) {

	defer utils.PrintPanicStack()

	var o options
	for _, opt := range opts {
		opt(&o)
	}

	rand.Seed(time.Now().UnixNano()) // 使用种子初始化随机数生成器，通常使用当前时间作为种子
	config.MustLoad(o.ConfigFile)

	// 设置默认时区
	var err error
	// utils.GetValidate() // 初始化 validate
	// logger.WithContext(ctx).Printf("Start server,#run_mode %s,#version %s,#pid %d", config.C.RunMode, o.Version, os.Getpid())
	loggerCleanFunc, err := InitLogger()
	if err != nil {
		return nil, err
	}
	logger.ZInfo("InitLogger", zap.Any("config", config.Get()), zap.String("file", o.ConfigFile))

	if config.Get().ServiceSettings.Timezone != "" {
		time.Local, err = time.LoadLocation(config.Get().ServiceSettings.Timezone) // 更换为你自己的时区
		if err != nil {
			log.Fatal("Invalid time zone:", err)
		}
		// 输出当前的时间
		fmt.Println("Current time:", time.Now().Local())
	}

	monitorCleanFunc := InitMonitor(ctx)
	injector, injectorCleanFunc, err := BuildInjector()
	if err != nil {
		return nil, err
	}
	httpServerCleanFunc := InitHTTPServer(ctx, injector.Engine)

	if config.Get().ServiceSettings.EnableWingo {
		go injector.Wingo.Start()
	}
	if config.Get().ServiceSettings.EnableNine {
		go injector.Nine.Start()
	}
	if config.Get().ServiceSettings.EnableMQ {
		// 队列启动
		mq.Start(injector.Service)
	}

	if config.Get().ServiceSettings.EnableTask {
		err = task.Start(injector.Service) //任务只能在一个地方启动
		if err != nil {
			log.Printf("Error starting the task scheduler: %v", err)
			return nil, err
		}
	}

	// 定时任务

	return func() {
		httpServerCleanFunc()
		injectorCleanFunc()
		monitorCleanFunc()
		loggerCleanFunc()
	}, nil
}

func InitMonitor(ctx context.Context) func() {

	return func() {}
}

func InitLogger() (func(), error) {
	logger := logger.GetLogger()
	defer logger.Sync()

	return func() {}, nil
}

func InitHTTPServer(ctx context.Context, handler http.Handler) func() {
	addr := config.Get().ServiceSettings.ListenAddress
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		// logger.WithContext(ctx).Printf("HTTP server is running at %s.", addr)

		var err error

		err = srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}

	}()

	return func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(10))
		defer cancel()

		srv.SetKeepAlivesEnabled(false)
		if err := srv.Shutdown(ctx); err != nil {
			// logger.WithContext(ctx).Errorf(err.Error())
		}
	}
}

func Run(ctx context.Context, opts ...Option) error {
	state := 1
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	cleanFunc, err := Init(ctx, opts...)
	if err != nil {
		return err
	}

EXIT:
	for {
		sig := <-sc
		// logger.WithContext(ctx).Infof("Receive signal[%s]", sig.String())
		switch sig {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			state = 0
			break EXIT
		case syscall.SIGHUP:
		default:
			break EXIT
		}
	}

	cleanFunc()
	// logger.WithContext(ctx).Infof("Server exit")
	time.Sleep(time.Second)
	os.Exit(state)
	return nil
}
