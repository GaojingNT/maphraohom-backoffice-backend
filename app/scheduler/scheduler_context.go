package scheduler

import (
	"log"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
	"maphraohom.app/maphraohom-backoffice/config"
	"maphraohom.app/maphraohom-backoffice/pkg/cache"
	"maphraohom.app/maphraohom-backoffice/pkg/logger"
	"maphraohom.app/maphraohom-backoffice/pkg/storage"
	"maphraohom.app/maphraohom-backoffice/pkg/tracing"
)

var GlobalScheduler IScheduler

func CurrentScheduler() IScheduler {
	return GlobalScheduler
}

type (
	// IContext is the context for service
	IContext interface {
		Log(tag string, message string)
	}

	IScheduler interface {
		NewJob(cronString string, h ServiceHandleFunc)
		Log() *logger.Logger
		GetTaskCount() int
		IncreaseTaskCount()
		Start()
		Shutdown()
	}

	// Scheduler implement IScheduler it is context for Scheduler service
	scheduler struct {
		Config     *config.Config
		DbClient   *gorm.DB
		Cacher     *cache.Cache
		Tracer     *tracing.MyTracer
		Logger     *logger.Logger
		FileSystem *storage.FileSystem
		cron       gocron.Scheduler
		TaskCount  int
	}

	// SchedulerContext implement IContext it is context for Scheduler
	SchedulerContext struct {
		sc *scheduler
	}

	ISchedulerContext interface {
		IContext
	}

	// ServiceHandleFunc is the handler for each Microservice
	ServiceHandleFunc func(ctx ISchedulerContext) error
)

// NewScheduler is the constructor function for Scheduler
func NewScheduler(
	config *config.Config,
	dbClient *gorm.DB,
	cacher *cache.Cache,
	tracer *tracing.MyTracer,
	fileSystem *storage.FileSystem,
) IScheduler {
	// Create a GoCron instance
	cron, err := gocron.NewScheduler(
		gocron.WithStopTimeout(1 * time.Second),
	)
	if err != nil {
		logger.CurrentLogger().Error("Error creating scheduler", zapcore.Field{Key: "error", Type: zapcore.ErrorType, Interface: err})
		return nil
	}

	return &scheduler{
		Config:     config,
		DbClient:   dbClient,
		Cacher:     cacher,
		Tracer:     tracer,
		FileSystem: fileSystem,
		cron:       cron,
	}
}

// NewSchedulerContext is the constructor function for SchedulerContext
func NewSchedulerContext(sc *scheduler) *SchedulerContext {
	return &SchedulerContext{
		sc: sc,
	}
}

// Start start all registered services
func (s *scheduler) NewJob(cronString string, h ServiceHandleFunc) {
	var err error

	// Check if is child process
	if fiber.IsChild() {
		return
	}

	// add a job to the scheduler
	if _, err = s.cron.NewJob(
		gocron.CronJob(cronString, false),
		gocron.NewTask(h, NewSchedulerContext(s)),
	); err != nil {
		s.Log().Error(
			"Error creating schedule job",
			zapcore.Field{Key: "error", Type: zapcore.ErrorType, Interface: err},
		)
	}

	// Increase task count
	s.IncreaseTaskCount()
}

// Log message to console
func (s *scheduler) Log() *logger.Logger {
	return logger.CurrentLogger()
}

// GetTaskCount get the number of registered tasks
func (s *scheduler) GetTaskCount() int {
	return s.TaskCount
}

// Increase task count
func (s *scheduler) IncreaseTaskCount() {
	s.TaskCount++
}

// Start start all registered scheduler jobs
func (s *scheduler) Start() {
	s.cron.Start()

	if !fiber.IsChild() {
		// Print service start
		s.Log().Info("Scheduler service is running in background process")
	}
}

// Shutdown stop all registered scheduler jobs
func (s *scheduler) Shutdown() {
	s.cron.Shutdown()

	if !fiber.IsChild() {
		// Print service start
		log.Println("Scheduler", "Shutting down...")
		s.Log().Info("Scheduler service shutting down...")
	}
}
