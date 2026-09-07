package grpc_server

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	"gorm.io/gorm"
	"maphraohom.app/maphraohom-backoffice/config"
	"maphraohom.app/maphraohom-backoffice/pkg/cache"
	"maphraohom.app/maphraohom-backoffice/pkg/logger"
	"maphraohom.app/maphraohom-backoffice/pkg/storage"
	"maphraohom.app/maphraohom-backoffice/pkg/tracing"
)

type (
	IGrpcServer interface {
		StartServer() error
		StopServer()
		Log() *logger.Logger

		startGRPC(exitChannel chan bool) error
	}
	// GrpcServer implement IGrpcServer it is context for GRPC service
	GrpcServer struct {
		Config      *config.Config
		Grpc        *grpc.Server
		GrpcOpts    GrpcServerOption
		DbClient    *gorm.DB
		Cacher      *cache.Cache
		Tracer      *tracing.MyTracer
		FileSystem  *storage.FileSystem
		ExitChannel chan bool
	}

	GrpcServerOptionFunc func(*GrpcServerOption)

	GrpcServerOption struct {
		Opts []grpc.ServerOption
	}
)

func defaultGrpcConfig() GrpcServerOption {
	return GrpcServerOption{
		Opts: make([]grpc.ServerOption, 0),
	}
}

// NewGrpcServer is the constructor function for GrpcServer
func NewGrpcServer(
	config *config.Config,
	dbClient *gorm.DB,
	cacher *cache.Cache,
	tracer *tracing.MyTracer,
	fileSystem *storage.FileSystem,
	opts ...GrpcServerOptionFunc,
) *GrpcServer {
	grpcOptions := defaultGrpcConfig()

	for _, fn := range opts {
		fn(&grpcOptions)
	}

	return &GrpcServer{
		Config:     config,
		Grpc:       grpc.NewServer(grpcOptions.Opts...),
		DbClient:   dbClient,
		Cacher:     cacher,
		Tracer:     tracer,
		FileSystem: fileSystem,
	}
}

// Start start all registered services
func (s *GrpcServer) StartServer() error {
	exitGRPC := make(chan bool, 1)
	go func() {
		s.startGRPC(exitGRPC)
	}()

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
			// Exit from GRPC as well
			if exitGRPC != nil {
				exitGRPC <- true
			}
			exit = true
		case <-s.ExitChannel:
			// Exit from GRPC as well
			if exitGRPC != nil {
				exitGRPC <- true
			}
			exit = true
		}
	}

	return nil
}

// Log message to console
func (s *GrpcServer) Log() *logger.Logger {
	return logger.CurrentLogger()
}

// startGRPC will start GRPC service, this function will block thread
func (s *GrpcServer) startGRPC(exitChannel chan bool) error {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", s.Config.App.GrpcPort))
	if err != nil {
		return err
	}

	if !fiber.IsChild() {
		log.Printf("[App] GRPC service is running at port %s\n", s.Config.App.GrpcPort)
		s.Log().Info("GRPC service is running at port " + s.Config.App.GrpcPort)
	}

	return s.Grpc.Serve(lis)
}

// stopGRPC will stop GRPC service graceful shutdown
func (s *GrpcServer) StopServer() {
	if !fiber.IsChild() {
		log.Println("[App] GrpcServer", "Shutting down...")
		s.Log().Info("Grpc server shutting down...")
	}

	s.Grpc.GracefulStop()
}
