package logger

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.elastic.co/ecszap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
	"maphraohom.app/maphraohom-backoffice/config"
	"maphraohom.app/maphraohom-backoffice/pkg/database"
	"maphraohom.app/maphraohom-backoffice/pkg/storage"
)

var AppLogger *Logger

func CurrentLogger() *Logger {
	return AppLogger
}

type ZapLoggerFactory interface {
	NewZapLogger() (*zap.Logger, error)
	NewZapLoggerWithElasticsearch() (*zap.Logger, error)
	NewZapLoggerWithDatabase(dbClient *gorm.DB) (*zap.Logger, error)
	NewZapLoggerWithFile() (*zap.Logger, *os.File, string, error)
	NewZapLoggerWithS3() (*zap.Logger, *os.File, string, error)
	NewZapLoggerWithGraylog() (*zap.Logger, error)
	NewZapLoggerWithGoogleCloudLogging() (*zap.Logger, error)
}

type zapLoggerFactory struct {
	config  *config.Config
	logPath string
}

type Log struct {
	ID        uint                   `json:"id" gorm:"primaryKey"`
	Timestamp time.Time              `json:"timestamp" gorm:"not null"`
	Level     string                 `json:"level" gorm:"size:10;not null"`
	Message   string                 `json:"message" gorm:"not null"`
	Fields    map[string]interface{} `json:"fields" gorm:"type:jsonb"`
}

type Logger struct {
	zap         *zap.Logger
	logFile     *os.File
	LogFilePath string
}

func NewLogger(zapLogger *zap.Logger, logFile *os.File, logFilePath string) *Logger {
	return &Logger{
		zap:         zapLogger,
		logFile:     logFile,
		LogFilePath: logFilePath,
	}
}

func NewZapLoggerFactory() ZapLoggerFactory {
	return &zapLoggerFactory{
		config:  config.Global,
		logPath: config.Global.Logger.LogPath,
	}
}

func Initialize() *Logger {
	var (
		logFactory  = NewZapLoggerFactory()
		zapLogger   *zap.Logger
		logFile     *os.File
		logFilePath string
		err         error
	)

	if strings.ToLower(config.Global.Logger.LogMode) == "elastic" {
		zapLogger, err = logFactory.NewZapLoggerWithElasticsearch()
		if err != nil {
			log.Printf("[App] Cannot initial logger (elastic)")
			log.Fatal("[App] LoggerError:", err)
		}
	} else if strings.ToLower(config.Global.Logger.LogMode) == "database" {
		zapLogger, err = logFactory.NewZapLoggerWithDatabase(database.CurrentDatabase())
		if err != nil {
			log.Printf("[App] Cannot initial logger (database)")
			log.Fatal("[App] LoggerError:", err)
		}
	} else if strings.ToLower(config.Global.Logger.LogMode) == "graylog" {
		zapLogger, err = logFactory.NewZapLoggerWithGraylog()
		if err != nil {
			log.Printf("[App] Cannot initial logger (graylog)")
			log.Fatal("[App] LoggerError:", err)
		}
	} else if strings.ToLower(config.Global.Logger.LogMode) == "s3" {
		zapLogger, logFile, logFilePath, err = logFactory.NewZapLoggerWithS3()
		if err != nil {
			log.Printf("[App] Cannot initial logger (s3)")
			log.Fatal("[App] LoggerError:", err)
		}
	} else if strings.ToLower(config.Global.Logger.LogMode) == "gcp" {
		zapLogger, err = logFactory.NewZapLoggerWithGoogleCloudLogging()
		if err != nil {
			log.Printf("[App] Cannot initial logger (gcp)")
			log.Fatal("[App] LoggerError:", err)
		}
	} else if strings.ToLower(config.Global.Logger.LogMode) == "file" {
		zapLogger, logFile, logFilePath, err = logFactory.NewZapLoggerWithFile()
		if err != nil {
			log.Printf("[App] Cannot initial logger (file)")
			log.Fatal("[App] LoggerError:", err)
		}
	} else {
		zapLogger, err = logFactory.NewZapLogger()
		if err != nil {
			log.Printf("[App] Cannot initial logger")
			log.Fatal("[App] LoggerError:", err)
		}
	}

	return NewLogger(zapLogger, logFile, logFilePath)
}

func (lf *zapLoggerFactory) NewZapLogger() (*zap.Logger, error) {
	log, err := zap.NewProduction(zap.AddCaller(), zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	return log, nil
}

func (lf *zapLoggerFactory) NewZapLoggerWithElasticsearch() (*zap.Logger, error) {
	encoderConfig := ecszap.NewDefaultEncoderConfig()
	core := ecszap.NewCore(encoderConfig, os.Stdout, zap.InfoLevel)
	log := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return log, nil
}

func (lf *zapLoggerFactory) NewZapLoggerWithDatabase(dbClient *gorm.DB) (*zap.Logger, error) {
	core := NewZapDBCore(dbClient, zapcore.DebugLevel)
	log := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return log, nil
}

func (lf *zapLoggerFactory) NewZapLoggerWithGraylog() (*zap.Logger, error) {
	var core zapcore.Core

	// Initialize a new graylog client with TLS if enabled
	if lf.config.Logger.GraylogConfig.TlsEnable {
		core = NewZapGraylogCoreWithLTS(lf.config, zapcore.DebugLevel)
	} else {
		core = NewZapGraylogCore(lf.config, zapcore.DebugLevel)
	}

	log := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return log, nil
}

func (lf *zapLoggerFactory) NewZapLoggerWithFile() (*zap.Logger, *os.File, string, error) {
	var (
		now         = time.Now()
		logFilePath = path.Join(lf.logPath, fmt.Sprintf("%s_%s.log", lf.config.App.ServiceName, now.Format("2006-01-02")))
	)

	// Create directory if not exist
	if _, err := os.Stat(lf.logPath); os.IsNotExist(err) {
		err := os.Mkdir(lf.logPath, 0700)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Create or open the log file
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, "", err
	}

	productionEncoder := zap.NewProductionEncoderConfig()
	productionEncoder.EncodeTime = zapcore.ISO8601TimeEncoder

	fileEncoder := zapcore.NewJSONEncoder(productionEncoder)
	consoleEncoder := zapcore.NewConsoleEncoder(productionEncoder)

	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.InfoLevel
	})

	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, zapcore.AddSync(file), highPriority),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), highPriority),
	)

	log := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return log, file, logFilePath, nil
}

func (lf *zapLoggerFactory) NewZapLoggerWithS3() (*zap.Logger, *os.File, string, error) {
	var (
		now         = time.Now()
		logFilePath = path.Join(lf.logPath, fmt.Sprintf("%s_%s.log", lf.config.App.ServiceName, now.Format("2006-01-02 15:04:05")))
	)

	// Create directory if not exist
	if _, err := os.Stat(lf.logPath); os.IsNotExist(err) {
		err := os.Mkdir(lf.logPath, 0700)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Create or open the log file
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, "", err
	}

	productionEncoder := zap.NewProductionEncoderConfig()
	productionEncoder.EncodeTime = zapcore.ISO8601TimeEncoder

	fileEncoder := zapcore.NewJSONEncoder(productionEncoder)
	consoleEncoder := zapcore.NewConsoleEncoder(productionEncoder)

	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.InfoLevel
	})

	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, zapcore.AddSync(file), highPriority),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), highPriority),
	)

	log := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return log, file, logFilePath, nil
}

func (lf *zapLoggerFactory) NewZapLoggerWithGoogleCloudLogging() (*zap.Logger, error) {
	// Initialize a new graylog client with TLS if enabled
	core := NewZapGoogleCloudLoggingCore(lf.config, zapcore.DebugLevel)
	log := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return log, nil
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	checkLogFileSizeLimit()
	currentLogger := CurrentLogger()
	currentLogger.zap.Info(msg, fields...)
	go sendLogToS3()
}

func (l *Logger) Warn(msg string, fields ...zap.Field) {
	checkLogFileSizeLimit()
	currentLogger := CurrentLogger()
	currentLogger.zap.Warn(msg, fields...)
	go sendLogToS3()
}

func (l *Logger) Debug(msg string, fields ...zap.Field) {
	checkLogFileSizeLimit()
	currentLogger := CurrentLogger()
	currentLogger.zap.Debug(msg, fields...)
	go sendLogToS3()
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	checkLogFileSizeLimit()
	currentLogger := CurrentLogger()
	currentLogger.zap.Error(msg, fields...)
	go sendLogToS3()
}

func (l *Logger) Panic(msg string, fields ...zap.Field) {
	checkLogFileSizeLimit()
	currentLogger := CurrentLogger()
	currentLogger.zap.Panic(msg, fields...)
	go sendLogToS3()
}

func (l *Logger) Sync() {
	// Sync zap logger
	l.zap.Sync()
}

func (l *Logger) GetLogFileKbSize() float64 {
	fi, err := l.logFile.Stat()
	if err != nil {
		log.Fatal(err)
	}

	return float64(fi.Size()) / 1024
}

func Close() {
	if !fiber.IsChild() {
		if config.Global.Logger.LogMode == "s3" {
			checkLogFileSizeLimit()
			sendLogToS3()
		} else if config.Global.Logger.LogMode == "file" {
			CurrentLogger().logFile.Close()
		}
	}
}

func checkLogFileSizeLimit() {
	var l = CurrentLogger()

	// Check if logger is not 's3' mode of 'file' mode
	if config.Global.Logger.LogMode != "s3" && config.Global.Logger.LogMode != "file" {
		return
	}

	// Get log file size
	fileSize := l.GetLogFileKbSize()

	// Check if log file size exceeds the limit (90%)
	if fileSize > (float64(config.Global.Logger.S3Config.MaxFileKbSize) * 0.90) {
		// Print file size
		log.Printf(
			"[App] Log file size reached 90%% of the limit: %0.2f KB (Max: %0.2f KB)",
			fileSize,
			float64(config.Global.Logger.S3Config.MaxFileKbSize),
		)

		// Close logger file
		l.logFile.Close()

		// Delete old log file
		err := os.Remove(l.LogFilePath)
		if err != nil {
			log.Printf("[App] Error deleting log file: %v", err)
		}

		// Reinitialize logger
		AppLogger = Initialize()
	}
}

func sendLogToS3() {
	var (
		ctx              = context.Background()
		l                = CurrentLogger()
		localLogFile     *os.File
		localLogFileName string
		s3LogPath        = fmt.Sprintf("%s/%s", config.Global.Logger.S3LogPath, time.Now().Format("2006-01"))
		err              error
	)

	// Check if logger is not 's3' mode
	if config.Global.Logger.LogMode != "s3" {
		return
	}

	// Reopen the log file
	localLogFile, err = os.Open(l.LogFilePath)
	if err != nil {
		log.Fatal(err)
	}

	// Get local log file name
	split := strings.Split(l.LogFilePath, "/")
	localLogFileName = split[len(split)-1]

	// Initial file system for S3 Log keeper
	logFileStorage := storage.NewFileSystem(
		"s3",
		storage.WithS3Endpoint(config.Global.Logger.S3Config.Endpoint),
		storage.WithS3Bucket(config.Global.Logger.S3Config.Bucket),
		storage.WithS3Region(config.Global.Logger.S3Config.Region),
		storage.WithS3AccessKeyID(config.Global.Logger.S3Config.AccessKeyID),
		storage.WithS3SecretKey(config.Global.Logger.S3Config.SecretKey),
		storage.WithS3UseSSL(config.Global.Logger.S3Config.UseSSL),
		storage.WithS3Prefix(config.Global.Logger.S3Config.Prefix),
		storage.WithSkipPrintLog(true),
	)

	// Upload the log file to s3 (Replace)
	err = logFileStorage.Put(ctx, s3LogPath, localLogFileName, localLogFile)
	if err != nil {
		log.Fatal("[App] LoggerS3UploadError: ", err)
	}
}
