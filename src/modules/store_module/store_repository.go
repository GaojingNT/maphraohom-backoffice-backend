package store_module

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/shopspring/decimal"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/internal/utils"
	"maphraohom.app/maphraohom-backoffice/pkg/database/paginator"
	"maphraohom.app/maphraohom-backoffice/src/models"
)

func (r Repository) GetStorePaginate(ctx context.Context, pagination *paginator.Pagination) (*paginator.Pagination, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetStorePaginateRepository", trace.WithAttributes(attribute.String("repository", "GetStorePaginate")))
		stores       = make([]models.Store, 0)
		err          error
	)

	searchAttribute, _ := pagination.GetStringAttribute("search")
	searchByAttribute, _ := pagination.GetStringAttribute("search_by")

	tx := r.db

	utils.Block{
		Try: func() {
			if err = tx.
				Scopes(models.SearchingScope(models.StoreSearchable(), searchAttribute, searchByAttribute)).
				Scopes(paginator.Paginate(stores, pagination, tx)).
				Find(&stores).Error; err != nil {
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

	pagination.Data = stores

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	return pagination, nil
}

func (r Repository) GetStoreByID(ctx context.Context, id int) (models.Store, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetStoreByIDRepository", trace.WithAttributes(attribute.String("repository", "GetStoreByID"), attribute.Int64("id", int64(id))))
		store        models.Store
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Preload("Owners").First(&store, id).Error; err != nil {
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
		return store, err
	}

	return store, nil
}

// UpdateStoreInput is the editable, non-image subset of a store — name,
// address, phone. The logo is only ever changed through its own upload/delete
// methods below.
type UpdateStoreInput struct {
	Name    string
	Address string
	Phone   string
}

func (r Repository) UpdateStore(ctx context.Context, id int, input UpdateStoreInput) (models.Store, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "UpdateStoreRepository", trace.WithAttributes(attribute.String("repository", "UpdateStore"), attribute.Int64("id", int64(id))))
		store        models.Store
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.First(&store, id).Error; err != nil {
				utils.Throw(err)
			}

			store.Name = input.Name
			store.Address = input.Address
			store.Phone = input.Phone

			if err = r.db.Save(&store).Error; err != nil {
				utils.Throw(err)
			}

			// Reload with owners for the response — loaded only after Save
			// so Save never touches the ownership rows.
			if err = r.db.Preload("Owners").First(&store, id).Error; err != nil {
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
		return store, err
	}

	return store, nil
}

// getStoreImageKey and updateStoreImage back GetStoreLogoKey/UpdateStoreLogo
// below — kept column-generic so another image column on tbl_stores only
// needs a pair of one-line wrappers.
func (r Repository) getStoreImageKey(ctx context.Context, id int, column string, spanName string) (string, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, spanName, trace.WithAttributes(attribute.String("repository", spanName), attribute.Int64("id", int64(id))))
		store        models.Store
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Select("id", column).First(&store, id).Error; err != nil {
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
		return "", err
	}

	return store.Logo, nil
}

func (r Repository) updateStoreImage(ctx context.Context, id int, column string, key string, spanName string) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, spanName, trace.WithAttributes(attribute.String("repository", spanName), attribute.Int64("id", int64(id))))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Model(&models.Store{}).Where("id = ?", id).Update(column, key).Error; err != nil {
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

	return err
}

func (r Repository) GetStoreLogoKey(ctx context.Context, id int) (string, error) {
	return r.getStoreImageKey(ctx, id, "logo", "GetStoreLogoKeyRepository")
}

func (r Repository) UpdateStoreLogo(ctx context.Context, id int, key string) error {
	return r.updateStoreImage(ctx, id, "logo", key, "UpdateStoreLogoRepository")
}

// lastPriceRow is the scan target for GetLastPrices' DISTINCT ON query.
type lastPriceRow struct {
	ProductID int
	Price     decimal.Decimal
}

// GetLastPrices returns, for every product, the price used in this store's
// most recent bill of the given type (bill_items.price snapshot) — used to
// prefill the create-bill form. Buy and sell prices never mix because the
// query is scoped by type. Soft-deleted bills are excluded.
func (r Repository) GetLastPrices(ctx context.Context, storeID int, billType string) ([]lastPriceRow, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetLastPricesRepository", trace.WithAttributes(attribute.String("repository", "GetLastPrices"), attribute.Int("storeId", storeID), attribute.String("type", billType)))
		rows         = make([]lastPriceRow, 0)
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.
				Table("tbl_bill_items AS bi").
				Select("DISTINCT ON (bi.product_id) bi.product_id AS product_id, bi.price AS price").
				Joins("JOIN tbl_bills b ON b.id = bi.bill_id").
				Where("b.store_id = ? AND b.type = ? AND b.deleted_at IS NULL", storeID, billType).
				Order("bi.product_id, b.created_at DESC").
				Scan(&rows).Error; err != nil {
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

	return rows, nil
}
