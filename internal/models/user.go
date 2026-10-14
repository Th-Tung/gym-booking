package models

import (
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	PasswordHash  string    `json:"-"`
	Role          string    `json:"role"` // admin, member
	Phone         string    `json:"phone"`
	CreatedAt     time.Time `json:"created_at"`
	TotalBookings int       `json:"total_bookings,omitempty"`
}

func (u *User) CheckPassword(plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(plainPassword))
	return err == nil
}

func HashPassword(plainPassword string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	return string(bytes), err
}

func GetUserByEmail(db *sql.DB, email string) (*User, error) {
	row := db.QueryRow("SELECT id, name, email, password_hash, role, COALESCE(phone, ''), created_at FROM users WHERE email = ?", email)
	u := &User{}
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.Phone, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

func GetUserByID(db *sql.DB, id int) (*User, error) {
	row := db.QueryRow("SELECT id, name, email, password_hash, role, COALESCE(phone, ''), created_at FROM users WHERE id = ?", id)
	u := &User{}
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.Phone, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

func CreateUser(db *sql.DB, name, email, plainPassword, phone string) (*User, error) {
	hash, err := HashPassword(plainPassword)
	if err != nil {
		return nil, err
	}
	res, err := db.Exec("INSERT INTO users (name, email, password_hash, role, phone) VALUES (?, ?, ?, 'member', ?)",
		name, email, hash, phone)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return GetUserByID(db, int(id))
}

func GetAllUsers(db *sql.DB, search string) ([]User, error) {
	query := `SELECT u.id, u.name, u.email, u.role, COALESCE(u.phone, ''), u.created_at, COUNT(b.id) AS total_bookings 
	          FROM users u 
	          LEFT JOIN bookings b ON u.id = b.user_id 
	          WHERE (? = '' OR u.name LIKE ? OR u.email LIKE ? OR u.phone LIKE ?)
	          GROUP BY u.id 
	          ORDER BY u.role ASC, u.id ASC`
	searchParam := "%" + search + "%"
	rows, err := db.Query(query, search, searchParam, searchParam, searchParam)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.Phone, &u.CreatedAt, &u.TotalBookings); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func UpdateUserRole(db *sql.DB, id int, role string) error {
	_, err := db.Exec("UPDATE users SET role = ? WHERE id = ?", role, id)
	return err
}
