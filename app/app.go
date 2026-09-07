package app

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v2"
	"maphraohom.app/maphraohom-backoffice/app/grpc_server"
	"maphraohom.app/maphraohom-backoffice/app/http_server"
	"maphraohom.app/maphraohom-backoffice/app/scheduler"
	"maphraohom.app/maphraohom-backoffice/config"
	"maphraohom.app/maphraohom-backoffice/internal/jwt"
	"maphraohom.app/maphraohom-backoffice/pkg/cache"
	"maphraohom.app/maphraohom-backoffice/pkg/database"
	"maphraohom.app/maphraohom-backoffice/pkg/logger"
	sentryException "maphraohom.app/maphraohom-backoffice/pkg/sentry"
	"maphraohom.app/maphraohom-backoffice/pkg/storage"
	"maphraohom.app/maphraohom-backoffice/pkg/tracing"
	"maphraohom.app/maphraohom-backoffice/src/routes"
)

type ApplicationServer struct {
	HttpServer *http_server.HttpServer
	GrpcServer *grpc_server.GrpcServer
}

type IApplicationServer interface {
	StartServer() error
	StopServer() error

	initialize()
	registerServerRouter()
	handle() error
	cleanup()
}

func NewApplicationServer() *ApplicationServer {
	return &ApplicationServer{}
}

func (app *ApplicationServer) StartServer() error {
	// Initialize the application components
	app.initialize()

	// Create microservice instances
	app.HttpServer = http_server.NewHttpServer(config.Global, database.DbClient, cache.AppCacher, tracing.AppTracer, storage.FileStorage)
	// app.GrpcServer = grpc_server.NewGrpcServer(config.Global, database.DbClient, cache.AppCacher, tracing.AppTracer, storage.FileStorage)

	// Initialize global scheduler
	// scheduler.GlobalScheduler = scheduler.NewScheduler(config.Global, database.DbClient, cache.AppCacher, tracing.AppTracer, storage.FileStorage)

	// Register routes
	app.registerServerRouter()

	return app.handle()
}

func (app *ApplicationServer) StopServer() error {
	// Ignore fiber child process
	if fiber.IsChild() {
		return nil
	}

	var timeMs = time.Now().UnixMilli()

	// Stop servers
	app.HttpServer.StopServer()
	// app.GrpcServer.StopServer()

	// Shutdown global scheduler
	// scheduler.GlobalScheduler.Shutdown()

	// Cleanup
	app.cleanup()

	// Print time taken to shutdown
	lastText := fmt.Sprintf("Time taken to shutdown successfully: %s ms", strconv.FormatInt(time.Now().UnixMilli()-timeMs, 10))
	log.Println(lastText)
	logger.CurrentLogger().Info(lastText)

	return nil
}

func (app *ApplicationServer) initialize() {
	// Load environment variables
	config.Global = config.NewConfig()

	// Initialize JWT
	jwt.JwtToken = jwt.NewJWT()

	// Initialize Postgresql connection
	database.DbClient = database.Initialize()

	// Initialize logger
	logger.AppLogger = logger.Initialize()

	// Initialize connection to cache
	cache.AppCacher = cache.NewCacher(
		cache.Initialize(),
		cache.WithPrefix(config.Global.Redis.RedisCachePrefix),
		cache.WithExpired(time.Minute*time.Duration(config.Global.Redis.RedisCacheDuration)),
	)

	// Initialize OpenTelemetry tracing
	tracing.AppTracer = tracing.InitTracer()

	// Initialize Sentry client for error logging and tracing
	sentryException.Initialize()

	// Initialize storage file system instance
	storage.FileStorage = storage.NewFileSystem(
		config.Global.FileSystem.Disk,
		storage.WithS3Endpoint(config.Global.FileSystem.S3Config.Endpoint),
		storage.WithS3Bucket(config.Global.FileSystem.S3Config.Bucket),
		storage.WithS3Region(config.Global.FileSystem.S3Config.Region),
		storage.WithS3AccessKeyID(config.Global.FileSystem.S3Config.AccessKeyID),
		storage.WithS3SecretKey(config.Global.FileSystem.S3Config.SecretKey),
		storage.WithS3UseSSL(config.Global.FileSystem.S3Config.UseSSL),
		storage.WithS3Prefix(config.Global.FileSystem.S3Config.Prefix),
	)
}

func (app *ApplicationServer) registerServerRouter() {
	// Setup routes
	routes.HTTPRoutes(app.HttpServer)
	// routes.GRPCRegisterServices(app.GrpcServer)
}

func (app *ApplicationServer) cleanup() {
	// Close all database connection
	if database.CurrentDatabase() != nil {
		sqlDB, _ := database.CurrentDatabase().DB()
		sqlDB.Close()
	}

	// Close all redis connection
	if cache.CurrentCacher() != nil {
		cache.CurrentCacher().Close()
	}

	// Sentry: Flush buffered events before the program terminates.
	if config.Global.Sentry.SentryDSN != "" {
		sentry.Flush(2 * time.Second)
		sentry.Recover()
	}

	// Close all tracer connection
	if tracing.CurrentTracer().TraceProvider != nil {
		tracing.CurrentTracer().Cleanup()
	}

	// Close logger
	if logger.CurrentLogger() != nil {
		logger.CurrentLogger().Sync()

		// Check if enable write log file, closing file I/O
		if config.Global.Logger.LogMode == "s3" || config.Global.Logger.LogMode == "file" {
			logger.Close()
		}
	}

	// Stop all schedule jobs
	if scheduler.CurrentScheduler() != nil {
		scheduler.CurrentScheduler().Shutdown()
	}
}
