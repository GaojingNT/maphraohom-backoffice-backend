package middlewares

import (
	"gorm.io/gorm"
	"maphraohom.app/maphraohom-backoffice/pkg/cache"
	"maphraohom.app/maphraohom-backoffice/pkg/database"
	"maphraohom.app/maphraohom-backoffice/pkg/logger"
	"maphraohom.app/maphraohom-backoffice/pkg/tracing"
)

type Middleware struct {
	db     *gorm.DB
	cacher *cache.Cache
	tracer *tracing.MyTracer
	logger *logger.Logger
}

func NewMiddleware() *Middleware {
	return &Middleware{
		db:     database.CurrentDatabase(),
		cacher: cache.CurrentCacher(),
		tracer: tracing.CurrentTracer(),
		logger: logger.CurrentLogger(),
	}
}
