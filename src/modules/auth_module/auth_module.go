package auth_module

import (
	"gorm.io/gorm"
	"maphraohom.app/maphraohom-backoffice/pkg/cache"
	"maphraohom.app/maphraohom-backoffice/pkg/database"
	"maphraohom.app/maphraohom-backoffice/pkg/logger"
	"maphraohom.app/maphraohom-backoffice/pkg/storage"
	"maphraohom.app/maphraohom-backoffice/pkg/tracing"
)

type (
	module struct {
		db         *gorm.DB
		cacher     *cache.Cache
		tracer     *tracing.MyTracer
		fileSystem *storage.FileSystem
	}

	Controller struct {
		m *module
	}

	Service struct {
		tracer     *tracing.MyTracer
		logger     *logger.Logger
		fileSystem *storage.FileSystem
	}

	Repository struct {
		db         *gorm.DB
		tracer     *tracing.MyTracer
		logger     *logger.Logger
		fileSystem *storage.FileSystem
	}
)

func NewModule() module {
	return module{
		db:         database.CurrentDatabase(),
		cacher:     cache.CurrentCacher(),
		tracer:     tracing.CurrentTracer(),
		fileSystem: storage.CurrentFileStorage(),
	}
}

func (m *module) Controller() Controller {
	return Controller{
		m: m,
	}
}

func NewService(
	tracer *tracing.MyTracer,
	logger *logger.Logger,
	fileSystem *storage.FileSystem,
) Service {
	return Service{
		tracer:     tracer,
		logger:     logger,
		fileSystem: fileSystem,
	}
}

func NewRepository(
	db *gorm.DB,
	tracer *tracing.MyTracer,
	logger *logger.Logger,
	fileSystem *storage.FileSystem,
) Repository {
	return Repository{
		db:         db,
		tracer:     tracer,
		logger:     logger,
		fileSystem: fileSystem,
	}
}

// Register controller services

func (c Controller) authService() Service {
	return NewService(tracing.CurrentTracer(), logger.CurrentLogger(), storage.CurrentFileStorage())
}

// Register controller repositories

func (s Service) authRepository() Repository {
	return NewRepository(database.CurrentDatabase(), tracing.CurrentTracer(), logger.CurrentLogger(), storage.CurrentFileStorage())
}
