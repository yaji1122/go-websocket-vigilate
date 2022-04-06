package dbrepo

import (
	"context"
	"github.com/tsawler/vigilate/internal/models"
	"log"
	"time"
)

func (m *postgresDBRepo) GetAllServiceStatusCount() (int, int, int, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `SELECT
			 (SELECT count(id) FROM host_services where active = true AND status  = 'healthy'),
			 (SELECT count(id) FROM host_services where active = true AND status  = 'warning'),
			 (SELECT count(id) FROM host_services where active = true AND status  = 'problem'),
			 (SELECT count(id) FROM host_services where active = true AND status  = 'pending')
			 `

	var healthy, warning, problem, pending int
	row := m.DB.QueryRowContext(ctx, stmt)
	err := row.Scan(
		&healthy,
		&warning,
		&problem,
		&pending)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return healthy, warning, problem, pending, nil
}

func (m *postgresDBRepo) GetServiceStatusCount(queryStatus string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `SELECT
			 (SELECT count(id) FROM host_services where active = true AND status = $1)
			 `

	var count int
	row := m.DB.QueryRowContext(ctx, stmt, queryStatus)
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *postgresDBRepo) GetHostServicesByHostId(id int) ([]models.HostService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	//Get Host
	stmt := `
			SELECT 
			       hs.id, hs.service_id, hs.host_id, hs.active, hs.schedule_number, hs.schedule_unit, hs.last_check, hs.created_at, hs.status,
				   s.service_name, s.icon
			FROM host_services hs LEFT JOIN services s on s.id = hs.service_id
			WHERE hs.host_id = $1`
	rows, err := m.DB.QueryContext(ctx, stmt, id)
	var hostServices []models.HostService

	for rows.Next() {
		var hs models.HostService
		err = rows.Scan(
			&hs.ID,
			&hs.ServiceID,
			&hs.HostID,
			&hs.Active,
			&hs.ScheduleNumber,
			&hs.ScheduleUnit,
			&hs.LastCheck,
			&hs.CreatedAt,
			&hs.Status,
			&hs.Service.ServiceName,
			&hs.Service.Icon)

		if err != nil {
			log.Println(err)
		}

		hostServices = append(hostServices, hs)
	}

	return hostServices, nil
}

func (m *postgresDBRepo) GetHostServicesByStatus(status string) ([]models.HostService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	//Get HostServices
	stmt := `
			SELECT 
			       hs.id, hs.service_id, hs.host_id, hs.active, hs.schedule_number, hs.schedule_unit, hs.last_check, hs.created_at, hs.status,
				   s.service_name, h.host_name
			FROM host_services hs LEFT JOIN services s on s.id = hs.service_id LEFT JOIN hosts h on h.id = hs.host_id
			WHERE hs.status = $1 AND hs.active = true
			ORDER BY h.host_name, s.service_name
			`
	rows, err := m.DB.QueryContext(ctx, stmt, status)
	var hostServices []models.HostService

	for rows.Next() {
		var hs models.HostService
		err = rows.Scan(
			&hs.ID,
			&hs.ServiceID,
			&hs.HostID,
			&hs.Active,
			&hs.ScheduleNumber,
			&hs.ScheduleUnit,
			&hs.LastCheck,
			&hs.CreatedAt,
			&hs.Status,
			&hs.Service.ServiceName,
			&hs.Host.HostName,
		)

		if err != nil {
			log.Println(err)
		}

		hostServices = append(hostServices, hs)
	}

	return hostServices, nil
}

func (m *postgresDBRepo) GetHostServicesById(id int) (models.HostService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	//Get Host
	stmt := `
			SELECT 
			       hs.id, hs.service_id, hs.host_id, hs.active, hs.schedule_number, hs.schedule_unit, hs.last_check, hs.created_at, hs.status,
				   s.service_name, s.icon, s.active, h.host_name
			FROM host_services hs 
			LEFT JOIN services s on s.id = hs.service_id
			LEFT JOIN hosts h on h.id = hs.host_id
			WHERE hs.id = $1`
	row := m.DB.QueryRowContext(ctx, stmt, id)

	var hs models.HostService
	err := row.Scan(
		&hs.ID,
		&hs.ServiceID,
		&hs.HostID,
		&hs.Active,
		&hs.ScheduleNumber,
		&hs.ScheduleUnit,
		&hs.LastCheck,
		&hs.CreatedAt,
		&hs.Status,
		&hs.Service.ServiceName,
		&hs.Service.Icon,
		&hs.Service.Active,
		&hs.Host.HostName)
	if err != nil {
		log.Println(err)
	}
	return hs, nil
}

func (m *postgresDBRepo) GetHostServicesByHostIDAndServiceID(hostID int, ServiceID int) (models.HostService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	//Get Host
	stmt := `
			SELECT 
			       hs.id, hs.service_id, hs.host_id, hs.active, hs.schedule_number, hs.schedule_unit, hs.last_check, hs.created_at, hs.status,
				   s.service_name, s.icon, s.active
			FROM host_services hs LEFT JOIN services s on s.id = hs.service_id
			WHERE hs.host_id = $1 AND hs.service_id = $2`
	row := m.DB.QueryRowContext(ctx, stmt, hostID, ServiceID)

	var hs models.HostService
	err := row.Scan(
		&hs.ID,
		&hs.ServiceID,
		&hs.HostID,
		&hs.Active,
		&hs.ScheduleNumber,
		&hs.ScheduleUnit,
		&hs.LastCheck,
		&hs.CreatedAt,
		&hs.Status,
		&hs.Service.ServiceName,
		&hs.Service.Icon,
		&hs.Service.Active)
	if err != nil {
		log.Println(err)
	}
	return hs, nil
}

func (m *postgresDBRepo) GetServicesToMonitor() ([]models.HostService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT 
			   hs.id, hs.service_id, hs.host_id, hs.active, hs.schedule_number, hs.schedule_unit, hs.last_check, hs.created_at, hs.status,
			   s.service_name, s.icon, s.active, h.host_name
			FROM host_services hs 
			LEFT JOIN services s on s.id = hs.service_id
			LEFT JOIN hosts h on h.id = hs.host_id
			WHERE h.active = true AND hs.active = true`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		log.Println(err)
	}

	var hostServices []models.HostService
	if rows.Next() {
		var hs models.HostService
		err = rows.Scan(
			&hs.ID,
			&hs.ServiceID,
			&hs.HostID,
			&hs.Active,
			&hs.ScheduleNumber,
			&hs.ScheduleUnit,
			&hs.LastCheck,
			&hs.CreatedAt,
			&hs.Status,
			&hs.Service.ServiceName,
			&hs.Service.Icon,
			&hs.Service.Active,
			&hs.Host.HostName)
		if err != nil {
			log.Println(err)
		}
		hostServices = append(hostServices, hs)
	}

	return hostServices, err
}

