package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Config struct {
	DBHost string
	DBPort int
	DBName      string
	DBUser      string
	DBPass      string
	Port        string
	DatabaseURL string
}

func GetConfig() Config {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "127.0.0.1"
	}

	port := 3306
	if pStr := os.Getenv("DB_PORT"); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil {
			port = p
		}
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "gym_booking"
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = "root"
	}

	pass := os.Getenv("DB_PASS")
	if pass == "" && os.Getenv("DB_HOST") == "" {
		pass = "abcd1234"
	}

	appPort := os.Getenv("PORT")
	if appPort == "" {
		appPort = "8000"
	}

	return Config{
		DBHost:      host,
		DBPort:      port,
		DBName:      dbName,
		DBUser:      user,
		DBPass:      pass,
		Port:        appPort,
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
}

// InitDB initializes database connection with fallback passwords
func InitDB(cfg Config) (*sql.DB, error) {
	if cfg.DatabaseURL != "" {
		d, err := sql.Open("mysql", cfg.DatabaseURL)
		if err == nil && d.Ping() == nil {
			log.Printf("[DB] Successfully connected via DATABASE_URL")
			return d, nil
		}
	}

	passwords := []string{cfg.DBPass, "", "root", "12345678"}
	var db *sql.DB
	var lastErr error

	for _, pass := range passwords {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local&charset=utf8mb4&collation=utf8mb4_unicode_ci",
			cfg.DBUser, pass, cfg.DBHost, cfg.DBPort, cfg.DBName)

		d, err := sql.Open("mysql", dsn)
		if err != nil {
			lastErr = err
			continue
		}

		d.SetMaxOpenConns(25)
		d.SetMaxIdleConns(10)
		d.SetConnMaxLifetime(5 * time.Minute)

		if err := d.Ping(); err == nil {
			log.Printf("[DB] Successfully connected to MySQL at %s:%d/%s", cfg.DBHost, cfg.DBPort, cfg.DBName)
			db = d
			break
		} else {
			d.Close()
			lastErr = err
		}
	}

	if db == nil {
		return nil, fmt.Errorf("could not connect to MySQL: %w", lastErr)
	}

	return db, nil
}
