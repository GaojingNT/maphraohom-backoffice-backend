package product_module

import (
	"context"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/internal/utils"
	"maphraohom.app/maphraohom-backoffice/src/models"
)

func (r Repository) GetProducts(ctx context.Context) ([]models.Product, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetProductsRepository", trace.WithAttributes(attribute.String("repository", "GetProducts")))
		products     = make([]models.Product, 0)
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Order("id ASC").Find(&products).Error; err != nil {
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

	return products, nil
}
