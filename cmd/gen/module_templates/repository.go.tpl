package {{.ModuleName}}_module

import (
	"context"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"maphraohom.app/maphraohom-backoffice/pkg/database/paginator"
	"maphraohom.app/maphraohom-backoffice/internal/utils"
	"maphraohom.app/maphraohom-backoffice/src/models"
)

type (
	IRepository interface {
		Get{{.PascalModuleName}}Paginate(ctx context.Context, pagination *paginator.Pagination) (*paginator.Pagination, error)
		Get{{.PascalModuleName}}ByID(ctx context.Context, id int) (models.{{.PascalModuleName}}, error)
		Create{{.PascalModuleName}}(ctx context.Context, {{.CamelModuleName}} *models.{{.PascalModuleName}}) error
		Update{{.PascalModuleName}}(ctx context.Context, id int, {{.CamelModuleName}} *models.{{.PascalModuleName}}) error
		Delete{{.PascalModuleName}}(ctx context.Context, id int) error
	}
)

func (r Repository) Get{{.PascalModuleName}}Paginate(ctx context.Context, pagination *paginator.Pagination) (*paginator.Pagination, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "Get{{.PascalModuleName}}PaginateRepository", trace.WithAttributes(attribute.String("repository", "Get{{.PascalModuleName}}Paginate")))
		{{.CamelModulePluralName}}        = make([]models.{{.PascalModuleName}}, 0)
		err          error
	)

	// Get attributes
	searchAttribute, _ := pagination.GetStringAttribute("search")
	searchByAttribute, _ := pagination.GetStringAttribute("search_by")

	// Set tracing attributes
	r.tracer.SetAttributes(childSpan, attribute.String("search", searchAttribute))
	r.tracer.SetAttributes(childSpan, attribute.String("search_by", searchByAttribute))

	// Create pagination query
	tx := r.db.
		Preload("Creator").
		Preload("Updater")

	utils.Block{
		Try: func() {
			// Execute query
			if err = tx.
				Scopes(models.SearchingScope(models.{{.PascalModuleName}}Searchable(), searchAttribute, searchByAttribute)).
				Scopes(paginator.Paginate({{.CamelModulePluralName}}, pagination, tx)).
				Find(&{{.CamelModulePluralName}}).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			// Logging
			r.logger.Error(err.Error())
			sentry.CaptureException(err)
			exception.SqlErrorMessage = err.Error()
			err = exception.ErrDbQueryStatement
		},
		Finally: nil,
	}.Do()

	// Set data
	pagination.Data = {{.CamelModulePluralName}}

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return nil, err
	}

	return pagination, nil
}

func (r Repository) Get{{.PascalModuleName}}ByID(ctx context.Context, id int) (models.{{.PascalModuleName}}, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "Get{{.PascalModuleName}}ByIDRepository", trace.WithAttributes(attribute.String("repository", "Get{{.PascalModuleName}}ByID"), attribute.Int64("id", int64(id))))
		{{.CamelModuleName}}         models.{{.PascalModuleName}}
		err          error
	)

	utils.Block{
		Try: func() {
			// Execute query
			if err = r.db.First(&{{.CamelModuleName}}, id).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else {
				err = e.(error)
				// Logging
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return {{.CamelModuleName}}, err
	}

	return {{.CamelModuleName}}, nil
}

func (r Repository) Create{{.PascalModuleName}}(ctx context.Context, {{.CamelModuleName}} *models.{{.PascalModuleName}}) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "Create{{.PascalModuleName}}Repository", trace.WithAttributes(attribute.String("repository", "Create{{.PascalModuleName}}")))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				// Execute query
				if err = r.db.Create(&{{.CamelModuleName}}).Error; err != nil {
					return err
				}

				return nil
			}); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			// Logging
			r.logger.Error(err.Error())
			sentry.CaptureException(err)
			exception.SqlErrorMessage = err.Error()
			err = exception.ErrDbQueryStatement
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return err
	}

	return nil
}

func (r Repository) Update{{.PascalModuleName}}(ctx context.Context, id int, {{.CamelModuleName}} *models.{{.PascalModuleName}}) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "Update{{.PascalModuleName}}Repository", trace.WithAttributes(attribute.String("repository", "Update{{.PascalModuleName}}"), attribute.Int64("id", int64(id))))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				// Execute query
				if err = r.db.
					Model(&models.{{.PascalModuleName}}{}).
					Where("id = ?", id).
					Updates(
						map[string]interface{}{
							// "name": {{.CamelModuleName}}.Name,
						},
					).Error; err != nil {
					return err
				}

				return nil
			}); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else {
				err = e.(error)
				// Logging
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return err
	}

	r.tracer.TraceEnd(childSpan)

	return nil
}

func (r Repository) Delete{{.PascalModuleName}}(ctx context.Context, id int) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "Delete{{.PascalModuleName}}Repository", trace.WithAttributes(attribute.String("repository", "Delete{{.PascalModuleName}}"), attribute.Int64("id", int64(id))))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				// Execute query
				if err = r.db.Delete(&models.{{.PascalModuleName}}{}, id).Error; err != nil {
					return err
				}

				return nil
			}); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else {
				err = e.(error)
				// Logging
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return err
	}

	return nil
}
