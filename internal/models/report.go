package models

import (
	"database/sql"
	"time"
)

type MaintenanceReport struct {
	ID               int        `json:"id"`
	EquipmentID      int        `json:"equipment_id"`
	EquipmentName    string     `json:"equipment_name,omitempty"`
	CategoryName     string     `json:"category_name,omitempty"`
	ImagePath        string     `json:"image_path,omitempty"`
	UserID           int        `json:"user_id"`
	UserName         string     `json:"user_name,omitempty"`
	UserEmail        string     `json:"user_email,omitempty"`
	UserPhone        string     `json:"user_phone,omitempty"`
	IssueDescription string     `json:"issue_description"`
	Status           string     `json:"status"` // pending, confirmed_maintenance, resolved, rejected
	AdminNotes       string     `json:"admin_notes"`
	CreatedAt        time.Time  `json:"created_at"`
	ResolvedAt       *time.Time `json:"resolved_at,omitempty"`
}

func CreateMaintenanceReport(db *sql.DB, equipmentID, userID int, issueDescription string) error {
	query := `INSERT INTO maintenance_reports (equipment_id, user_id, issue_description, status) 
	          VALUES (?, ?, ?, 'pending')`
	_, err := db.Exec(query, equipmentID, userID, issueDescription)
	return err
}

func GetAllMaintenanceReports(db *sql.DB, statusFilter string) ([]MaintenanceReport, error) {
	query := `SELECT r.id, r.equipment_id, COALESCE(e.name, ''), COALESCE(c.name, ''), COALESCE(e.image_path, ''),
	                 r.user_id, COALESCE(u.name, ''), COALESCE(u.email, ''), COALESCE(u.phone, ''),
	                 r.issue_description, r.status, COALESCE(r.admin_notes, ''), r.created_at, r.resolved_at
	          FROM maintenance_reports r
	          JOIN equipments e ON r.equipment_id = e.id
	          LEFT JOIN categories c ON e.category_id = c.id
	          JOIN users u ON r.user_id = u.id
	          WHERE (? = '' OR ? = 'all' OR r.status = ?)
	          ORDER BY r.created_at DESC`
	rows, err := db.Query(query, statusFilter, statusFilter, statusFilter)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []MaintenanceReport
	for rows.Next() {
		var r MaintenanceReport
		var resAt sql.NullTime
		if err := rows.Scan(&r.ID, &r.EquipmentID, &r.EquipmentName, &r.CategoryName, &r.ImagePath,
			&r.UserID, &r.UserName, &r.UserEmail, &r.UserPhone,
			&r.IssueDescription, &r.Status, &r.AdminNotes, &r.CreatedAt, &resAt); err != nil {
			return nil, err
		}
		if resAt.Valid {
			r.ResolvedAt = &resAt.Time
		}
		list = append(list, r)
	}
	return list, nil
}

// ConfirmMaintenanceReport accepts the report, marks equipment as maintenance, and cancels future bookings for that equipment
func ConfirmMaintenanceReport(db *sql.DB, reportID int, adminNotes string) error {
	var equipmentID int
	err := db.QueryRow("SELECT equipment_id FROM maintenance_reports WHERE id = ?", reportID).Scan(&equipmentID)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Update report status
	_, err = tx.Exec("UPDATE maintenance_reports SET status = 'confirmed_maintenance', admin_notes = ? WHERE id = ?", adminNotes, reportID)
	if err != nil {
		return err
	}

	// 2. Set equipment to maintenance (closes equipment)
	_, err = tx.Exec("UPDATE equipments SET status = 'maintenance' WHERE id = ?", equipmentID)
	if err != nil {
		return err
	}

	// 3. Cancel upcoming confirmed bookings on this equipment
	nowStr := time.Now().Format("2006-01-02 15:04:05")
	cancelNote := "\n[System]: คิวถูกยกเลิกเนื่องจากอุปกรณ์ปิดปรับปรุงซ่อมบำรุงชั่วคราว"
	if adminNotes != "" {
		cancelNote += " (" + adminNotes + ")"
	}
	_, err = tx.Exec(`UPDATE bookings 
	                  SET status = 'cancelled', notes = CONCAT(COALESCE(notes, ''), ?) 
	                  WHERE equipment_id = ? AND status = 'confirmed' AND start_time >= ?`,
		cancelNote, equipmentID, nowStr)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// ResolveMaintenanceReport marks the equipment as repaired and back to available
func ResolveMaintenanceReport(db *sql.DB, reportID int, adminNotes string) error {
	var equipmentID int
	err := db.QueryRow("SELECT equipment_id FROM maintenance_reports WHERE id = ?", reportID).Scan(&equipmentID)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()
	_, err = tx.Exec("UPDATE maintenance_reports SET status = 'resolved', admin_notes = ?, resolved_at = ? WHERE id = ?", adminNotes, now, reportID)
	if err != nil {
		return err
	}

	// Set equipment back to available
	_, err = tx.Exec("UPDATE equipments SET status = 'available' WHERE id = ?", equipmentID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// RejectMaintenanceReport rejects the report if equipment is checked and fine
func RejectMaintenanceReport(db *sql.DB, reportID int, adminNotes string) error {
	_, err := db.Exec("UPDATE maintenance_reports SET status = 'rejected', admin_notes = ? WHERE id = ?", adminNotes, reportID)
	return err
}

func GetPendingReportsCount(db *sql.DB) int {
	var count int
	_ = db.QueryRow("SELECT COUNT(*) FROM maintenance_reports WHERE status = 'pending'").Scan(&count)
	return count
}
