package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/tsawler/vigilate/internal/certificateutils"
	"github.com/tsawler/vigilate/internal/channeldata"
	"github.com/tsawler/vigilate/internal/helpers"
	"github.com/tsawler/vigilate/internal/models"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	HTTP           = 1
	HTTPS          = 2
	SSLCertificate = 3
)

const (
	HOST_SERVICE_STATUS_CHANGE = "HOST_SERVICE_STATUS_CHANGE"
	SCHEDULE_CHANGE            = "SCHEDULE_CHANGE"
	SCHEDULE_ITEM_REMOVE       = "SCHEDULE_ITEM_REMOVE"
)

type jsonResp struct {
	OK            bool      `json:"ok"`
	Message       string    `json:"message"`
	ServiceID     int       `json:"service_id"`
	HostServiceID int       `json:"host_service_id"`
	HostID        int       `json:"host_id"`
	OldStatus     string    `json:"old_status"`
	NewStatus     string    `json:"new_status"`
	LastCheck     time.Time `json:"last_check"`
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
	} else {
		resp.OK = false
	}

	out, _ := json.MarshalIndent(resp, "", "  ")
	w.Header().Set("Content-type", "application/json")
	w.Write(out)
}

//testServiceForHost check host service
func (repo *DBRepo) testServiceForHost(h models.Host, hs models.HostService) (string, string) {
	var msg, newStatus string
	switch hs.ServiceID {
	case HTTP:
		msg, newStatus = testHTTPForHost(h.Url)
	case HTTPS:
		msg, newStatus = testHTTPSForHost(h.Url)
	case SSLCertificate:
		msg, newStatus = testSLLCertificateForHost(h.Url)
	}
	//update hostService
	hs.UpdatedAt = time.Now()
	hs.LastCheck = time.Now()
	// broadcast to clients
	if hs.Status != newStatus {
		log.Println("Insert event")
		event := models.Event{
			EventType:     newStatus,
			HostServiceID: hs.ID,
			HostID:        h.ID,
			HostName:      h.HostName,
			ServiceName:   hs.Service.ServiceName,
			Message:       msg,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		err := repo.DB.InsertEvent(event)
		if err != nil {
			log.Println(err)
		}

		//send email if appropriate
		if repo.App.PreferenceMap["notify_via_email"] == "1" {
			if hs.Status != "pending" {
				mailMessage := channeldata.MailData{
					ToName:    repo.App.PreferenceMap["notify_name"],
					ToAddress: repo.App.PreferenceMap["notify_email"],
					Subject:   fmt.Sprintf("%s: service %s on %s", strings.ToUpper(newStatus), hs.Service.ServiceName, h.HostName),
					Content: template.HTML(fmt.Sprintf(`
					<p>Service %s on %s reported %s status</p>
					<p><strong>Message Received: %s</strong></p>		
					`, hs.Service.ServiceName, h.HostName, strings.ToUpper(newStatus), msg)),
				}
				helpers.SendEmail(mailMessage)
			}
		}

		//send sms if appropriate

		hs.Status = newStatus
		//通知狀態變更
		repo.pushStatusChangeEvent(h, hs, newStatus)
		//更新Dashboard
		repo.updateHostServiceStatusCount(hs, newStatus, msg)
		//狀態變更事件
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

//testHTTPForHost test http
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

//testHTTPSForHost test https
func testHTTPSForHost(url string) (string, string) {
	if strings.HasSuffix(url, "/") {
		url = strings.TrimSuffix(url, "/")
	}

	url = strings.Replace(url, "http://", "https://", -1)

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

//testSLLCertificateForHost test https
func testSLLCertificateForHost(url string) (string, string) {
	if strings.HasPrefix(url, "https://") {
		url = strings.Replace(url, "https://", "", -1)
	}

	if strings.HasPrefix(url, "http://") {
		url = strings.Replace(url, "http://", "", -1)
	}

	url = strings.Replace(url, "http://", "https://", -1)

	certDetailsChannel := make(chan certificateutils.CertificateDetails, 1)
	errorsChannel := make(chan error, 1)

	var msg, newStatus string
	scanHost(url, certDetailsChannel, errorsChannel)

	for i := 0; i < len(certDetailsChannel); i++ {
		certDetails := <-certDetailsChannel
		certificateutils.CheckExpirationStatus(&certDetails, 30)
		msg = certDetails.Hostname + " expiring in " + strconv.Itoa(certDetails.DaysUntilExpiration) + " days"
		if certDetails.ExpiringSoon {
			if certDetails.DaysUntilExpiration < 7 {
				newStatus = "problem"
			} else {
				newStatus = "warning"
			}
		} else {
			newStatus = "healthy"
		}
	}

	if len(errorsChannel) > 0 {
		log.Printf("There were %d error(s):\n", len(errorsChannel))
		for i, errorsInChannel := 0, len(errorsChannel); i < errorsInChannel; i++ {
			log.Printf("%s\n", <-errorsChannel)
		}
		log.Printf("\n")
	}

	return msg, newStatus
}

func scanHost(url string, certDetailChannel chan certificateutils.CertificateDetails, errorsChannel chan error) {
	res, err := certificateutils.GetCertificateDetails(url, 10)
	if err != nil {
		errorsChannel <- err
	} else {
		certDetailChannel <- res
	}
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
