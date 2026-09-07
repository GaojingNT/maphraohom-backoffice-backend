package http_server

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"maphraohom.app/maphraohom-backoffice/config"
	"maphraohom.app/maphraohom-backoffice/pkg/cache"
	"maphraohom.app/maphraohom-backoffice/pkg/logger"
	"maphraohom.app/maphraohom-backoffice/pkg/storage"
	"maphraohom.app/maphraohom-backoffice/pkg/tracing"
)

type (
	IHttpServer interface {
		StartServer() error
		StopServer() error
		Log(tag string, message string)

		startHTTP(exitChannel chan bool) error
	}
	// HttpServer implement IHttpServer it is context for HTTP service
	HttpServer struct {
		Config      *config.Config
		fiber       *fiber.App
		Router      *Route
		DbClient    *gorm.DB
		Cacher      *cache.Cache
		Tracer      *tracing.MyTracer
		Logger      *logger.Logger
		FileSystem  *storage.FileSystem
		ExitChannel chan bool
	}

	IRoute interface {
		// HTTP Services
		Use(args ...interface{})
		Group(prefix string, h ...func(*fiber.Ctx) error) fiber.Router

		// HTTP Methods
		Get(path string, h ...func(*fiber.Ctx) error) fiber.Router
		Post(path string, h ...func(*fiber.Ctx) error) fiber.Router
		Put(path string, h ...func(*fiber.Ctx) error) fiber.Router
		Patch(path string, h ...func(*fiber.Ctx) error) fiber.Router
		Delete(path string, h ...func(*fiber.Ctx) error) fiber.Router
	}

	Route struct {
		fiber *fiber.App
	}
)

// NewRoute is the constructor function for Route
func NewRoute(fiber *fiber.App) *Route {
	return &Route{
		fiber: fiber,
	}
}

// NewHttpServer is the constructor function for HttpServer
func NewHttpServer(
	config *config.Config,
	dbClient *gorm.DB,
	cacher *cache.Cache,
	tracer *tracing.MyTracer,
	fileSystem *storage.FileSystem,
) *HttpServer {
	fiberApp := fiber.New(config.Fiber.Config)

	return &HttpServer{
		Config:     config,
		fiber:      fiberApp,
		Router:     NewRoute(fiberApp),
		DbClient:   dbClient,
		Cacher:     cacher,
		Tracer:     tracer,
		FileSystem: fileSystem,
	}
}

// Start start all registered services
func (s *HttpServer) StartServer() error {

	httpN := len(s.fiber.Stack())
	var exitHTTP chan bool
	if httpN > 0 {
		exitHTTP = make(chan bool, 1)
		go func() {
			s.startHTTP(exitHTTP)
		}()
	}

	// There are 2 ways to exit from Microservices
	// 1. The SigTerm can be send from outside program such as from k8s
	// 2. Send true to ms.exitChannel
	osQuit := make(chan os.Signal, 1)
	s.ExitChannel = make(chan bool, 1)
	signal.Notify(osQuit, syscall.SIGTERM, syscall.SIGINT)
	exit := false
	for {
		if exit {
			break
		}
		select {
		case <-osQuit:
			// Exit from HTTP as well
			if exitHTTP != nil {
				exitHTTP <- true
			}
			exit = true
		case <-s.ExitChannel:
			// Exit from HTTP as well
			if exitHTTP != nil {
				exitHTTP <- true
			}
			exit = true
		}
	}

	return nil
}

// Log message to console
func (s *HttpServer) Log() *logger.Logger {
	return logger.CurrentLogger()
}

// startHTTP will start HTTP service, this function will block thread
func (s *HttpServer) startHTTP(exitChannel chan bool) error {
	if !fiber.IsChild() {
		log.Printf("[App] HTTP service is running at port %s\n", s.Config.App.HttpPort)
		s.Log().Info("HTTP service is running at port " + s.Config.App.HttpPort)
	}

	return s.fiber.Listen(":" + s.Config.App.HttpPort)
}

// stopHTTP will stop HTTP service graceful shutdown
func (s *HttpServer) StopServer() error {
	if !fiber.IsChild() {
		log.Println("[App] HttpServer", "Shutting down...")
		s.Log().Info("Http server shutting down...")
	}

	if config.Global.App.Env == "production" {
		// In production mode we can use graceful shutdown with timeout
		return s.fiber.ShutdownWithTimeout(15 * time.Second)
	} else {
		// In development mode we can use force shutdown
		return s.fiber.Shutdown()
	}
}

// Call router
func (s *HttpServer) Route() *Route {
	return s.Router
}

func (s *HttpServer) MainRoute() fiber.Router {
	return s.Route().Group("")
}
