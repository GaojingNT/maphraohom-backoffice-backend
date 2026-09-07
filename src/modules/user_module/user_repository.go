package user_module

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

func (r Repository) GetUserPaginate(ctx context.Context, pagination *paginator.Pagination) (*paginator.Pagination, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetUserPaginateRepository", trace.WithAttributes(attribute.String("repository", "GetUserPaginate")))
		users        = make([]models.User, 0)
		err          error
	)

	// Get attributes
	searchAttribute, _ := pagination.GetStringAttribute("search")
	searchByAttribute, _ := pagination.GetStringAttribute("search_by")

	// Set tracing attributes
	r.tracer.SetAttributes(childSpan, attribute.String("search", searchAttribute))
	r.tracer.SetAttributes(childSpan, attribute.String("search_by", searchByAttribute))

	// Create pagination query
	tx := r.db

	utils.Block{
		Try: func() {
			// Execute query
			if err = tx.
				Scopes(models.SearchingScope(models.UserSearchable(), searchAttribute, searchByAttribute)).
				Scopes(paginator.Paginate(users, pagination, tx)).
				Find(&users).Error; err != nil {
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
	pagination.Data = users

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return nil, err
	}

	return pagination, nil
}

func (r Repository) GetUserByID(ctx context.Context, id int) (models.User, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetUserByIDRepository", trace.WithAttributes(attribute.String("repository", "GetUserByID"), attribute.Int64("id", int64(id))))
		user         models.User
		err          error
	)

	utils.Block{
		Try: func() {
			// Execute query
			if err = r.db.
				Preload("Role").
				Preload("Role.Permissions").
				First(&user, id).Error; err != nil {
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
		return user, err
	}

	return user, nil
}

func (r Repository) CreateUser(ctx context.Context, user *models.User) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CreateUserRepository", trace.WithAttributes(attribute.String("repository", "CreateUser")))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				// Execute query
				if err = r.db.Create(&user).Error; err != nil {
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

func (r Repository) UpdateUser(ctx context.Context, id int, user *models.User) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "UpdateUserRepository", trace.WithAttributes(attribute.String("repository", "UpdateUser"), attribute.Int64("id", int64(id))))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				// Execute query
				if err = r.db.
					Model(&models.User{}).
					Where("id = ?", id).
					Updates(
						map[string]interface{}{
							"first_name": user.FirstName,
							"last_name":  user.LastName,
							"email":      user.Email,
							"role_id":    user.RoleID,
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

func (r Repository) DeleteUser(ctx context.Context, id int) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "DeleteUserRepository", trace.WithAttributes(attribute.String("repository", "DeleteUser"), attribute.Int64("id", int64(id))))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				// Execute query
				if err = r.db.Delete(&models.User{}, id).Error; err != nil {
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
