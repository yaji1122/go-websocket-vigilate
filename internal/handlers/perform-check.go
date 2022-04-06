package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/tsawler/vigilate/internal/models"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	HTTP = 1
	HTTPS = 2
	SSLCertificate = 3
)

const (
	HOST_SERVICE_STATUS_CHANGE = "HOST_SERVICE_STATUS_CHANGE"
	SCHEDULE_CHANGE = "SCHEDULE_CHANGE"
	SCHEDULE_ITEM_REMOVE = "SCHEDULE_ITEM_REMOVE"
)

type jsonResp struct {
	OK bool `json:"ok"`
	Message string `json:"message"`
	ServiceID int `json:"service_id"`
	HostServiceID int `json:"host_service_id"`
	HostID int `json:"host_id"`
	OldStatus string `json:"old_status"`
	NewStatus string `json:"new_status"`
	LastCheck time.Time `json:"last_check"`
}
// ScheduledCheck performs a scheduled check on a host service by id
func (repo *DBRepo) ScheduledCheck(hostServiceID int) {
	log.Println("******** Running check for", hostServiceID)
	hs, err := repo.DB.GetHostServicesById(hostServiceID)
	oldStatus := hs.Status
	if err != nil {
		log.Println(err)
		log.Println("******** Can't find host service :", hostServiceID)
	}
	h, _ := repo.DB.GetHostById(hs.HostID)
	msg, newStatus := repo.testServiceForHost(h, hs)
	log.Printf("Host servcie %d check, old status: %s, new status: %s, msg: %s", hostServiceID, oldStatus, newStatus, msg)

	if newStatus != oldStatus {
		repo.updateHostServiceStatusCount(hs, newStatus, msg)
	}
}

// CheckNow manually test a host service and sends JSON response
func (repo *DBRepo) CheckNow(w http.ResponseWriter, r *http.Request) {
	hostServiceID, _ := strconv.Atoi(chi.URLParam(r, "id"))
	oldStatus := chi.URLParam(r, "oldStatus")
	ok := true
	// get host service
	hs, err := repo.DB.GetHostServicesById(hostServiceID)

	if err != nil {
		log.Println(err)
		ok = false
	}

	// get host ?
	h, err := repo.DB.GetHostById(hs.HostID)
	if err != nil {
		log.Println(err)
		ok = false
	}
	// test the service
	msg, newStatus := repo.testServiceForHost(h, hs)

	log.Printf("change status from %s to %s, msg:%s", oldStatus, newStatus, msg)

	// send json to client
	var resp jsonResp
	if ok {
		now := time.Now()
		resp.OK = ok
		resp.Message = msg
		resp.ServiceID = hs.ServiceID
		resp.HostID = hs.HostID
		resp.OldStatus = oldStatus
		resp.NewStatus = newStatus
		resp.LastCheck = now
		//broadcast service change event


		//update hostService
		hs.UpdatedAt = now
		hs.LastCheck = now
		hs.Status = newStatus
		repo.updateHostServiceStatusCount(hs, newStatus, msg)
	} else {
		resp.OK = false
	}

	out, _ := json.MarshalIndent(resp, "", "  ")
	w.Header().Set("Content-type", "application/json")
	w.Write(out)
}

func (repo *DBRepo) testServiceForHost(h models.Host, hs models.HostService) (string, string){
	var msg, newStatus string
	switch hs.ServiceID {
	case HTTP:
		msg, newStatus = testHTTPForHost(h.Url)
	}

	// broadcast to clients
	if hs.Status != newStatus {
		repo.pushStatusChangeEvent(h ,hs, newStatus)
	}

	if hs.Active {
		repo.pushScheduleChangedEvent(hs, newStatus)
	} else {
		repo.pushScheduleChangedEvent(hs, "pending")
	}

	// send email/sms

	return msg, newStatus
}
// pushStatusChangeEvent broadcast if host service status change
func (repo *DBRepo) pushStatusChangeEvent(h models.Host, hs models.HostService, newStatus string) {
	data := make(map[string]string)
	data["host_id"] = strconv.Itoa(hs.HostID)
	data["host_service_id"] = strconv.Itoa(hs.ID)
	data["host_name"] = h.HostName
	data["service_name"] = hs.Service.ServiceName
	data["icon"] = hs.Service.Icon
	data["status"] = newStatus
	data["message"] = fmt.Sprintf("%s on %s reports %s", hs.Service.ServiceName, h.HostName, newStatus)
	data["last_check"] = time.Now().Format("2006-01-02 3:04:06 PM")

	broadcastMessage("public-channel", HOST_SERVICE_STATUS_CHANGE, data)
}

// pushScheduleChangedEvent broadcast if schedule changed
func (repo *DBRepo) pushScheduleChangedEvent(hs models.HostService, newStatus string) {
	yearOne := time.Date(0001, 1, 1, 0, 0, 0, 1, time.UTC)
	data := make(map[string]string)
	data["host_service_id"] = strconv.Itoa(hs.ID)
	data["service_id"] = strconv.Itoa(hs.ServiceID)
	data["host_id"] = strconv.Itoa(hs.HostID)

	if app.Scheduler.Entry(repo.App.MonitorMap[hs.ID]).Next.After(yearOne) {
		data["next_run"] = repo.App.Scheduler.Entry(repo.App.MonitorMap[hs.ID]).Next.Format("2006-01-02 3:04:05 PM")
	} else {
		data["next_run"] = "Pending..."
	}
	data["last_run"] = time.Now().Format("2006-01-02 3:04:05 PM")
	data["host"] = hs.Host.HostName
	data["service"] = hs.Service.ServiceName
	data["schedule"] = fmt.Sprintf("@every %d%s", hs.ScheduleNumber, hs.ScheduleUnit)
	data["status"] = newStatus
	data["icon"] = hs.Service.Icon

	broadcastMessage("public-channel", SCHEDULE_CHANGE, data)
}

func testHTTPForHost(url string) (string, string) {
	if strings.HasSuffix(url, "/") {
		url = strings.TrimSuffix(url, "/")
	}

	url = strings.Replace(url, "https://", "http://", -1)

	resp, err := http.Get(url)

	if err != nil {
		return fmt.Sprintf("%s - %s", url, "error connecting"), "problem"
	}
	//Close 必須放在此，代表連線失敗
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("%s - %s", url, resp.Status), "problem"
	}

	return fmt.Sprintf("%s - %s", url, resp.Status), "healthy"
}


//updateHostServiceStatusCount Update Dashboard status count
func (repo *DBRepo) updateHostServiceStatusCount(hs models.HostService, newStatus, msg string) {
	//update hostService
	hs.Status = newStatus
	hs.UpdatedAt = time.Now()
	hs.LastCheck = time.Now()
	if err := repo.DB.UpdateHostServices(hs); err != nil {
		log.Println(err)
	}

	healthy, warning, problem, pending, err := repo.DB.GetAllServiceStatusCount()
	//broadcast changed message
	payload := make(map[string]string)
	//update dashboard counts
	payload["message"] = msg
	payload["healthy_count"] = strconv.Itoa(healthy)
	payload["warning_count"] = strconv.Itoa(warning)
	payload["problem_count"] = strconv.Itoa(problem)
	payload["pending_count"] = strconv.Itoa(pending)
	broadcastMessage("public-channel", HOST_SERVICE_STATUS_CHANGE, payload)
	if err != nil {
		log.Println(err)
		return
	}
}

//broadcastMessage Broadcast message to all connected client
func broadcastMessage(channel, messageType string, data map[string]string) {
	if err := app.WsClient.Trigger(channel, messageType, data); err != nil {
		log.Println(err)
	}
}
