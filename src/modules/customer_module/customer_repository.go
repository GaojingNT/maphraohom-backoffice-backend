package customer_module

import (
	"context"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/internal/utils"
	"maphraohom.app/maphraohom-backoffice/pkg/database/paginator"
	"maphraohom.app/maphraohom-backoffice/src/models"
)

func (r Repository) GetCustomerPaginate(ctx context.Context, pagination *paginator.Pagination) (*paginator.Pagination, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetCustomerPaginateRepository", trace.WithAttributes(attribute.String("repository", "GetCustomerPaginate")))
		customers    = make([]models.Customer, 0)
		err          error
	)

	searchAttribute, _ := pagination.GetStringAttribute("search")
	searchByAttribute, _ := pagination.GetStringAttribute("search_by")

	tx := r.db

	utils.Block{
		Try: func() {
			if err = tx.
				Scopes(models.SearchingScope(models.CustomerSearchable(), searchAttribute, searchByAttribute)).
				Scopes(paginator.Paginate(customers, pagination, tx)).
				Find(&customers).Error; err != nil {
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

	pagination.Data = customers

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	return pagination, nil
}

func (r Repository) GetCustomerByID(ctx context.Context, id int) (models.Customer, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetCustomerByIDRepository", trace.WithAttributes(attribute.String("repository", "GetCustomerByID"), attribute.Int64("id", int64(id))))
		customer     models.Customer
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.First(&customer, id).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else {
				err = e.(error)
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return customer, err
	}

	return customer, nil
}

func (r Repository) GetCustomerAddresses(ctx context.Context, customerID int) ([]models.CustomerAddress, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetCustomerAddressesRepository", trace.WithAttributes(attribute.String("repository", "GetCustomerAddresses"), attribute.Int("customerId", customerID)))
		addresses    = make([]models.CustomerAddress, 0)
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.
				Where("customer_id = ?", customerID).
				Order("is_default DESC, id DESC").
				Find(&addresses).Error; err != nil {
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

	return addresses, nil
}
