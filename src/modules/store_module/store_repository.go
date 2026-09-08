package store_module

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
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

// GetStoreProducts returns the currently effective price for every product
// that has ever been priced at this store (one row per product — the one
// with the latest effective_from that isn't in the future).
func (r Repository) GetStoreProducts(ctx context.Context, storeID int) ([]models.StoreProductPrice, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetStoreProductsRepository", trace.WithAttributes(attribute.String("repository", "GetStoreProducts"), attribute.Int("storeId", storeID)))
		allPrices    = make([]models.StoreProductPrice, 0)
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.
				Preload("Product").
				Where("store_id = ? AND effective_from <= ?", storeID, time.Now()).
				Order("product_id ASC, effective_from DESC").
				Find(&allPrices).Error; err != nil {
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

	// Keep only the newest row per product (allPrices is already ordered
	// product_id ASC, effective_from DESC, so the first occurrence wins).
	seen := make(map[int]bool, len(allPrices))
	latest := make([]models.StoreProductPrice, 0, len(allPrices))
	for _, price := range allPrices {
		if seen[price.ProductID] {
			continue
		}
		seen[price.ProductID] = true
		latest = append(latest, price)
	}

	return latest, nil
}

// GetStoreProduct returns one product's currently effective price at this store.
func (r Repository) GetStoreProduct(ctx context.Context, storeID int, productID int) (models.StoreProductPrice, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetStoreProductRepository", trace.WithAttributes(attribute.String("repository", "GetStoreProduct"), attribute.Int("storeId", storeID), attribute.Int("productId", productID)))
		price        models.StoreProductPrice
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.
				Preload("Product").
				Where("store_id = ? AND product_id = ? AND effective_from <= ?", storeID, productID, time.Now()).
				Order("effective_from DESC").
				First(&price).Error; err != nil {
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
		return price, err
	}

	return price, nil
}
