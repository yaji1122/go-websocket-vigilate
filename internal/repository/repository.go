package repository

import (
	"github.com/tsawler/vigilate/internal/models"
)

// DatabaseRepo is the database repository
type DatabaseRepo interface {
	// preferences
	AllPreferences() ([]models.Preference, error)
	SetSystemPref(name, value string) error
	InsertOrUpdateSitePreferences(pm map[string]string) error

	// users and authentication
	GetUserById(id int) (models.User, error)
	InsertUser(u models.User) (int, error)
	UpdateUser(u models.User) error
	DeleteUser(id int) error
	UpdatePassword(id int, newPassword string) error
	Authenticate(email, testPassword string) (int, string, error)
	AllUsers() ([]*models.User, error)
	InsertRememberMeToken(id int, token string) error
	DeleteToken(token string) error
	CheckForToken(id int, token string) bool

	// host
	AllHosts() ([]*models.Host, error)
	GetHostById(id int) (models.Host, error)
	InsertHost(h models.Host) (int, error)
	UpdateHost(h models.Host) error

	// hostServices
	GetServicesToMonitor() ([]models.HostService, error)
	GetAllServiceStatusCount() (int, int, int, int, error)
	GetServiceStatusCount(queryStatus string) (int, error)
	GetHostServicesByStatus(status string) ([]models.HostService, error)
	GetHostServicesById(id int) (models.HostService, error)
	GetHostServicesByHostIDAndServiceID(hostID int, ServiceID int) (models.HostService, error)
	GetHostServicesByHostId(id int) ([]models.HostService, error)
	UpdateHostServices(hs models.HostService) error

	//events
	GetAllEvents() ([]*models.Event, error)
	InsertEvent(e models.Event) error
}
