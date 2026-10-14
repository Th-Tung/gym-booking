package models

import (
	"database/sql"
	"fmt"
	"time"
)

type Booking struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	UserName      string    `json:"user_name,omitempty"`
	UserEmail     string    `json:"user_email,omitempty"`
	UserPhone     string    `json:"user_phone,omitempty"`
	EquipmentID   int       `json:"equipment_id"`
	EquipmentName string    `json:"equipment_name,omitempty"`
	CategoryName  string    `json:"category_name,omitempty"`
	ImagePath     string    `json:"image_path,omitempty"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Status        string    `json:"status"` // pending, confirmed, cancelled
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
}

type DashboardStats struct {
	TotalUsers       int
	TotalEquipments  int
	AvailableEqCount int
	MaintenanceCount int
	TodayBookings    int
	TotalBookings    int
}

type PopularEquipment struct {
	Name          string
	CategoryName  string
	TotalBookings int
}

type UserFavoriteEquipment struct {
	ID           int
	Name         string
	ImagePath    string
	BookingCount int
}

func CheckOverlap(db *sql.DB, equipmentID int, startTime, endTime string) (bool, error) {
	query := `SELECT COUNT(*) FROM bookings 
	          WHERE equipment_id = ? 
	            AND status != 'cancelled' 
	            AND start_time < ? 
	            AND end_time > ?`
	var count int
	err := db.QueryRow(query, equipmentID, endTime, startTime).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func CreateBooking(db *sql.DB, userID, equipmentID int, startTime, endTime, notes string) (*Booking, error) {
	// First check overlap
	hasOverlap, err := CheckOverlap(db, equipmentID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	if hasOverlap {
		return nil, fmt.Errorf("ช่วงเวลานี้มีผู้จองแล้ว กรุณาเลือกช่วงเวลาอื่น")
	}

	query := `INSERT INTO bookings (user_id, equipment_id, start_time, end_time, status, notes) 
	          VALUES (?, ?, ?, ?, 'confirmed', ?)`
	res, err := db.Exec(query, userID, equipmentID, startTime, endTime, notes)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return GetBookingByID(db, int(id))
}

func GetBookingByID(db *sql.DB, id int) (*Booking, error) {
	query := `SELECT b.id, b.user_id, COALESCE(u.name, ''), COALESCE(u.email, ''), COALESCE(u.phone, ''),
	                 b.equipment_id, COALESCE(e.name, ''), COALESCE(c.name, ''), COALESCE(e.image_path, ''),
	                 b.start_time, b.end_time, b.status, COALESCE(b.notes, ''), b.created_at
	          FROM bookings b
	          JOIN users u ON b.user_id = u.id
	          JOIN equipments e ON b.equipment_id = e.id
	          LEFT JOIN categories c ON e.category_id = c.id
	          WHERE b.id = ?`
	row := db.QueryRow(query, id)
	var b Booking
	err := row.Scan(&b.ID, &b.UserID, &b.UserName, &b.UserEmail, &b.UserPhone,
		&b.EquipmentID, &b.EquipmentName, &b.CategoryName, &b.ImagePath,
		&b.StartTime, &b.EndTime, &b.Status, &b.Notes, &b.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func GetBookedSlotsForDate(db *sql.DB, equipmentID int, date string) ([]Booking, error) {
	query := `SELECT id, start_time, end_time, status 
	          FROM bookings 
	          WHERE equipment_id = ? 
	            AND status != 'cancelled' 
	            AND (DATE(start_time) = ? OR DATE(end_time) = ?)
	          ORDER BY start_time ASC`
	rows, err := db.Query(query, equipmentID, date, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Booking
	for rows.Next() {
		var b Booking
		if err := rows.Scan(&b.ID, &b.StartTime, &b.EndTime, &b.Status); err != nil {
			return nil, err
		}
		list = append(list, b)
	}
	return list, nil
}

func GetUpcomingBookingsByUser(db *sql.DB, userID int, limit int) ([]Booking, error) {
	nowStr := time.Now().Format("2006-01-02 15:04:05")
	query := `SELECT b.id, b.user_id, COALESCE(u.name, ''), COALESCE(u.email, ''), COALESCE(u.phone, ''),
	                 b.equipment_id, COALESCE(e.name, ''), COALESCE(c.name, ''), COALESCE(e.image_path, ''),
	                 b.start_time, b.end_time, b.status, COALESCE(b.notes, ''), b.created_at
	          FROM bookings b
	          JOIN users u ON b.user_id = u.id
	          JOIN equipments e ON b.equipment_id = e.id
	          LEFT JOIN categories c ON e.category_id = c.id
	          WHERE b.user_id = ? AND b.status = 'confirmed' AND b.end_time >= ?
	          ORDER BY b.start_time ASC LIMIT ?`
	rows, err := db.Query(query, userID, nowStr, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Booking
	for rows.Next() {
		var b Booking
		if err := rows.Scan(&b.ID, &b.UserID, &b.UserName, &b.UserEmail, &b.UserPhone,
			&b.EquipmentID, &b.EquipmentName, &b.CategoryName, &b.ImagePath,
			&b.StartTime, &b.EndTime, &b.Status, &b.Notes, &b.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, b)
	}
	return list, nil
}

func GetAllBookingsByUser(db *sql.DB, userID int, filterStatus string) ([]Booking, error) {
	nowStr := time.Now().Format("2006-01-02 15:04:05")
	query := `SELECT b.id, b.user_id, COALESCE(u.name, ''), COALESCE(u.email, ''), COALESCE(u.phone, ''),
	                 b.equipment_id, COALESCE(e.name, ''), COALESCE(c.name, ''), COALESCE(e.image_path, ''),
	                 b.start_time, b.end_time, b.status, COALESCE(b.notes, ''), b.created_at
	          FROM bookings b
	          JOIN users u ON b.user_id = u.id
	          JOIN equipments e ON b.equipment_id = e.id
	          LEFT JOIN categories c ON e.category_id = c.id
	          WHERE b.user_id = ?`

	params := []interface{}{userID}
	if filterStatus == "upcoming" {
		query += " AND b.status = 'confirmed' AND b.end_time >= ?"
		params = append(params, nowStr)
	} else if filterStatus == "completed" {
		query += " AND b.status = 'confirmed' AND b.end_time < ?"
		params = append(params, nowStr)
	} else if filterStatus == "cancelled" {
		query += " AND b.status = 'cancelled'"
	}

	query += " ORDER BY b.start_time DESC"
	rows, err := db.Query(query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Booking
	for rows.Next() {
		var b Booking
		if err := rows.Scan(&b.ID, &b.UserID, &b.UserName, &b.UserEmail, &b.UserPhone,
			&b.EquipmentID, &b.EquipmentName, &b.CategoryName, &b.ImagePath,
			&b.StartTime, &b.EndTime, &b.Status, &b.Notes, &b.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, b)
	}
	return list, nil
}

func GetAllBookings(db *sql.DB, dateFilter string, equipmentID int, statusFilter string) ([]Booking, error) {
	query := `SELECT b.id, b.user_id, COALESCE(u.name, ''), COALESCE(u.email, ''), COALESCE(u.phone, ''),
	                 b.equipment_id, COALESCE(e.name, ''), COALESCE(c.name, ''), COALESCE(e.image_path, ''),
	                 b.start_time, b.end_time, b.status, COALESCE(b.notes, ''), b.created_at
	          FROM bookings b
	          JOIN users u ON b.user_id = u.id
	          JOIN equipments e ON b.equipment_id = e.id
	          LEFT JOIN categories c ON e.category_id = c.id
	          WHERE (? = '' OR DATE(b.start_time) = ?)
	            AND (? = 0 OR b.equipment_id = ?)
	            AND (? = 'all' OR b.status = ?)
	          ORDER BY b.start_time DESC LIMIT 100`
	rows, err := db.Query(query, dateFilter, dateFilter, equipmentID, equipmentID, statusFilter, statusFilter)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Booking
	for rows.Next() {
		var b Booking
		if err := rows.Scan(&b.ID, &b.UserID, &b.UserName, &b.UserEmail, &b.UserPhone,
			&b.EquipmentID, &b.EquipmentName, &b.CategoryName, &b.ImagePath,
			&b.StartTime, &b.EndTime, &b.Status, &b.Notes, &b.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, b)
	}
	return list, nil
}

func CancelBooking(db *sql.DB, bookingID, userID int, isAdmin bool, reason string) error {
	b, err := GetBookingByID(db, bookingID)
	if err != nil {
		return fmt.Errorf("ไม่พบรายการจองนี้: %w", err)
	}
	if !isAdmin && b.UserID != userID {
		return fmt.Errorf("คุณไม่มีสิทธิ์ยกเลิกรายการจองของผู้อื่น")
	}
	if b.Status == "cancelled" {
		return fmt.Errorf("รายการนี้ถูกยกเลิกไปแล้ว")
	}

	notes := b.Notes
	who := "Member"
	if isAdmin {
		who = "Admin"
	}
	if reason != "" {
		if notes != "" {
			notes += "\n"
		}
		notes += fmt.Sprintf("[Cancelled by %s]: %s", who, reason)
	}

	_, err = db.Exec("UPDATE bookings SET status = 'cancelled', notes = ? WHERE id = ?", notes, bookingID)
	return err
}

func GetDashboardStats(db *sql.DB) (DashboardStats, error) {
	var s DashboardStats
	today := time.Now().Format("2006-01-02")

	_ = db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'member'").Scan(&s.TotalUsers)
	_ = db.QueryRow("SELECT COUNT(*) FROM equipments").Scan(&s.TotalEquipments)
	_ = db.QueryRow("SELECT COUNT(*) FROM equipments WHERE status = 'available'").Scan(&s.AvailableEqCount)
	_ = db.QueryRow("SELECT COUNT(*) FROM equipments WHERE status = 'maintenance'").Scan(&s.MaintenanceCount)
	_ = db.QueryRow("SELECT COUNT(*) FROM bookings WHERE DATE(start_time) = ? AND status = 'confirmed'", today).Scan(&s.TodayBookings)
	_ = db.QueryRow("SELECT COUNT(*) FROM bookings WHERE status = 'confirmed'").Scan(&s.TotalBookings)

	return s, nil
}

func GetPopularEquipments(db *sql.DB) ([]PopularEquipment, error) {
	query := `SELECT e.name, COALESCE(c.name, ''), COUNT(b.id) AS total_bookings 
	          FROM equipments e 
	          LEFT JOIN categories c ON e.category_id = c.id 
	          LEFT JOIN bookings b ON e.id = b.equipment_id AND b.status = 'confirmed' 
	          GROUP BY e.id 
	          ORDER BY total_bookings DESC LIMIT 5`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []PopularEquipment
	for rows.Next() {
		var p PopularEquipment
		if err := rows.Scan(&p.Name, &p.CategoryName, &p.TotalBookings); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func GetUserFavoriteEquipments(db *sql.DB, userID int) ([]UserFavoriteEquipment, error) {
	query := `SELECT e.id, e.name, COALESCE(e.image_path, 'assets/images/treadmill.svg'), COUNT(b.id) AS booking_count 
	          FROM bookings b 
	          JOIN equipments e ON b.equipment_id = e.id 
	          WHERE b.user_id = ? 
	          GROUP BY e.id ORDER BY booking_count DESC LIMIT 3`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []UserFavoriteEquipment
	for rows.Next() {
		var f UserFavoriteEquipment
		if err := rows.Scan(&f.ID, &f.Name, &f.ImagePath, &f.BookingCount); err != nil {
			return nil, err
		}
		list = append(list, f)
	}
	return list, nil
}
