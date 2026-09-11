package promotion_module

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/internal/utils"
	"maphraohom.app/maphraohom-backoffice/src/models"
)

// CreatePromotionItemInput is one product's special price within a
// create-promotion request.
type CreatePromotionItemInput struct {
	ProductID int
	Price     float64
}

// CreatePromotionInput carries the already-validated fields for creating a
// promotion and its per-product prices.
type CreatePromotionInput struct {
	Name     string
	StoreID  int
	StartsAt time.Time
	EndsAt   time.Time
	Items    []CreatePromotionItemInput
}

// GetPromotions lists promotions, optionally filtered to one store.
// storeID <= 0 means no store filter.
func (r Repository) GetPromotions(ctx context.Context, storeID int) ([]models.Promotion, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetPromotionsRepository", trace.WithAttributes(attribute.String("repository", "GetPromotions"), attribute.Int("storeId", storeID)))
		promotions   = make([]models.Promotion, 0)
		err          error
	)

	tx := r.db
	if storeID > 0 {
		tx = tx.Where("store_id = ?", storeID)
	}

	utils.Block{
		Try: func() {
			if err = tx.
				Preload("Store").
				Order("starts_at DESC").
				Find(&promotions).Error; err != nil {
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

	return promotions, nil
}

// GetPromotionByID returns one promotion with its Store and per-product
// prices preloaded.
func (r Repository) GetPromotionByID(ctx context.Context, id int) (models.Promotion, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetPromotionByIDRepository", trace.WithAttributes(attribute.String("repository", "GetPromotionByID"), attribute.Int("id", id)))
		promotion    models.Promotion
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Preload("Store").Preload("Prices.Product").First(&promotion, id).Error; err != nil {
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
		return promotion, err
	}

	return promotion, nil
}

// CreatePromotion validates that the new promotion's time range does not
// overlap any existing promotion for the same store (a store may have at
// most one active promotion at a time — checked at the store level, not
// per-product), then creates it with its per-product prices in one
// transaction.
func (r Repository) CreatePromotion(ctx context.Context, input CreatePromotionInput) (models.Promotion, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CreatePromotionRepository", trace.WithAttributes(attribute.String("repository", "CreatePromotion")))
		promotion    models.Promotion
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				var conflicts []models.Promotion
				if txErr := tx.
					Where("store_id = ? AND starts_at < ? AND ends_at > ?", input.StoreID, input.EndsAt, input.StartsAt).
					Find(&conflicts).Error; txErr != nil {
					return txErr
				}
				if len(conflicts) > 0 {
					exception.PromotionOverlapMessage = conflicts[0].Name
					return exception.ErrPromotionOverlap
				}

				promotion = models.Promotion{
					Name:     input.Name,
					StoreID:  input.StoreID,
					StartsAt: input.StartsAt,
					EndsAt:   input.EndsAt,
				}
				if txErr := tx.Create(&promotion).Error; txErr != nil {
					return txErr
				}

				prices := make([]models.PromotionPrice, 0, len(input.Items))
				for _, item := range input.Items {
					prices = append(prices, models.PromotionPrice{
						PromotionID: promotion.ID,
						ProductID:   item.ProductID,
						Price:       item.Price,
					})
				}
				if txErr := tx.Create(&prices).Error; txErr != nil {
					return txErr
				}

				promotion.Prices = prices

				return nil
			}); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			switch e {
			case exception.ErrPromotionOverlap:
				err = exception.ErrPromotionOverlap
			default:
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
		return promotion, err
	}

	// Reload with Store + Prices.Product preloaded to match the shape
	// responses.PromotionDetailResponse.Make expects.
	if err = r.db.Preload("Store").Preload("Prices.Product").First(&promotion, promotion.ID).Error; err != nil {
		return promotion, err
	}

	return promotion, nil
}

// UpdatePromotionInput carries the already-validated fields for replacing a
// promotion's editable fields and its full set of per-product prices.
type UpdatePromotionInput struct {
	Name     string
	StoreID  int
	StartsAt time.Time
	EndsAt   time.Time
	Items    []CreatePromotionItemInput
}

// UpdatePromotion re-validates the overlap rule (excluding this promotion
// itself), then replaces the promotion's fields and prices wholesale in one
// transaction.
func (r Repository) UpdatePromotion(ctx context.Context, id int, input UpdatePromotionInput) (models.Promotion, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "UpdatePromotionRepository", trace.WithAttributes(attribute.String("repository", "UpdatePromotion"), attribute.Int("id", id)))
		promotion    models.Promotion
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				if txErr := tx.First(&promotion, id).Error; txErr != nil {
					if txErr == gorm.ErrRecordNotFound {
						return exception.ErrRecordNotFound
					}
					return txErr
				}

				var conflicts []models.Promotion
				if txErr := tx.
					Where("store_id = ? AND id <> ? AND starts_at < ? AND ends_at > ?", input.StoreID, id, input.EndsAt, input.StartsAt).
					Find(&conflicts).Error; txErr != nil {
					return txErr
				}
				if len(conflicts) > 0 {
					exception.PromotionOverlapMessage = conflicts[0].Name
					return exception.ErrPromotionOverlap
				}

				promotion.Name = input.Name
				promotion.StoreID = input.StoreID
				promotion.StartsAt = input.StartsAt
				promotion.EndsAt = input.EndsAt
				if txErr := tx.Save(&promotion).Error; txErr != nil {
					return txErr
				}

				if txErr := tx.Where("promotion_id = ?", promotion.ID).Delete(&models.PromotionPrice{}).Error; txErr != nil {
					return txErr
				}

				prices := make([]models.PromotionPrice, 0, len(input.Items))
				for _, item := range input.Items {
					prices = append(prices, models.PromotionPrice{
						PromotionID: promotion.ID,
						ProductID:   item.ProductID,
						Price:       item.Price,
					})
				}
				if txErr := tx.Create(&prices).Error; txErr != nil {
					return txErr
				}

				promotion.Prices = prices

				return nil
			}); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			switch e {
			case exception.ErrRecordNotFound:
				err = exception.ErrRecordNotFound
			case exception.ErrPromotionOverlap:
				err = exception.ErrPromotionOverlap
			default:
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
		return promotion, err
	}

	if err = r.db.Preload("Store").Preload("Prices.Product").First(&promotion, promotion.ID).Error; err != nil {
		return promotion, err
	}

	return promotion, nil
}

// DeletePromotion hard-deletes a promotion (promotions keep no history — see
// spec). Its promotion_prices cascade-delete via FK; any bill_items that
// referenced it have their promotion_id set to NULL, keeping their price
// snapshot but dropping the "was this a promo sale" flag.
func (r Repository) DeletePromotion(ctx context.Context, id int) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "DeletePromotionRepository", trace.WithAttributes(attribute.String("repository", "DeletePromotion"), attribute.Int("id", id)))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Delete(&models.Promotion{}, id).Error; err != nil {
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

	return err
}
