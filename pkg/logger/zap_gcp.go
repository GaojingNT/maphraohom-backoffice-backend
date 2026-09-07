package logger

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"cloud.google.com/go/logging"
	"github.com/Devatoria/go-graylog"
	"github.com/dhduvall/gcloudzap"
	"github.com/goccy/go-json"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"maphraohom.app/maphraohom-backoffice/config"
)

type googleCloudLoggingCore struct {
	zapcore.LevelEnabler
	config *config.Config
	logger *zap.Logger
	fields map[string]interface{}
}

func (c *googleCloudLoggingCore) With(fields []zapcore.Field) zapcore.Core {
	clonedFields := make(map[string]interface{}, len(c.fields))
	for k, v := range c.fields {
		clonedFields[k] = v
	}
	for _, f := range fields {
		clonedFields[f.Key] = f.Interface
	}
	return &googleCloudLoggingCore{
		LevelEnabler: c.LevelEnabler,
		fields:       clonedFields,
	}
}

func (c *googleCloudLoggingCore) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return checkedEntry.AddCore(entry, c)
	}
	return checkedEntry
}

func (c *googleCloudLoggingCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	jsonFields := make(map[string]string, len(c.fields)+len(fields))
	for k, v := range c.fields {
		vStr, ok := v.(string)
		if !ok {
			jsonFields[k] = fmt.Sprintf("%+v", v)
			continue
		}

		jsonFields[k] = vStr
	}
	for _, f := range fields {
		jsonFields[f.Key] = f.String
	}

	entryLevel, _ := strconv.Atoi(entry.Level.String())

	// Set default values
	defaultMessageVersion := "1.1"
	defaultMessageHost := c.config.App.ServiceName

	if jsonFields["version"] == "" {
		jsonFields["version"] = defaultMessageVersion
	}
	if jsonFields["host"] == "" {
		jsonFields["host"] = defaultMessageHost
	}
	if jsonFields["short_message"] == "" {
		// Cut the message if it's too long (Max 250 characters)
		if len(entry.Message) > 250 {
			// Add ... to the end of the message
			jsonFields["short_message"] = entry.Message[:247] + "..."
		} else {
			jsonFields["short_message"] = entry.Message
		}
	}

	// Create a new message (GELF format)
	graylogMessage := graylog.Message{
		Version:      jsonFields["version"],
		Host:         jsonFields["host"],
		ShortMessage: jsonFields["short_message"],
		FullMessage:  entry.Message,
		Timestamp:    entry.Time.Unix(),
		Level:        uint(entryLevel),
		Extra:        jsonFields,
	}

	// Create JSON message
	jsonMessage, err := json.Marshal(graylogMessage)
	if err != nil {
		return fmt.Errorf("failed to convert log to JSON: %w", err)
	}

	// Send log to GCP with switch case by entry level
	switch entry.Level {
	case zapcore.DebugLevel:
		c.logger.Debug(string(jsonMessage), fields...)
	case zapcore.InfoLevel:
		c.logger.Info(string(jsonMessage), fields...)
	case zapcore.WarnLevel:
		c.logger.Warn(string(jsonMessage), fields...)
	case zapcore.ErrorLevel:
		c.logger.Error(string(jsonMessage), fields...)
	case zapcore.DPanicLevel:
		c.logger.DPanic(string(jsonMessage), fields...)
	case zapcore.PanicLevel:
		c.logger.Panic(string(jsonMessage), fields...)
	case zapcore.FatalLevel:
		c.logger.Fatal(string(jsonMessage), fields...)
	}

	// Print log to console
	fmt.Println(time.Now().Format("2006/01/02 15:04:05"), "Log:", string(jsonMessage))

	return nil
}

func (c *googleCloudLoggingCore) Sync() error {
	return nil
}

func encodeLevel() zapcore.LevelEncoder {
	return func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		switch l {
		case zapcore.DebugLevel:
			enc.AppendString("DEBUG")
		case zapcore.InfoLevel:
			enc.AppendString("INFO")
		case zapcore.WarnLevel:
			enc.AppendString("WARNING")
		case zapcore.ErrorLevel:
			enc.AppendString("ERROR")
		case zapcore.DPanicLevel:
			enc.AppendString("CRITICAL")
		case zapcore.PanicLevel:
			enc.AppendString("ALERT")
		case zapcore.FatalLevel:
			enc.AppendString("EMERGENCY")
		}
	}
}

func NewZapGoogleCloudLoggingCore(config *config.Config, level zapcore.LevelEnabler) zapcore.Core {
	// Initialize a new Google Cloud Logging client
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "severity",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    encodeLevel(),
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Logger configuration
	loggerCfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapcore.InfoLevel),
		Encoding:         "json",
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	// New Google Cloud Logging instance
	gcpLoggingClient, err := logging.NewClient(context.Background(), config.Logger.GoogleCloudLoggingConfig.ParentID)
	if err != nil {
		log.Fatal("[App] GoogleCloudLoggingError gcpLoggingClient: ", err)
	}

	// Create a logger
	logger, err := gcloudzap.New(loggerCfg, gcpLoggingClient, config.Logger.GoogleCloudLoggingConfig.LogID)
	if err != nil {
		log.Fatal("[App] GoogleCloudLoggingError logger: ", err)
	}

	return &googleCloudLoggingCore{
		LevelEnabler: level,
		config:       config,
		logger:       logger,
		fields:       make(map[string]interface{}),
	}
}
