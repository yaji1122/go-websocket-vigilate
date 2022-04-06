package dbrepo

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/tsawler/vigilate/internal/models"
	"log"
	"time"
)

func (m *postgresDBRepo) AllHosts() ([]*models.Host, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var hosts []*models.Host
	stmt := `SELECT * FROM hosts`
	rows, _ := m.DB.QueryContext(ctx, stmt)

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(rows)

	for rows.Next() {
		h := &models.Host{}
		err := rows.Scan(
			&h.ID,
			&h.HostName,
			&h.CanonicalName,
			&h.Url,
			&h.IP,
			&h.IPV6,
			&h.Location,
			&h.OS,
			&h.CreatedAt,
			&h.UpdatedAt,
			&h.Active,
		)
		h.HostServices, _ = m.GetHostServicesByHostId(h.ID)
		if err != nil {
			log.Println(err)
		}
		hosts = append(hosts, h)
	}

	return hosts, nil
}

func (m *postgresDBRepo) GetHostById(id int) (models.Host, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	//Get Host
	stmt := `SELECT id, host_name, canonical_name, url, ip, ipv6, location, os, active, created_at, updated_at FROM hosts where id = $1`
	row := m.DB.QueryRowContext(ctx, stmt, id)
	var h models.Host

	err := row.Scan(
		&h.ID,
		&h.HostName,
		&h.CanonicalName,
		&h.Url,
		&h.IP,
		&h.IPV6,
		&h.Location,
		&h.OS,
		&h.Active,
		&h.CreatedAt,
		&h.UpdatedAt,
	)

	if err != nil {
		log.Println(err)
		return h, err
	}

	return h, nil
}

// InsertHost inserts a host into database
func (m *postgresDBRepo) InsertHost(h models.Host) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Create a bcrypt hash of the plain-text password.

	stmt := `
	INSERT INTO hosts (host_name, canonical_name, url, ip, ipv6, location, os, active, created_at, updated_at)
    VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) returning id `

	var newId int
	err := m.DB.QueryRowContext(ctx, stmt,
		h.HostName,
		h.CanonicalName,
		h.Url,
		h.IP,
		h.IPV6,
		h.Location,
		h.OS,
		h.Active,
		time.Now(),
		time.Now(),
		).Scan(&newId)

	if err != nil {
		log.Println(err)
		return newId, err
	}

	stmt = `
		insert into host_services (host_id, service_id, active, schedule_number, schedule_unit, status, created_at, updated_at) 
		values ($1, 1, 0, 3, 'm', 'pending', $2, $3)`
	_, err = m.DB.ExecContext(ctx, stmt, newId, time.Now(), time.Now())

	if err != nil {
		log.Println(err)
		return newId, err
	}

	return newId, err
}

// UpdateHost update a host
func (m *postgresDBRepo) UpdateHost(h models.Host) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var err error
	// Create a bcrypt hash of the plain-text password.

	stmt := `UPDATE hosts 
			 SET host_name = $1, 
			     canonical_name = $2, 
			     url = $3, 
			     ip = $4, 
			     ipv6 = $5, 
			     location =$6, 
			     os = $7, 
			     active = $8, 
			     updated_at = $9 
			 WHERE id = $10`

	_, err = m.DB.ExecContext(ctx, stmt,
		h.HostName,
		h.CanonicalName,
		h.Url,
		h.IP,
		h.IPV6,
		h.Location,
		h.OS,
		h.Active,
		time.Now(),
		h.ID,
	)
	if err != nil {
		log.Println(err)
	}
	return err
}

func  (m *postgresDBRepo) UpdateHostServices(hs models.HostService) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `UPDATE host_services 
			 SET active = $1,
			     last_check = $2,
			     status = $3,
			     updated_at = $4
			 WHERE id = $5`

	_, err := m.DB.ExecContext(ctx, stmt,
		hs.Active,
		hs.LastCheck,
		hs.Status,
		time.Now(),
		hs.ID,
	)
	if err != nil {
		log.Println(err)
	}
	return err
}

//InsertEvent insert event to database
func (m *postgresDBRepo) InsertEvent(e models.Event) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `insert into events (event_type, host_service_id, host_id, service_id, service_name, host_name, message, created_at, updated_at) 
			 values ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := m.DB.ExecContext(ctx, stmt,
		    e.EventType,
			e.HostServiceID,
			e.HostID,
			e.ServiceID,
			e.ServiceName,
			e.HostName,
			e.Message,
			time.Now(),
			time.Now(),
		)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (m *postgresDBRepo) GetAllEvents() ([]*models.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `SELECT event_type, host_service_id, host_id, service_id, service_name, host_name, message, created_at FROM events`

	var events []*models.Event

	rows, err := m.DB.QueryContext(ctx, stmt)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if rows.Next() {
		event := &models.Event{}
		err = rows.Scan(
			&event.EventType,
			&event.HostServiceID,
			&event.HostID,
			&event.ServiceID,
			&event.ServiceName,
			&event.HostName,
			&event.Message,
			&event.CreatedAt)
		events = append(events, event)
	}

	return events, err
}