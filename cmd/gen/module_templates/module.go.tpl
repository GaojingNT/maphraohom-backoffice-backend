package {{.ModuleName}}_module

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
	controller struct {
		m *module
	}
	service struct {
		tracer     *tracing.MyTracer
		logger     *logger.Logger
		fileSystem *storage.FileSystem
	}
	repository struct {
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

func (m *module) Controller() controller {
	return controller{
		m: m,
	}
}

func NewService() IService {
	return service{
		tracer:     tracing.CurrentTracer(),
		logger:     logger.CurrentLogger(),
		fileSystem: storage.CurrentFileStorage(),
	}
}

func NewRepository() IRepository {
	return repository{
		db:         database.CurrentDatabase(),
		tracer:     tracing.CurrentTracer(),
		logger:     logger.CurrentLogger(),
		fileSystem: storage.CurrentFileStorage(),
	}
}

// Register controller services
func (c Controller) {{.CamelModuleName}}Service() IService     { return NewService() }

// Register service repositories
func (s Service) {{.CamelModuleName}}Repository() IRepository { return NewRepository() }
