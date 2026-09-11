package store_module

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/internal/pricing"
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

// GetStoreProducts returns the store's base price for every product that
// has one configured (one row per product — store_product_prices has a
// unique (store_id, product_id) row, no history).
func (r Repository) GetStoreProducts(ctx context.Context, storeID int) ([]models.StoreProductPrice, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetStoreProductsRepository", trace.WithAttributes(attribute.String("repository", "GetStoreProducts"), attribute.Int("storeId", storeID)))
		prices       = make([]models.StoreProductPrice, 0)
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.
				Preload("Product").
				Where("store_id = ?", storeID).
				Order("product_id ASC").
				Find(&prices).Error; err != nil {
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

	return prices, nil
}

// GetActivePromotion returns the store's currently active promotion, or nil
// if none is active (not treated as an error).
func (r Repository) GetActivePromotion(ctx context.Context, storeID int) (*models.Promotion, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetActivePromotionRepository", trace.WithAttributes(attribute.String("repository", "GetActivePromotion"), attribute.Int("storeId", storeID)))
		promotion    models.Promotion
		notFound     bool
		err          error
	)

	utils.Block{
		Try: func() {
			if promotion, err = pricing.ActivePromotion(r.db, storeID, time.Now()); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			if err == gorm.ErrRecordNotFound {
				notFound = true
				err = nil
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
		return nil, err
	}
	if notFound {
		return nil, nil
	}

	return &promotion, nil
}

// GetStoreProduct returns one product's resolved price (promotion-first)
// at this store, along with the product itself.
func (r Repository) GetStoreProduct(ctx context.Context, storeID int, productID int) (models.Product, pricing.Resolved, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetStoreProductRepository", trace.WithAttributes(attribute.String("repository", "GetStoreProduct"), attribute.Int("storeId", storeID), attribute.Int("productId", productID)))
		product      models.Product
		resolved     pricing.Resolved
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.First(&product, productID).Error; err != nil {
				utils.Throw(err)
			}

			resolved, err = pricing.ResolveOne(r.db, storeID, productID, time.Now())
			if err != nil {
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
		return product, resolved, err
	}

	return product, resolved, nil
}
