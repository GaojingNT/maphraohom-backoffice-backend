package platform_type_module

import (
	"context"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/internal/utils"
	"maphraohom.app/maphraohom-backoffice/src/models"
)

func (r Repository) GetActivePlatformTypes(ctx context.Context) ([]models.PlatformType, error) {
	var (
		_, childSpan  = r.tracer.TraceStart(ctx, "GetActivePlatformTypesRepository", trace.WithAttributes(attribute.String("repository", "GetActivePlatformTypes")))
		platformTypes = make([]models.PlatformType, 0)
		err           error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Where("is_active = ?", true).Order("sort_order ASC, id ASC").Find(&platformTypes).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			r.logger.Error(err.Error())
			sentry.CaptureException(err)
			exception.SqlErrorMessage = err.Error()
			err = exception.ErrDbQueryStatement
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	return platformTypes, nil
}
