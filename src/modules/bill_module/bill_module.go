package bill_module

import (
	"gorm.io/gorm"
	"maphraohom.app/maphraohom-backoffice/pkg/database"
	"maphraohom.app/maphraohom-backoffice/pkg/logger"
	"maphraohom.app/maphraohom-backoffice/pkg/storage"
	"maphraohom.app/maphraohom-backoffice/pkg/tracing"
)

type (
	module struct {
		db         *gorm.DB
		tracer     *tracing.MyTracer
		fileSystem *storage.FileSystem
	}
	Controller struct {
		m *module
	}
	Service struct {
		tracer     *tracing.MyTracer
		fileSystem *storage.FileSystem
		log        *logger.Logger
	}
	Repository struct {
		db     *gorm.DB
		tracer *tracing.MyTracer
		logger *logger.Logger
	}
)

func NewModule() module {
	return module{
		db:         database.CurrentDatabase(),
		tracer:     tracing.CurrentTracer(),
		fileSystem: storage.CurrentFileStorage(),
	}
}

func (m *module) Controller() Controller {
	return Controller{
		m: m,
	}
}

func NewService() Service {
	return Service{
		tracer:     tracing.CurrentTracer(),
		fileSystem: storage.CurrentFileStorage(),
		log:        logger.CurrentLogger(),
	}
}

func NewRepository() Repository {
	return Repository{
		db:     database.CurrentDatabase(),
		tracer: tracing.CurrentTracer(),
		logger: logger.CurrentLogger(),
	}
}

// Register controller services
func (c Controller) billService() Service { return NewService() }

// Register service repositories
func (s Service) billRepository() Repository { return NewRepository() }
