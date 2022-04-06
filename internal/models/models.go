package models

import (
	"errors"
	"github.com/robfig/cron/v3"
	"time"
)

var (
	// ErrNoRecord no record found in database error
	ErrNoRecord = errors.New("models: no matching record found")
	// ErrInvalidCredentials invalid username/password error
	ErrInvalidCredentials = errors.New("models: invalid credentials")
	// ErrDuplicateEmail duplicate email error
	ErrDuplicateEmail = errors.New("models: duplicate email")
	// ErrInactiveAccount inactive account error
	ErrInactiveAccount = errors.New("models: Inactive Account")
)

// User model
type User struct {
	ID          int
	FirstName   string
	LastName    string
	UserActive  int
	AccessLevel int
	Email       string
	Password    []byte
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   time.Time
	Preferences map[string]string
}

// Preference model
type Preference struct {
	ID         int
	Name       string
	Preference []byte
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
// Host is the model for hosts
type Host struct {
	ID int
	HostName string
	CanonicalName string
	Url string
	IP string
	IPV6 string
	Location string
	OS string
	Active bool
	CreatedAt time.Time
	UpdatedAt time.Time
	HostServices []HostService
}
// Service is the model for services
type Service struct {
	ID int
	ServiceName string
	Active bool
	Icon string
	CreatedAt time.Time
	UpdatedAt time.Time
}
// HostService is the model for hostServices
type HostService struct {
	ID int
	HostID int
	Host Host
	ServiceID int
	Service Service
	Active bool
	ScheduleNumber int
	ScheduleUnit string
	Status string
	LastCheck time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Schedule is schedule
type Schedule struct {
	ID int
	EntryID cron.EntryID
	Entry cron.Entry
	Host string
	Service string
	LastRun time.Time
	HostServiceID int
	ScheduleText string
}

type Event struct {
	ID int
	EventType string
	HostServiceID int
	ServiceID int
	HostID int
	ServiceName string
	HostName string
	Message string
	CreatedAt time.Time
	UpdatedAt time.Time
}