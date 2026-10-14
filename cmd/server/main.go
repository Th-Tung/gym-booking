package main

import (
	"fmt"
	"log"
	"net/http"

	"gym-booking/internal/config"
	"gym-booking/internal/handlers"
	"gym-booking/internal/middleware"
)

func main() {
	cfg := config.GetConfig()

	db, err := config.InitDB(cfg)
	if err != nil {
		log.Fatalf("Fatal: Database initialization failed: %v", err)
	}
	defer db.Close()

	h := handlers.NewHandler(db)
	mux := http.NewServeMux()

	// Static Assets
	fs := http.FileServer(http.Dir("assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// Public Routes
	mux.HandleFunc("GET /{$}", h.ShowHome)
	mux.HandleFunc("GET /equipments", h.ShowEquipments)
	mux.HandleFunc("GET /login", h.ShowLogin)
	mux.HandleFunc("POST /login", h.ProcessLogin)
	mux.HandleFunc("GET /register", h.ShowRegister)
	mux.HandleFunc("POST /register", h.ProcessRegister)
	mux.HandleFunc("GET /logout", h.ProcessLogout)
	mux.HandleFunc("GET /api/availability", h.GetAvailabilityAPI)

	// Member Protected Routes
	mux.HandleFunc("GET /dashboard", middleware.RequireAuth(h.ShowDashboard))
	mux.HandleFunc("GET /booking", middleware.RequireAuth(h.ShowBooking))
	mux.HandleFunc("POST /booking", middleware.RequireAuth(h.ProcessBooking))
	mux.HandleFunc("GET /my-bookings", middleware.RequireAuth(h.ShowMyBookings))
	mux.HandleFunc("POST /api/cancel-booking", middleware.RequireAuth(h.CancelBookingAPI))
	mux.HandleFunc("POST /report-issue", middleware.RequireAuth(h.ProcessReportIssue))

	// Admin Protected Routes
	mux.HandleFunc("GET /admin", middleware.RequireAdmin(h.ShowAdminDashboard))
	mux.HandleFunc("GET /admin/reports", middleware.RequireAdmin(h.ShowAdminReports))
	mux.HandleFunc("POST /admin/reports/confirm", middleware.RequireAdmin(h.ProcessConfirmReport))
	mux.HandleFunc("POST /admin/reports/resolve", middleware.RequireAdmin(h.ProcessResolveReport))
	mux.HandleFunc("POST /admin/reports/reject", middleware.RequireAdmin(h.ProcessRejectReport))
	mux.HandleFunc("GET /admin/equipments", middleware.RequireAdmin(h.ShowAdminEquipments))
	mux.HandleFunc("POST /admin/equipments/add", middleware.RequireAdmin(h.ProcessAddEquipment))
	mux.HandleFunc("POST /admin/equipments/edit", middleware.RequireAdmin(h.ProcessEditEquipment))
	mux.HandleFunc("POST /admin/equipments/toggle", middleware.RequireAdmin(h.ProcessToggleEquipment))
	mux.HandleFunc("POST /admin/equipments/delete", middleware.RequireAdmin(h.ProcessDeleteEquipment))
	mux.HandleFunc("GET /admin/categories", middleware.RequireAdmin(h.ShowAdminCategories))
	mux.HandleFunc("POST /admin/categories/add", middleware.RequireAdmin(h.ProcessAddCategory))
	mux.HandleFunc("POST /admin/categories/edit", middleware.RequireAdmin(h.ProcessEditCategory))
	mux.HandleFunc("POST /admin/categories/delete", middleware.RequireAdmin(h.ProcessDeleteCategory))
	mux.HandleFunc("GET /admin/bookings", middleware.RequireAdmin(h.ShowAdminBookings))
	mux.HandleFunc("GET /admin/users", middleware.RequireAdmin(h.ShowAdminUsers))
	mux.HandleFunc("POST /admin/users/role", middleware.RequireAdmin(h.ProcessUpdateUserRole))

	// Global Middleware Pipeline
	handler := middleware.InjectUser(db)(mux)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("=====================================================")
	log.Printf(" Gym Equipment Booking System - Go Web Server Running ")
	log.Printf(" URL: http://localhost%s                              ", addr)
	log.Printf(" Admin Account: admin@gym.com / Admin@1234           ")
	log.Printf(" Member Account: member@gym.com / Member@1234         ")
	log.Printf("=====================================================")

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server stopped: %v", err)
	}
}
