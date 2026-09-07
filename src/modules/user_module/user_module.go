package user_module

import (
	"gorm.io/gorm"
	"maphraohom.app/maphraohom-backoffice/pkg/cache"
	"maphraohom.app/maphraohom-backoffice/pkg/database"
	"maphraohom.app/maphraohom-backoffice/pkg/logger"
	"maphraohom.app/maphraohom-backoffice/pkg/storage"
	"maphraohom.app/maphraohom-backoffice/pkg/tracing"
	pbUser "maphraohom.app/maphraohom-backoffice/src/modules/user_module/proto/user"
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
	GrpcController struct {
		m *module
		pbUser.UnimplementedUserServiceServer
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

func (m *module) GrpcController() GrpcController {
	return GrpcController{
		m: m,
	}
}

func NewService() Service {
	return Service{
		tracer:     tracing.CurrentTracer(),
		logger:     logger.CurrentLogger(),
		fileSystem: storage.CurrentFileStorage(),
	}
}

func NewRepository() Repository {
	return Repository{
		db:         database.CurrentDatabase(),
		tracer:     tracing.CurrentTracer(),
		logger:     logger.CurrentLogger(),
		fileSystem: storage.CurrentFileStorage(),
	}
}

// Register controller services
func (c Controller) userService() Service     { return NewService() }
func (c GrpcController) userService() Service { return NewService() }

// Register service repositories
func (s Service) userRepository() Repository { return NewRepository() }
