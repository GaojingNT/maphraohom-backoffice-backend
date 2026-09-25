package auth_module

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/mail"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/goccy/go-json"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
	"maphraohom.app/maphraohom-backoffice/config"
	"maphraohom.app/maphraohom-backoffice/internal/encryption"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/internal/helpers"
	"maphraohom.app/maphraohom-backoffice/internal/jwt"
	"maphraohom.app/maphraohom-backoffice/internal/utils"
	"maphraohom.app/maphraohom-backoffice/src/models"
)

func (r Repository) Authenticate(ctx context.Context, username string, password string) (string, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "AuthenticateRepository", trace.WithAttributes(attribute.String("repository", "Authenticate")))
		authUser     *models.User
		isAuth       bool
		accessToken  string
		err          error
	)

	// CASE: Authenticate user
	// 1. Get user by username
	// 2. Attempt password
	// 3. Generate access token

	utils.Block{
		Try: func() {
			// Get user by username
			authUser, err = r.GetUserByEmail(ctx, username)
			if err != nil {
				utils.Throw(err)
			}

			log.Println("authUser passed")

			// Attempt password
			if isAuth, err = r.Attempt(ctx, authUser, password); err != nil {
				utils.Throw(err)
			} else if !isAuth {
				err = exception.ErrInvalidLoginCredential
				utils.Throw(err)
			}

			log.Println("Attempt passed")

			// Generate personal access token
			accessToken, err = r.GenerateJWT(
				ctx,
				map[string]interface{}{
					"user": map[string]interface{}{
						"id":    authUser.ID,
						"email": authUser.Email,
						"role": func() map[string]interface{} {
							var role map[string]interface{}

							if authUser.Role != nil {
								role = map[string]interface{}{
									"id":   authUser.RoleID,
									"name": authUser.Role.Name,
								}
							}

							return role
						}(),
					},
				},
				(time.Duration(config.Global.Auth.SessionLifetime) * time.Second),
			)
			if err != nil {
				utils.Throw(err)
			}

			log.Println("GenerateJWT passed")
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			switch {
			// Unknown email and wrong password answer the same, so the
			// sign-in form can't be used to probe which emails exist.
			case errors.Is(err, exception.ErrRecordNotFound), errors.Is(err, gorm.ErrRecordNotFound), errors.Is(err, exception.ErrInvalidLoginCredential):
				err = exception.ErrInvalidLoginCredential
			default:
				// Logging
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func (r Repository) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetUserByIDRepository", trace.WithAttributes(attribute.String("repository", "GetUserByID")))
		user         models.User
		err          error
	)

	utils.Block{
		Try: func() {
			// Get user by email
			if err = r.db.
				Preload("Role.RoleGroups").
				Preload("Role.Permissions").
				Preload("Role.Menus.SubMenus").
				Preload("Stores", func(db *gorm.DB) *gorm.DB { return db.Order("tbl_stores.id") }).
				First(&user, id).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			if err == gorm.ErrRecordNotFound { // CASE: User not found
				err = exception.ErrRecordNotFound
			} else {
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
		return nil, err
	}

	return &user, nil
}

func (r Repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetUserByEmailRepository", trace.WithAttributes(attribute.String("repository", "GetUserByEmail")))
		user         models.User
		err          error
	)

	utils.Block{
		Try: func() {
			// Get user by email
			if err = r.db.
				Preload("Role").
				Where("email = ?", email).
				First(&user).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			if err == gorm.ErrRecordNotFound { // CASE: User not found
				err = exception.ErrRecordNotFound
			} else {
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
		return nil, err
	}

	return &user, nil
}

func (r Repository) Attempt(ctx context.Context, user *models.User, password string) (bool, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "AttemptRepository", trace.WithAttributes(attribute.String("repository", "Attempt")))
		isAuth       bool
		err          error
	)

	utils.Block{
		Try: func() {
			// Compare password
			isAuth = encryption.VerifyPassword(password, user.Password)
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
		return false, err
	}

	return isAuth, nil
}

func (r Repository) ForgotPassword(ctx context.Context, email string) error {
	var (
		_, childSpan       = r.tracer.TraceStart(ctx, "ForgotPasswordRepository", trace.WithAttributes(attribute.String("repository", "ForgotPassword")))
		user               *models.User
		resetPasswordToken string
		err                error
	)

	// CASE: Forgot password
	// 1. Get user by email
	// 2. Generate reset password token
	// 3. Update user reset password token and reset password token expired date time
	// 4. Send reset password email

	utils.Block{
		Try: func() {
			// Get user by email
			user, err = r.GetUserByEmail(ctx, email)
			if err != nil {
				utils.Throw(err)
			}

			// Generate reset password token
			resetPasswordToken, err = helpers.RandomHexString(16)
			if err != nil {
				utils.Throw(err)
			}

			// Update user reset password token
			if err = r.db.
				Model(&user).
				Updates(map[string]interface{}{
					"reset_password_token":      resetPasswordToken,
					"reset_password_expired_at": time.Now().Add(24 * time.Hour),
				}).Error; err != nil {
				utils.Throw(err)
			}

			// Send reset password email
			if err = r.SendResetPasswordEmail(ctx, user, resetPasswordToken); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			if err == gorm.ErrRecordNotFound { // CASE: User not found
				err = exception.ErrRecordNotFound
			}
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

func (r Repository) SendResetPasswordEmail(ctx context.Context, user *models.User, resetPasswordToken string) error {
	var (
		_, childSpan         = r.tracer.TraceStart(ctx, "ForgotPasswordRepository", trace.WithAttributes(attribute.String("repository", "ForgotPassword")))
		mailConn             *gomail.Dialer
		mailMessage          = gomail.NewMessage()
		emailSubject         = "Reset Password"
		emailContent         string
		resetPasswordPayload = map[string]interface{}{"email": user.Email, "resetPasswordToken": resetPasswordToken}
		payloadBytes         []byte
		base64Payload        string
		resetPasswordUri     string
		err                  error
	)

	// CASE: Send reset password email
	// 1. Make email content
	// 2. Set email subject
	// 3. Set email receiver
	// 4. Send email

	utils.Block{
		Try: func() {
			// Make email content
			// 1. Encode payload
			// 2. Body (Content with reset password button)
			// 3. Footer (Signature)

			// Encode payload JSON
			payloadBytes, err = json.Marshal(resetPasswordPayload)
			if err != nil {
				utils.Throw(err)
			}

			// Base64 encode payload
			base64Payload = base64.StdEncoding.EncodeToString(payloadBytes)

			// Make reset password URI
			resetPasswordUri = fmt.Sprintf(
				"%s/reset-password?payload=%s",
				config.Global.App.AppFrontendUrl,
				base64Payload,
			)

			// Body
			emailContent = fmt.Sprintf(
				"<p style=\"font: small/1.5 Arial,Helvetica,sans-serif;\"><b>Hello!,</b></p>"+
					"You are receiving this email because we received a password reset request for your account<br>"+
					"<br>"+
					"<a href=\"%s\"><input type=\"button\" value=\"Reset Password\"/></a><br>"+
					"<br>"+
					"If you're having trouble clicking the \"Reset Password\" button, copy and paste the URL below into your web browser:<br>"+
					"<a href=\"%s\">%s</a><br>"+
					"</p>",
				resetPasswordUri,
				resetPasswordUri,
				resetPasswordUri,
			)

			// Validation email
			_, err = mail.ParseAddress(config.Global.Mail.SmtpUsername)
			if err != nil {
				utils.Throw(err)
			}
			_, err = mail.ParseAddress(user.Email)
			if err != nil {
				utils.Throw(err)
			}

			// Set sender and receiver
			mailMessage.SetHeader("From", config.Global.Mail.SmtpUsername)
			mailMessage.SetHeader("To", user.Email)

			// Set subject
			mailMessage.SetHeader("Subject", emailSubject)

			// Set body
			mailMessage.SetBody("text/html", emailContent)

			// Setup dial
			mailConn = gomail.NewDialer(
				config.Global.Mail.SmtpHost,
				config.Global.Mail.SmtpPort,
				config.Global.Mail.SmtpUsername,
				config.Global.Mail.SmtpPassword,
			)

			// Retry parameters
			maxRetries := 3
			retryDelay := time.Second * 5

			// Loop send mail
			for attempt := 1; attempt <= maxRetries; attempt++ {
				if err = mailConn.DialAndSend(mailMessage); err != nil {
					time.Sleep(retryDelay)
					continue
				}
				break
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

func (r Repository) ResetPassword(ctx context.Context, email string, password string, resetPasswordToken string) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "ResetPasswordRepository", trace.WithAttributes(attribute.String("repository", "ResetPassword")))
		user         *models.User
		err          error
	)

	// CASE: Reset password
	// 1. Get user by email
	// 2. Check reset password token
	// 3. Update user password and reset password token to empty string

	utils.Block{
		Try: func() {
			// Get user by email
			user, err = r.GetUserByEmail(ctx, email)
			if err != nil {
				utils.Throw(err)
			}

			// Check reset password token
			if user.ResetPasswordToken != resetPasswordToken {
				err = exception.ErrInvalidResetPasswordToken
				utils.Throw(err)
			}

			// Update user password and reset password token
			if err = r.db.
				Model(&user).
				Updates(map[string]interface{}{
					"password":                  encryption.EncryptPassword(password, ""),
					"reset_password_token":      "",
					"reset_password_expired_at": sql.NullTime{Time: time.Time{}, Valid: false},
				}).Error; err != nil {
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

func (r Repository) GenerateJWT(ctx context.Context, content interface{}, duration time.Duration) (string, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GenerateJWTRepository", trace.WithAttributes(attribute.String("repository", "GenerateJWT")))
		token        string
		err          error
	)

	utils.Block{
		Try: func() {
			// Generate JWT
			token, err = jwt.JwtToken.Create(
				duration,
				content,
			)
			if err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			// Logging
			r.logger.Error(err.Error())
			exception.SqlErrorMessage = err.Error()
			sentry.CaptureException(e.(error))
		},
		Finally: func() {},
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return "", err
	}

	return token, nil
}
