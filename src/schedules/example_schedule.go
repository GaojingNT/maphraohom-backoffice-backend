package schedules

import (
	"maphraohom.app/maphraohom-backoffice/app/scheduler"
)

func (s SchedulerTask) ExampleScheduleJob() {
	s.sc.NewJob(s.cronString, func(ctx scheduler.ISchedulerContext) error {
		// Hello world
		ctx.Log("example-schedule-job", "Example schedule job is executed")

		// Return nil to indicate that the job is executed successfully
		return nil
	})
}
