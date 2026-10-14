package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"gym-booking/internal/middleware"
	"gym-booking/internal/models"
)

type PageData struct {
	Title     string
	User      *models.User
	Flash     *middleware.Flash
	CSRFToken string
	ActiveNav string
	Data      interface{}
}

type Handler struct {
	DB *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{DB: db}
}

var funcMap = template.FuncMap{
	"formatDate": func(t time.Time) string {
		return t.Format("02/01/2006")
	},
	"formatTime": func(t time.Time) string {
		return t.Format("15:04")
	},
	"formatDateTime": func(t time.Time) string {
		return t.Format("02/01/2006 15:04")
	},
	"isPast": func(t time.Time) bool {
		return t.Before(time.Now())
	},
	"padID": func(id int, prefix string) string {
		return fmt.Sprintf("#%s%05d", prefix, id)
	},
	"add": func(a, b int) int {
		return a + b
	},
	"isOngoing": func(start, end time.Time) bool {
		now := time.Now()
		return (now.After(start) || now.Equal(start)) && (now.Before(end) || now.Equal(end))
	},
}

func RenderTemplate(w http.ResponseWriter, r *http.Request, tmplName string, title string, activeNav string, data interface{}) {
	basePath := filepath.Join("templates", "layouts", "base.html")
	viewPath := filepath.Join("templates", tmplName)

	tmpl, err := template.New("base.html").Funcs(funcMap).ParseFiles(basePath, viewPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Template parse error: %v", err), http.StatusInternalServerError)
		return
	}

	pData := PageData{
		Title:     title,
		User:      middleware.GetCurrentUser(r),
		Flash:     middleware.GetFlash(w, r),
		CSRFToken: middleware.GenerateCSRFToken(w, r),
		ActiveNav: activeNav,
		Data:      data,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base.html", pData); err != nil {
		http.Error(w, fmt.Sprintf("Template execute error: %v", err), http.StatusInternalServerError)
	}
}
