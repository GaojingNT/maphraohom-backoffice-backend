package app

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
)

func (app *ApplicationServer) handle() error {
	// End of initialization
	if !fiber.IsChild() {
		// Print timezone
		zone, _ := time.Now().Zone()
		log.Printf("Timezone: %s", zone)
		// Print break line
		log.Println("--------------------------------------------------")
	}

	// Start servers (HTTP, GRPC, Scheduler)
	go app.HttpServer.StartServer()
	// go app.GrpcServer.StartServer()
	// go schedules.StartScheduler()

	// There are 2 ways to exit from Microservices
	// 1. The SigTerm can be send from outside program such as from k8s
	// 2. Send true to ms.exitChannel
	osQuit := make(chan os.Signal, 1)
	signal.Notify(osQuit, syscall.SIGTERM, syscall.SIGINT)
	exit := false
	for {
		if exit {
			break
		}
		select {
		// Exit from OS
		case <-osQuit:
			exit = true
		case <-app.HttpServer.ExitChannel:
			// Exit from HTTP as well
			exit = true
			// case <-app.GrpcServer.ExitChannel:
			// 	// Exit from GRPC as well
			// 	exit = true
		}
	}

	return nil
}
