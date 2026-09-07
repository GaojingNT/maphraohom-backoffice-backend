package schedules

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"maphraohom.app/maphraohom-backoffice/app/scheduler"
	"maphraohom.app/maphraohom-backoffice/internal/helpers/color"
)

type (
	SchedulerTask struct {
		sc         scheduler.IScheduler
		cronString string
	}
)

func NewSchedulerTask(
	cronString string,
) *SchedulerTask {
	return &SchedulerTask{
		sc:         scheduler.CurrentScheduler(),
		cronString: cronString,
	}
}

// StartScheduler initializes the global scheduler and registers all scheduled tasks.
// ! This function should be called in the main application to register all scheduled tasks.
func StartScheduler() {
	if scheduler.GlobalScheduler == nil {
		return
	}

	// Register all jobs
	RegisterJobs()

	// Print schedule task count
	if !fiber.IsChild() {
		log.Println("ScheduleTask:", color.Format(color.GREEN, fmt.Sprint(scheduler.GlobalScheduler.GetTaskCount())), color.Format(color.GREEN, "jobs"))
		// Print break line
		log.Println("--------------------------------------------------")
	}

	// Start scheduler service to run all jobs in the background (non-blocking)
	go scheduler.GlobalScheduler.Start()
}

// RegisterJobs registers all the jobs to the scheduler. (Cron format: minute, hour, day, month, day of week)
func RegisterJobs() {
	// Add scheduler job here... (minute, hour, day, month, day of week)
	NewSchedulerTask("* * * * *").ExampleScheduleJob()
}
