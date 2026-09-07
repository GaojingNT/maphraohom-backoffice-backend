package scheduler

import "go.uber.org/zap"

// Log will log a message
func (ctx *SchedulerContext) Log(tag string, message string) {
	ctx.sc.Logger.Info(message, zap.String("tag", tag))
}
