package auth_module

import (
	"context"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"maphraohom.app/maphraohom-backoffice/internal/encryption"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/internal/utils"
	"maphraohom.app/maphraohom-backoffice/src/models"
)

// UpdateProfileInput is the self-editable subset of a user — the signature
// is only ever changed through UpdateUserSignature.
type UpdateProfileInput struct {
	FirstName string
	LastName  string
	Email     string
}

// UpdateProfile replaces the user's first name, last name and email.
// Returns exception.ErrEmailAlreadyTaken if another user (soft-deleted
// ones included — tbl_users.email's unique index covers them too) already
// has that email.
func (r Repository) UpdateProfile(ctx context.Context, id int, input UpdateProfileInput) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "UpdateProfileRepository", trace.WithAttributes(attribute.String("repository", "UpdateProfile"), attribute.Int64("id", int64(id))))
		taken        int64
		result       *gorm.DB
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Unscoped().
				Model(&models.User{}).
				Where("LOWER(email) = LOWER(?) AND id <> ?", input.Email, id).
				Count(&taken).Error; err != nil {
				utils.Throw(err)
			}
			if taken > 0 {
				err = exception.ErrEmailAlreadyTaken
				utils.Throw(err)
			}

			result = r.db.
				Model(&models.User{}).
				Where("id = ?", id).
				Updates(map[string]interface{}{
					"first_name": input.FirstName,
					"last_name":  input.LastName,
					"email":      input.Email,
				})
			if err = result.Error; err != nil {
				utils.Throw(err)
			}
			if result.RowsAffected == 0 {
				err = exception.ErrRecordNotFound
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			if err == exception.ErrEmailAlreadyTaken || err == exception.ErrRecordNotFound {
				return
			}
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

func (r Repository) GetUserSignatureKey(ctx context.Context, id int) (string, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetUserSignatureKeyRepository", trace.WithAttributes(attribute.String("repository", "GetUserSignatureKey"), attribute.Int64("id", int64(id))))
		user         models.User
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Select("id", "signature").First(&user, id).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
				return
			}
			r.logger.Error(err.Error())
			sentry.CaptureException(err)
			exception.SqlErrorMessage = err.Error()
			err = exception.ErrDbQueryStatement
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return "", err
	}

	return user.Signature, nil
}

func (r Repository) UpdateUserSignature(ctx context.Context, id int, key string) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "UpdateUserSignatureRepository", trace.WithAttributes(attribute.String("repository", "UpdateUserSignature"), attribute.Int64("id", int64(id))))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Model(&models.User{}).Where("id = ?", id).Update("signature", key).Error; err != nil {
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

// ChangePassword verifies currentPassword against the stored hash, then
// replaces it with newPassword's hash. Returns
// exception.ErrCurrentPasswordIncorrect when the current password is wrong.
func (r Repository) ChangePassword(ctx context.Context, id int, currentPassword string, newPassword string) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "ChangePasswordRepository", trace.WithAttributes(attribute.String("repository", "ChangePassword"), attribute.Int64("id", int64(id))))
		user         models.User
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Select("id", "password").First(&user, id).Error; err != nil {
				utils.Throw(err)
			}

			if !encryption.VerifyPassword(currentPassword, user.Password) {
				err = exception.ErrCurrentPasswordIncorrect
				utils.Throw(err)
			}

			if err = r.db.Model(&models.User{}).Where("id = ?", id).Update("password", encryption.EncryptPassword(newPassword, "")).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			switch err {
			case exception.ErrCurrentPasswordIncorrect:
				return
			case gorm.ErrRecordNotFound:
				err = exception.ErrRecordNotFound
				return
			}
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
