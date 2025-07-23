package logger

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gookit/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/gorm/logger"
)

/************************************************************gorm*******************************************************************/
// GormLogger 实现了 gorm.logger.Interface，所以可以作为 GORM 的日志记录器
type GormLogger struct {
	ZapLogger *zap.Logger
}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	// 根据需要返回一个新的 logger 实例，或者调整现有的日志级别
	return l
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	// 将信息日志输出至 zap
	l.ZapLogger.Sugar().Infof(msg, data...)
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	// 将警告日志输出至 zap
	l.ZapLogger.Sugar().Warnf(msg, data...)
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	// 将错误日志输出至 zap
	l.ZapLogger.Sugar().Errorf(msg, data...)
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	// 将慢查询和其他重要信息输出至 zap
	elapsed := time.Since(begin)
	sql, rows := fc()
	if err != nil {
		l.ZapLogger.Sugar().Errorf("[%.2fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)
	} else {
		l.ZapLogger.Sugar().Debugf("[%.2fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)
	}
}

/*******************************************************************************************************************************/

func Debug(args ...interface{}) {
	zapLogger.Sugar().Debug(args...)
}

func Error(args ...interface{}) {
	zapLogger.Sugar().Error(args...)
}

func Info(args ...interface{}) {
	zapLogger.Sugar().Info(args...)
}

func Warn(args ...interface{}) {
	zapLogger.Sugar().Debug(args...)
}

func Debugf(msg string, args ...interface{}) {
	zapLogger.Sugar().Debugf(msg, args...)
}

func Errorf(msg string, args ...interface{}) {
	zapLogger.Sugar().Errorf(msg, args...)
}

func Infof(msg string, args ...interface{}) {
	zapLogger.Sugar().Infof(msg, args...)
}
func Warnf(msg string, args ...interface{}) {
	zapLogger.Sugar().Warnf(msg, args...)
}

func Printf(msg string, args ...interface{}) {
	zapLogger.Sugar().Infof(msg, args...)
}

////////////////////////////////////////////////////////////////////////////////

func ZDebug(msg string, fields ...zap.Field) {
	zapLogger.Debug(msg, fields...)
}

// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func ZInfo(msg string, fields ...zap.Field) {
	zapLogger.Info(msg, fields...)
}

func ZWarn(msg string, fields ...zap.Field) {
	zapLogger.Warn(msg, fields...)
}

func ZError(msg string, fields ...zap.Field) {
	zapLogger.Error(msg, fields...)
}

// DPanic logs a message at DPanicLevel. The message includes any fields
// passed at the log site, as well as any fields accumulated on the logger.
//
// If the logger is in development mode, it then panics (DPanic means
// "development panic"). This is useful for catching errors that are
// recoverable, but shouldn't ever happen.
func ZDPanic(msg string, fields ...zap.Field) {
	zapLogger.Error(msg, fields...)
}

// Panic logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then panics, even if logging at PanicLevel is disabled.
func ZPanic(msg string, fields ...zap.Field) {
	zapLogger.Panic(msg, fields...)
}

// Fatal logs a message at FatalLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then calls os.Exit(1), even if logging at FatalLevel is
// disabled.
func ZFatal(msg string, fields ...zap.Field) {
	zapLogger.Fatal(msg, fields...)
}

var (
	once      sync.Once
	zapLogger *zap.Logger
)

func initLogger() {
	// ... 创建 logger 的代码 ...
	var err error
	zapLogger, err = createLogger()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(zapLogger)
}

func GetLogger() *zap.Logger {
	once.Do(initLogger)
	return zapLogger
}

// 自定义的时间格式器
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05")) // 使用Go的时间格式化布局
}

func createLogger() (*zap.Logger, error) {
	logsDir := "logs"
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		os.Mkdir(logsDir, 0755)
	}

	// 获取当前日期并格式化为字符串
	// currentDate := time.Now().Format("2006-01-02")
	// logPath := filepath.Join(logsDir, fmt.Sprintf("app-%s.log", "currentDate"))

	logPath := filepath.Join(logsDir, "app.log")

	encoderConfig := zap.NewProductionEncoderConfig()
	// encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// 设置自定义的时间格式器
	encoderConfig.EncodeTime = customTimeEncoder

	encoder := zapcore.NewJSONEncoder(encoderConfig)

	syncWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:  logPath,
		MaxSize:   10,   // 10M
		MaxAge:    30,   // days
		LocalTime: true, // 使用本地时间为备份文件命名，而非UTC时间
		// MaxBackups: 3, // 备份旧文件的最大个数
		// Compress: true, // 是否压缩/归档旧文件
	})

	// 创建用于文件的Core
	fileCore := zapcore.NewCore(encoder, syncWriter, zapcore.InfoLevel)

	// 使用NewConsoleEncoder创建一个用于控制台输出的Core
	encoderConfig.EncodeLevel = colorLevelEncoder
	consoleCore := zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), zapcore.AddSync(os.Stdout), zapcore.InfoLevel)

	// 使用NewTee将两个Core组合在一起
	core := zapcore.NewTee(fileCore, consoleCore)

	logger := zap.New(core)
	// Add the hook to the logger
	// logger = addDatabaseHookWithInit(logger) //添加入库钩子

	// Replace global logger with our logger
	zap.ReplaceGlobals(logger)

	return logger, nil
}

/****************************************颜色控制**************************************************************************/

func colorLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var clr color.Style
	switch level {
	case zapcore.DebugLevel:
		clr = color.New(color.FgCyan)
	case zapcore.InfoLevel:
		clr = color.New(color.FgGreen)
	case zapcore.WarnLevel:
		clr = color.New(color.FgYellow)
	case zapcore.ErrorLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		clr = color.New(color.FgRed)
	}

	if clr != nil {
		enc.AppendString(clr.Sprint(level.CapitalString()))
	} else {
		enc.AppendString(level.CapitalString())
	}
}

/****************************************入库钩子**************************************************************************/

// 需要初始化
func addDatabaseHookWithInit(logger *zap.Logger) {
	logEntryChannel = make(chan zapcore.Entry, 100) // Assumes a buffer size of 100
	wg.Add(1)
	go processLogEntries()

	addDatabaseHook(logger) //添加钩子
}

func addDatabaseHook(logger *zap.Logger) *zap.Logger {
	newCore := zapcore.RegisterHooks(logger.Core(), func(entry zapcore.Entry) error {
		return writeLogToDatabase(entry)
	})

	logger = logger.WithOptions(zap.WrapCore(func(zapcore.Core) zapcore.Core {
		return newCore
	}))

	return logger
}

func writeLogToDatabase(entry zapcore.Entry) error {
	// 只有错误级别以上的日志会被记录
	if entry.Level < zapcore.ErrorLevel {
		return nil
	}
	logEntryChannel <- entry
	return nil
}

var (
	logEntryChannel chan zapcore.Entry
	wg              sync.WaitGroup
)

func processLogEntries() {
	defer wg.Done()

	for entry := range logEntryChannel {
		if entry.Level < zapcore.ErrorLevel {
			continue
		}
		// _, err := db.Exec("INSERT INTO logs (level, msg, time) VALUES ($1, $2, $3)", entry.Level.String(), entry.Message, entry.Time)
	}
}

/*****************************************************颜色 控制台***************************************************************/

/********************************************************************************************************************/
