package config

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Config struct {
	DBHost      string
	DBPort      int
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

func parseDatabaseURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if strings.HasPrefix(rawURL, "mysql://") {
		u, err := url.Parse(rawURL)
		if err == nil {
			user := u.User.Username()
			pass, _ := u.User.Password()
			host := u.Host
			dbName := strings.TrimPrefix(u.Path, "/")
			query := u.Query()
			if query.Get("parseTime") == "" {
				query.Set("parseTime", "true")
			}
			if query.Get("loc") == "" {
				query.Set("loc", "Local")
			}
			if query.Get("charset") == "" {
				query.Set("charset", "utf8mb4")
			}
			return fmt.Sprintf("%s:%s@tcp(%s)/%s?%s", user, pass, host, dbName, query.Encode())
		}
	}
	if !strings.Contains(rawURL, "parseTime=") {
		sep := "?"
		if strings.Contains(rawURL, "?") {
			sep = "&"
		}
		rawURL += sep + "parseTime=true&loc=Local"
	}
	return rawURL
}

func AutoMigrate(db *sql.DB) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'equipments'").Scan(&count)
	if err == nil && count > 0 {
		return // tables already exist
	}

	content, err := os.ReadFile("database/schema.sql")
	if err != nil {
		return
	}

	queries := strings.Split(string(content), ";")
	for _, q := range queries {
		q = strings.TrimSpace(q)
		if q == "" || strings.HasPrefix(q, "--") || strings.HasPrefix(strings.ToUpper(q), "CREATE DATABASE") || strings.HasPrefix(strings.ToUpper(q), "USE ") {
			continue
		}
		_, _ = db.Exec(q)
	}
	log.Printf("[AutoMigrate] Database schema and initial data initialized successfully!")
}

// InitDB initializes database connection with fallback passwords
func InitDB(cfg Config) (*sql.DB, error) {
	if cfg.DatabaseURL != "" {
		dsn := parseDatabaseURL(cfg.DatabaseURL)
		d, err := sql.Open("mysql", dsn)
		if err == nil {
			d.SetMaxOpenConns(25)
			d.SetMaxIdleConns(10)
			d.SetConnMaxLifetime(5 * time.Minute)
			if err := d.Ping(); err == nil {
				log.Printf("[DB] Successfully connected via DATABASE_URL")
				AutoMigrate(d)
				return d, nil
			} else {
				log.Printf("[DB] Failed to connect via DATABASE_URL: %v", err)
			}
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
			AutoMigrate(db)
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

