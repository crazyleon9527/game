package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"rk-api/internal/app/config"
	"rk-api/internal/app/game/crash"
	"rk-api/internal/app/game/dice"
	hash "rk-api/internal/app/game/hash"
	"rk-api/internal/app/game/limbo"
	"rk-api/internal/app/game/mine"
	rg "rk-api/internal/app/game/rg"
	"rk-api/internal/app/service"
	"rk-api/internal/app/service/async"
	"rk-api/internal/app/service/repository"
	"rk-api/internal/app/telegram"
	"rk-api/internal/app/utils"
	"rk-api/pkg/rds"
	"rk-api/pkg/storage"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/minio/minio-go/v7"
	redis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	tb "gopkg.in/telebot.v3"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"gorm.io/plugin/dbresolver"
)

var ProviderSet = wire.NewSet(
	provideMysql,
	provideRDS,
	provideWingo,
	provideNine,
	provideService,
	provideMinio,

	hash.GameManageSet, // 引入 SDGame 依赖
	crash.CrashGameSet, // 引入 CrashGame 依赖
	mine.MineGameSet,   // 引入 MineGame 依赖
	dice.DiceGameSet,   // 引入 DiceGame 依赖
	limbo.LimboGameSet, // 引入 LimboGame 依赖
	// provideTelegramBot(),
)

func provideService(srv *service.AsyncServiceManager) async.IAsyncService {
	return srv
}

func provideNine(srv *service.NineService) rg.INine {
	nine := rg.NewNine(srv)
	err := nine.Init()
	if err != nil {
		panic(fmt.Errorf("failed to initialize Wingo: %w", err))
	}
	fmt.Println("provideNine")
	return nine
}

func provideWingo(srv *service.WingoService) rg.IWingo {
	// wingo := game.NewWingo(srv)
	wingo := rg.NewMultiWingo(srv)
	err := wingo.Init()
	if err != nil {
		panic(fmt.Errorf("failed to initialize Wingo: %w", err))
	}
	fmt.Println("provideWingo")
	return wingo
}

// providers.go
func provideSqliteDB() *gorm.DB {
	dsn := config.Get().DBSettings.DataSource
	if config.Get().DBSettings.Driver == "sqlite3" {
		_ = os.MkdirAll(filepath.Dir(dsn), 0777)
	}
	gormDB, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	if config.Get().DBSettings.EnableAutoMigrate {
		zap.L().Info(utils.GetSelfFuncName(), zap.String("msg", "AutoMigrate"))
		if err := repository.AutoMigrate(gormDB); err != nil {
			panic(err)
		}
	}

	fmt.Println("provideDB")
	return gormDB
}

// https://gorm.io/zh_CN/docs/dbresolver.html 具体设置
func provideMysql() *gorm.DB {
	var gormZapLogger logger.Interface

	if gin.Mode() == "debug" {
		gormZapLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormZapLogger = logger.Default
	}
	// gormZapLogger := &rkLogger.GormLogger{ZapLogger: rkLogger.GetLogger()}

	dbSettings := config.Get().DBSettings
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dbSettings.DataSource, // DSN data source name
		DefaultStringSize:         256,                   // string 类型字段的默认长度
		DisableDatetimePrecision:  true,                  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,                  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,                  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false,                 // 根据版本自动配置
	}), &gorm.Config{
		Logger: gormZapLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		panic(err)
	}

	// 设置数据库连接池参数
	sqlDB, err := gormDB.DB()
	if err != nil {
		panic("failed to get database instance")
	}
	sqlDB.SetMaxIdleConns(dbSettings.MaxIdleConns)
	sqlDB.SetMaxOpenConns(dbSettings.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(*dbSettings.QueryTimeout) * time.Second)

	sources := make([]gorm.Dialector, 0, len(dbSettings.DataSourceReplicas)) //主库列表
	for _, dsn := range dbSettings.DataSourceReplicas {
		sources = append(sources, mysql.Open(dsn))
	}

	replicas := make([]gorm.Dialector, 0, len(dbSettings.DataSourceSearchReplicas)) //读库列表
	for _, dsn := range dbSettings.DataSourceSearchReplicas {
		replicas = append(replicas, mysql.Open(dsn))
	}

	gormDB.Use(dbresolver.
		Register(dbresolver.Config{
			Sources:  sources,  // 写操作在源库中已经 有了
			Replicas: replicas, // 读取
			// Replicas: []gorm.Dialector{mysql.Open(connWrite), mysql.Open(connWrite)}, // 读操作
			Policy: dbresolver.RandomPolicy{}, // sources/replicas 负载均衡策略
		}).SetConnMaxIdleTime(time.Hour).
		SetConnMaxLifetime(time.Duration(*dbSettings.QueryTimeout) * time.Second).
		SetMaxIdleConns(dbSettings.MaxIdleConns).
		SetMaxOpenConns(dbSettings.MaxOpenConns))

	if config.Get().DBSettings.EnableAutoMigrate {
		zap.L().Info(utils.GetSelfFuncName(), zap.String("msg", "AutoMigrate"))
		if err := repository.AutoMigrate(gormDB); err != nil {
			panic(err)
		}
	}

	if err != nil {
		panic("failed to connect database")
	}

	fmt.Println("provideDB")

	return gormDB
}

func provideRDS() redis.UniversalClient {
	setting := config.Get().RDBSettings
	rdsClient, err := rds.InitRDS(setting)
	if err != nil {
		panic("failed to connect redis")
	}
	return rdsClient
}

func provideMinio() *minio.Client {
	setting := config.Get().StorageSettings

	rds, err := storage.InitMinio(&storage.Options{
		Endpoint:        setting.Endpoint,
		AccessKeyID:     setting.AccessKey,
		SecretAccessKey: setting.SecretKey,
		UseSSL:          setting.UseSSL,
	})

	if err != nil {
		panic("failed to connect minio")
	}
	fmt.Println("provideMinio")
	return rds
}

func provideTelegramBot() *tb.Bot {
	bot, err := telegram.InitBot()
	if err != nil {
		panic("failed to init telegram bot")
	}

	fmt.Println("provideTelegramBot")
	return bot
}
