package logger

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/Devatoria/go-graylog"
	"go.uber.org/zap/zapcore"
	"maphraohom.app/maphraohom-backoffice/config"
)

type graylogCore struct {
	zapcore.LevelEnabler
	config          *config.Config
	graylogInstance *graylog.Graylog
	fields          map[string]interface{}
}

func (c *graylogCore) With(fields []zapcore.Field) zapcore.Core {
	clonedFields := make(map[string]interface{}, len(c.fields))
	for k, v := range c.fields {
		clonedFields[k] = v
	}
	for _, f := range fields {
		clonedFields[f.Key] = f.Interface
	}
	return &graylogCore{
		LevelEnabler: c.LevelEnabler,
		fields:       clonedFields,
	}
}

func (c *graylogCore) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return checkedEntry.AddCore(entry, c)
	}
	return checkedEntry
}

func (c *graylogCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
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

	// Create a new message
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

	// Send log to graylog
	if err := c.graylogInstance.Send(graylogMessage); err != nil {
		return fmt.Errorf("failed to write log to graylog: %w", err)
	}

	// Print log to console
	fmt.Println(time.Now().Format("2006/01/02 15:04:05"), "Log:", string(jsonMessage))

	return nil
}

func (c *graylogCore) Sync() error {
	return nil
}

func NewZapGraylogCore(config *config.Config, level zapcore.LevelEnabler) zapcore.Core {
	// Initialize a new graylog client with TLS
	graylogInstance, err := graylog.NewGraylog(
		graylog.Endpoint{
			Transport: graylog.TCP,
			Address:   config.Logger.GraylogConfig.Endpoint,
			Port:      uint(config.Logger.GraylogConfig.Port),
		},
	)
	if err != nil {
		log.Fatal("[App] GraylogLoggerError: ", err)
	}

	return &graylogCore{
		LevelEnabler:    level,
		config:          config,
		graylogInstance: graylogInstance,
		fields:          make(map[string]interface{}),
	}
}

func NewZapGraylogCoreWithLTS(config *config.Config, level zapcore.LevelEnabler) zapcore.Core {
	// Initialize a new graylog client with TLS
	graylogInstance, err := graylog.NewGraylogTLS(
		graylog.Endpoint{
			Transport: graylog.TCP,
			Address:   config.Logger.GraylogConfig.Endpoint,
			Port:      uint(config.Logger.GraylogConfig.Port),
		},
		3*time.Second,
		&tls.Config{
			InsecureSkipVerify: config.Logger.GraylogConfig.TlsSkipVerify,
		},
	)
	if err != nil {
		log.Fatal("[App] GraylogLoggerError: ", err)
	}

	return &graylogCore{
		LevelEnabler:    level,
		config:          config,
		graylogInstance: graylogInstance,
		fields:          make(map[string]interface{}),
	}
}
