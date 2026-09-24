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

func (r Repository) GetCustomerPhones(ctx context.Context, customerID int) ([]models.CustomerPhone, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetCustomerPhonesRepository", trace.WithAttributes(attribute.String("repository", "GetCustomerPhones"), attribute.Int("customerId", customerID)))
		phones       = make([]models.CustomerPhone, 0)
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.
				Where("customer_id = ?", customerID).
				Order("is_default DESC, id DESC").
				Find(&phones).Error; err != nil {
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

	return phones, nil
}

func (r Repository) CreateCustomer(ctx context.Context, name string) (models.Customer, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CreateCustomerRepository", trace.WithAttributes(attribute.String("repository", "CreateCustomer")))
		customer     = models.Customer{Name: name}
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Create(&customer).Error; err != nil {
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
		return customer, err
	}

	return customer, nil
}

// CreateCustomerAddress adds a new address for a customer. When isDefault is
// true, every other address of this customer is demoted first so at most
// one stays default.
func (r Repository) CreateCustomerAddress(ctx context.Context, customerID int, address string, isDefault bool) (models.CustomerAddress, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CreateCustomerAddressRepository", trace.WithAttributes(attribute.String("repository", "CreateCustomerAddress"), attribute.Int("customerId", customerID)))
		record       models.CustomerAddress
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				if txErr := tx.First(&models.Customer{}, customerID).Error; txErr != nil {
					return txErr
				}

				if isDefault {
					if txErr := tx.Model(&models.CustomerAddress{}).
						Where("customer_id = ?", customerID).
						Update("is_default", false).Error; txErr != nil {
						return txErr
					}
				}

				record = models.CustomerAddress{
					CustomerID: customerID,
					Address:    address,
					IsDefault:  isDefault,
				}
				return tx.Create(&record).Error
			}); err != nil {
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
		return record, err
	}

	return record, nil
}

// CreateCustomerPhone adds a new phone number for a customer. When isDefault
// is true, every other phone of this customer is demoted first so at most
// one stays default.
func (r Repository) CreateCustomerPhone(ctx context.Context, customerID int, phone string, isDefault bool) (models.CustomerPhone, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CreateCustomerPhoneRepository", trace.WithAttributes(attribute.String("repository", "CreateCustomerPhone"), attribute.Int("customerId", customerID)))
		record       models.CustomerPhone
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				if txErr := tx.First(&models.Customer{}, customerID).Error; txErr != nil {
					return txErr
				}

				if isDefault {
					if txErr := tx.Model(&models.CustomerPhone{}).
						Where("customer_id = ?", customerID).
						Update("is_default", false).Error; txErr != nil {
						return txErr
					}
				}

				record = models.CustomerPhone{
					CustomerID: customerID,
					Phone:      phone,
					IsDefault:  isDefault,
				}
				return tx.Create(&record).Error
			}); err != nil {
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
		return record, err
	}

	return record, nil
}
