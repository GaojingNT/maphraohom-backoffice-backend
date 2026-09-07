package healthcheck_module

import (
	"context"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/internal/utils"
)

type (
	IRepository interface {
		CheckDatabaseConnection(ctx context.Context) error
	}
)

func (r Repository) CheckDatabaseConnection(ctx context.Context) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CheckDatabaseConnectionRepository", trace.WithAttributes(attribute.String("repository", "CheckDatabaseConnection")))
		err          error
	)

	utils.Block{
		Try: func() {
			sqlDB, _ := r.db.DB()
			if err = sqlDB.Ping(); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			// Logging
			r.logger.Error(err.Error())
			sentry.CaptureException(err)
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

// // Note: Example for how to coding upload file to file system
// func (r Repository) TestUploadFile(ctx context.Context, fh *multipart.FileHeader) error {
// 	var (
// 		_, childSpan = r.tracer.TraceStart(ctx, "TestUploadFileRepository", trace.WithAttributes(attribute.String("repository", "TestUploadFile")))
// 		err          error
// 	)

// 	utils.Block{
// 		Try: func() {
// 			if err = r.fileSystem.Put(ctx, "test-folder", "test-name", fh); err != nil {
// 				utils.Throw(err)
// 			}
// 		},
// 		Catch: func(e utils.Exception) {
// 			err = e.(error)
// 			// Logging
// 			r.Logger().Error(err.Error())
// 			sentry.CaptureException(err)
// 		},
// 		Finally: nil,
// 	}.Do()

// 	r.tracer.TraceEnd(childSpan)

// 	// Check error
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
