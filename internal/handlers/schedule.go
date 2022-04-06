package handlers

import (
	"fmt"
	"github.com/CloudyKit/jet/v6"
	"github.com/tsawler/vigilate/internal/helpers"
	"github.com/tsawler/vigilate/internal/models"
	"log"
	"net/http"
	"sort"
)

type SortByHost []models.Schedule
// Len is used to sort by host
func (a SortByHost) Len() int {return len(a)}
// Less is used to sort by host
func (a SortByHost) Less(i, j int) bool {return a[i].Host < a[j].Host}
// Swap is used to sort by host
func (a SortByHost) Swap(i, j int) {a[i], a[j] = a[j], a[i]}

// ListEntries lists schedule entries
func (repo *DBRepo) ListEntries(w http.ResponseWriter, r *http.Request) {
	var items []models.Schedule

	for k, v := range repo.App.MonitorMap {
		var item models.Schedule

		item.ID = k
		item.EntryID = v
		item.Entry = repo.App.Scheduler.Entry(v)
		hs, err := repo.DB.GetHostServicesById(k)
		if err != nil {
			log.Println(err)
			return
		}
		item.ScheduleText = fmt.Sprintf("@every %d%s", hs.ScheduleNumber, hs.ScheduleUnit)
		item.Service = hs.Service.ServiceName
		item.Host = hs.Host.HostName
		item.LastRun = hs.LastCheck
		items = append(items, item)
	}

	//sort the slice
	sort.Sort(SortByHost(items))

	data := make(jet.VarMap)
	data.Set("items", items)

	err := helpers.RenderPage(w, r, "schedule", data, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}
