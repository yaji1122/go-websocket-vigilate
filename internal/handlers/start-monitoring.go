package handlers

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

type job struct {
	HostServiceID int
}

// Run runs the scheduled job
func (j job) Run() {
	Repo.ScheduledCheck(j.HostServiceID)
}

func (repo *DBRepo) StartMonitor() {
	if app.PreferenceMap["monitoring_live"] == "1" {
		log.Println("************** Starting monitor")
		data := make(map[string]string)
		data["message"] = "Monitoring is starting..."
		// trigger a message to broadcast to all clients that app is starting to monitor.
		if err := app.WsClient.Trigger("public-channel", "app-starting", data); err != nil {
			log.Println(err)
		}

		// get all the services that we want to monitor
		hostServices, err := Repo.DB.GetServicesToMonitor()
		log.Printf("** Monitoring %d services.", len(hostServices))
		if err != nil {
			log.Println(err)
		}
		//range through the services
		for _, hs := range hostServices {
			// get the schedule unit and number
			var sch string
			if hs.ScheduleUnit == "d" {
				sch = fmt.Sprintf("@every %d%s", hs.ScheduleNumber*24, "h")
			} else {
				sch = fmt.Sprintf("@every %d%s", hs.ScheduleNumber, hs.ScheduleUnit)
			}

			// create a job
			var j = job{hs.ID}
			entryID, err := app.Scheduler.AddJob(sch, j)
			if err != nil {
				log.Println(err)
			}
			// save the id of the job, so we can start/stop it
			app.MonitorMap[hs.ID] = entryID

			// broadcast over websockets. The fact that the service is scheduled.
			payload := make(map[string]string)
			payload["message"] = "scheduling"
			payload["host_service_id"] = strconv.Itoa(hs.ID)
			yearOne := time.Date(0001, 11, 17, 20, 34, 56, 78000000, time.UTC)
			//get next schedule time
			if app.Scheduler.Entry(app.MonitorMap[hs.ID]).Next.After(yearOne) {
				payload["next_run"] = app.Scheduler.Entry(app.MonitorMap[hs.ID]).Next.Format("2006-01-02 3:04:05 PM")
			} else {
				payload["next_run"] = "Pending...."
			}
			payload["host"] = hs.Host.HostName
			payload["service"] = hs.Service.ServiceName
			// host service last checked time
			if hs.LastCheck.After(yearOne) {
				payload["last_run"] = hs.LastCheck.Format("2006-01-02 03:04:05")
			} else {
				payload["last_run"] = "Pending..."
			}
			payload["schedule"] = fmt.Sprintf("@every %d%s", hs.ScheduleNumber, hs.ScheduleUnit)

			if err = app.WsClient.Trigger("public-channel", "next-run-event", payload); err != nil {
				log.Println(err)
			}

			if err = app.WsClient.Trigger("public-channel", "schedule-changed-event", payload); err != nil {
				log.Println(err)
			}

			// end scheduler
		}
	}
}
