package handlers

import (
	"github.com/CloudyKit/jet/v6"
	"github.com/tsawler/vigilate/internal/helpers"
	"log"
	"net/http"
)

// AllHealthyServices lists all healthy services
func (repo *DBRepo) AllHealthyServices(w http.ResponseWriter, r *http.Request) {
	vars := make(jet.VarMap)
	hostServices, err := repo.DB.GetHostServicesByStatus("healthy")
	if err != nil {
		log.Println(err)
	}
	hostServiceMessage := make(map[int]string)
	for _, hs := range hostServices {
		message := ""
		lastEvent, err := repo.DB.GetLastEventByHostServiceId(hs.ID)
		if err == nil {
			message = lastEvent.Message
		}
		hostServiceMessage[hs.ID] = message
	}
	vars.Set("hostServiceMessage", hostServiceMessage)
	vars.Set("hostServices", hostServices)

	err = helpers.RenderPage(w, r, "healthy", vars, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}

// AllWarningServices lists all warning services
func (repo *DBRepo) AllWarningServices(w http.ResponseWriter, r *http.Request) {
	vars := make(jet.VarMap)
	hostServices, err := repo.DB.GetHostServicesByStatus("warning")
	if err != nil {
		log.Println(err)
	}
	hostServiceMessage := make(map[int]string)
	for _, hs := range hostServices {
		message := ""
		lastEvent, err := repo.DB.GetLastEventByHostServiceId(hs.ID)
		if err == nil {
			message = lastEvent.Message
		}
		hostServiceMessage[hs.ID] = message
	}
	vars.Set("hostServiceMessage", hostServiceMessage)
	vars.Set("hostServices", hostServices)
	err = helpers.RenderPage(w, r, "warning", vars, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}

// AllProblemServices lists all problem services
func (repo *DBRepo) AllProblemServices(w http.ResponseWriter, r *http.Request) {
	vars := make(jet.VarMap)
	hostServices, err := repo.DB.GetHostServicesByStatus("problem")
	if err != nil {
		log.Println(err)
	}
	hostServiceMessage := make(map[int]string)
	for _, hs := range hostServices {
		message := ""
		lastEvent, err := repo.DB.GetLastEventByHostServiceId(hs.ID)
		if err == nil {
			message = lastEvent.Message
		}
		hostServiceMessage[hs.ID] = message
	}
	vars.Set("hostServiceMessage", hostServiceMessage)
	vars.Set("hostServices", hostServices)
	err = helpers.RenderPage(w, r, "problems", vars, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}

// AllPendingServices lists all pending services
func (repo *DBRepo) AllPendingServices(w http.ResponseWriter, r *http.Request) {
	vars := make(jet.VarMap)
	hostServices, err := repo.DB.GetHostServicesByStatus("pending")
	if err != nil {
		log.Println(err)
	}
	hostServiceMessage := make(map[int]string)
	for _, hs := range hostServices {
		message := ""
		lastEvent, err := repo.DB.GetLastEventByHostServiceId(hs.ID)
		if err == nil {
			message = lastEvent.Message
		}
		hostServiceMessage[hs.ID] = message
	}
	vars.Set("hostServiceMessage", hostServiceMessage)
	vars.Set("hostServices", hostServices)
	err = helpers.RenderPage(w, r, "pending", vars, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}
