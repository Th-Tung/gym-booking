package models

import (
	"database/sql"
	"time"
)

type Equipment struct {
	ID           int       `json:"id"`
	CategoryID   int       `json:"category_id"`
	CategoryName string    `json:"category_name,omitempty"`
	CategoryIcon string    `json:"category_icon,omitempty"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	ImagePath    string    `json:"image_path"`
	Status       string    `json:"status"` // available, maintenance
	CreatedAt    time.Time `json:"created_at"`
	IsInUse      bool      `json:"is_in_use,omitempty"`
}

func GetAllEquipments(db *sql.DB, categoryID int, status, search string) ([]Equipment, error) {
	query := `SELECT e.id, e.category_id, COALESCE(c.name, ''), COALESCE(c.icon, 'bi-tag'), 
	                 e.name, COALESCE(e.description, ''), COALESCE(e.image_path, 'assets/images/treadmill.svg'), 
	                 e.status, e.created_at 
	          FROM equipments e 
	          LEFT JOIN categories c ON e.category_id = c.id 
	          WHERE (? = 0 OR e.category_id = ?)
	            AND (? = '' OR e.status = ?)
	            AND (? = '' OR e.name LIKE ? OR e.description LIKE ?)
	          ORDER BY e.id ASC`
	searchParam := "%" + search + "%"
	rows, err := db.Query(query, categoryID, categoryID, status, status, search, searchParam, searchParam)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var equipments []Equipment
	for rows.Next() {
		var eq Equipment
		if err := rows.Scan(&eq.ID, &eq.CategoryID, &eq.CategoryName, &eq.CategoryIcon,
			&eq.Name, &eq.Description, &eq.ImagePath, &eq.Status, &eq.CreatedAt); err != nil {
			return nil, err
		}
		equipments = append(equipments, eq)
	}

	// Determine real-time in-use status
	nowStr := time.Now().Format("2006-01-02 15:04:05")
	inUseRows, err := db.Query("SELECT DISTINCT equipment_id FROM bookings WHERE status = 'confirmed' AND ? BETWEEN start_time AND end_time", nowStr)
	if err == nil {
		defer inUseRows.Close()
		inUseMap := make(map[int]bool)
		for inUseRows.Next() {
			var eqID int
			if err := inUseRows.Scan(&eqID); err == nil {
				inUseMap[eqID] = true
			}
		}
		for i := range equipments {
			if inUseMap[equipments[i].ID] {
				equipments[i].IsInUse = true
			}
		}
	}

	return equipments, nil
}

func GetEquipmentByID(db *sql.DB, id int) (*Equipment, error) {
	query := `SELECT e.id, e.category_id, COALESCE(c.name, ''), COALESCE(c.icon, 'bi-tag'), 
	                 e.name, COALESCE(e.description, ''), COALESCE(e.image_path, 'assets/images/treadmill.svg'), 
	                 e.status, e.created_at 
	          FROM equipments e 
	          LEFT JOIN categories c ON e.category_id = c.id 
	          WHERE e.id = ?`
	row := db.QueryRow(query, id)
	var eq Equipment
	err := row.Scan(&eq.ID, &eq.CategoryID, &eq.CategoryName, &eq.CategoryIcon,
		&eq.Name, &eq.Description, &eq.ImagePath, &eq.Status, &eq.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &eq, nil
}

func CreateEquipment(db *sql.DB, name string, categoryID int, description, imagePath, status string) error {
	_, err := db.Exec("INSERT INTO equipments (name, category_id, description, image_path, status) VALUES (?, ?, ?, ?, ?)",
		name, categoryID, description, imagePath, status)
	return err
}

func UpdateEquipment(db *sql.DB, id int, name string, categoryID int, description, imagePath, status string) error {
	_, err := db.Exec("UPDATE equipments SET name = ?, category_id = ?, description = ?, image_path = ?, status = ? WHERE id = ?",
		name, categoryID, description, imagePath, status, id)
	return err
}

func ToggleEquipmentStatus(db *sql.DB, id int) (string, error) {
	eq, err := GetEquipmentByID(db, id)
	if err != nil {
		return "", err
	}
	newStatus := "available"
	if eq.Status == "available" {
		newStatus = "maintenance"
	}
	_, err = db.Exec("UPDATE equipments SET status = ? WHERE id = ?", newStatus, id)
	return newStatus, err
}

func DeleteEquipment(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM equipments WHERE id = ?", id)
	return err
}
