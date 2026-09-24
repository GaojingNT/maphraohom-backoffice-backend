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
			if err = r.db.First(&store, id).Error; err != nil {
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
