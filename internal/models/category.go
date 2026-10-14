package models

import (
	"database/sql"
	"time"
)

type Category struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Icon            string    `json:"icon"`
	CreatedAt       time.Time `json:"created_at"`
	TotalEquipments int       `json:"total_equipments,omitempty"`
}

func GetAllCategories(db *sql.DB) ([]Category, error) {
	query := `SELECT c.id, c.name, COALESCE(c.description, ''), COALESCE(c.icon, 'bi-collection'), c.created_at, COUNT(e.id) AS total_equipments 
	          FROM categories c 
	          LEFT JOIN equipments e ON c.id = e.category_id 
	          GROUP BY c.id 
	          ORDER BY c.id ASC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.Icon, &c.CreatedAt, &c.TotalEquipments); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

func GetCategoryByID(db *sql.DB, id int) (*Category, error) {
	row := db.QueryRow("SELECT id, name, COALESCE(description, ''), COALESCE(icon, 'bi-collection'), created_at FROM categories WHERE id = ?", id)
	var c Category
	err := row.Scan(&c.ID, &c.Name, &c.Description, &c.Icon, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func CreateCategory(db *sql.DB, name, description, icon string) error {
	_, err := db.Exec("INSERT INTO categories (name, description, icon) VALUES (?, ?, ?)", name, description, icon)
	return err
}

func UpdateCategory(db *sql.DB, id int, name, description, icon string) error {
	_, err := db.Exec("UPDATE categories SET name = ?, description = ?, icon = ? WHERE id = ?", name, description, icon, id)
	return err
}

func DeleteCategory(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM categories WHERE id = ?", id)
	return err
}
