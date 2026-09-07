package logger

import (
	"context"
	"fmt"

	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
)

type dbCore struct {
	zapcore.LevelEnabler
	db     *gorm.DB
	fields map[string]interface{}
}

func (c *dbCore) With(fields []zapcore.Field) zapcore.Core {
	clonedFields := make(map[string]interface{}, len(c.fields))
	for k, v := range c.fields {
		clonedFields[k] = v
	}
	for _, f := range fields {
		clonedFields[f.Key] = f.Interface
	}
	return &dbCore{
		LevelEnabler: c.LevelEnabler,
		db:           c.db,
		fields:       clonedFields,
	}
}

func (c *dbCore) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return checkedEntry.AddCore(entry, c)
	}
	return checkedEntry
}

func (c *dbCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	jsonFields := make(map[string]interface{}, len(c.fields)+len(fields))
	for k, v := range c.fields {
		jsonFields[k] = v
	}
	for _, f := range fields {
		jsonFields[f.Key] = f.String
	}

	log := Log{
		Timestamp: entry.Time,
		Level:     entry.Level.String(),
		Message:   entry.Message,
		Fields:    jsonFields,
	}

	if err := c.db.WithContext(context.Background()).Create(&log).Error; err != nil {
		return fmt.Errorf("failed to write log to database: %w", err)
	}
	return nil
}

func (c *dbCore) Sync() error {
	return nil
}

func NewZapDBCore(db *gorm.DB, level zapcore.LevelEnabler) zapcore.Core {
	return &dbCore{
		LevelEnabler: level,
		db:           db,
		fields:       make(map[string]interface{}),
	}
}
