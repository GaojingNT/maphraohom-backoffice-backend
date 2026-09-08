package file_module

import (
	"maphraohom.app/maphraohom-backoffice/pkg/storage"
	"maphraohom.app/maphraohom-backoffice/pkg/tracing"
)

type (
	module struct {
		tracer     *tracing.MyTracer
		fileSystem *storage.FileSystem
	}
	Controller struct {
		m *module
	}
	Service struct {
		tracer     *tracing.MyTracer
		fileSystem *storage.FileSystem
	}
)

func NewModule() module {
	return module{
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
	}
}

// Register controller services
func (c Controller) fileService() Service { return NewService() }
