package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gym-booking/internal/middleware"
	"gym-booking/internal/models"
)

func (h *Handler) ShowHome(w http.ResponseWriter, r *http.Request) {
	stats, err := models.GetDashboardStats(h.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	equipments, err := models.GetAllEquipments(h.DB, 0, "", "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(equipments) > 6 {
		equipments = equipments[:6]
	}

	RenderTemplate(w, r, "home.html", "หน้าหลัก", "home", map[string]interface{}{
		"Stats":      stats,
		"Equipments": equipments,
	})
}

func (h *Handler) ShowEquipments(w http.ResponseWriter, r *http.Request) {
	catID, _ := strconv.Atoi(r.URL.Query().Get("category"))
	status := r.URL.Query().Get("status")
	query := strings.TrimSpace(r.URL.Query().Get("q"))

	categories, err := models.GetAllCategories(h.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	equipments, err := models.GetAllEquipments(h.DB, catID, status, query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	RenderTemplate(w, r, "equipments.html", "เครื่องเล่นและอุปกรณ์", "equipments", map[string]interface{}{
		"Categories":       categories,
		"Equipments":       equipments,
		"SelectedCategory": catID,
		"SearchQuery":      query,
	})
}

func (h *Handler) ShowBooking(w http.ResponseWriter, r *http.Request) {
	preselectedEqID, _ := strconv.Atoi(r.URL.Query().Get("equipment_id"))
	selectedDate := r.URL.Query().Get("date")
	if selectedDate == "" {
		selectedDate = time.Now().Format("2006-01-02")
	}

	equipments, err := models.GetAllEquipments(h.DB, 0, "", "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	minDate := time.Now().Format("2006-01-02")
	maxDate := time.Now().AddDate(0, 0, 14).Format("2006-01-02")

	RenderTemplate(w, r, "booking.html", "จองเวลาใช้อุปกรณ์", "booking", map[string]interface{}{
		"Equipments":      equipments,
		"PreselectedEqID": preselectedEqID,
		"SelectedDate":    selectedDate,
		"MinDate":         minDate,
		"MaxDate":         maxDate,
		"Errors":          []string{},
	})
}

func (h *Handler) ProcessBooking(w http.ResponseWriter, r *http.Request) {
	u := middleware.GetCurrentUser(r)
	if !middleware.VerifyCSRF(r) {
		middleware.SetFlash(w, r, "danger", "Session หมดอายุหรือไม่ถูกต้อง กรุณาลองใหม่อีกครั้ง")
		http.Redirect(w, r, "/booking", http.StatusSeeOther)
		return
	}

	eqID, _ := strconv.Atoi(r.FormValue("equipment_id"))
	startTimeStr := strings.TrimSpace(r.FormValue("start_time"))
	endTimeStr := strings.TrimSpace(r.FormValue("end_time"))
	notes := strings.TrimSpace(r.FormValue("notes"))

	var errors []string
	if eqID <= 0 {
		errors = append(errors, "กรุณาเลือกอุปกรณ์ที่ต้องการจอง")
	}
	if startTimeStr == "" || endTimeStr == "" {
		errors = append(errors, "กรุณาเลือกช่วงเวลาที่ต้องการจอง")
	}

	startTime, err1 := time.Parse("2006-01-02 15:04:05", startTimeStr)
	endTime, err2 := time.Parse("2006-01-02 15:04:05", endTimeStr)

	if err1 != nil || err2 != nil {
		errors = append(errors, "รูปแบบเวลาไม่ถูกต้อง")
	} else if startTime.Before(time.Now()) {
		errors = append(errors, "ไม่สามารถจองช่วงเวลาในอดีตได้ กรุณาเลือกช่วงเวลาถัดไป")
	} else if !startTime.Before(endTime) {
		errors = append(errors, "เวลาเริ่มต้นต้องมาก่อนเวลาสิ้นสุด")
	}

	if len(errors) == 0 {
		eq, err := models.GetEquipmentByID(h.DB, eqID)
		if err != nil || eq == nil {
			errors = append(errors, "ไม่พบอุปกรณ์ที่ระบุ")
		} else if eq.Status == "maintenance" {
			errors = append(errors, fmt.Sprintf("ขออภัย อุปกรณ์ %s กำลังปิดปรับปรุงซ่อมบำรุง ไม่สามารถจองได้", eq.Name))
		} else {
			// Create booking with double-booking check
			_, err := models.CreateBooking(h.DB, u.ID, eqID, startTimeStr, endTimeStr, notes)
			if err != nil {
				errors = append(errors, err.Error())
			} else {
				middleware.SetFlash(w, r, "success", fmt.Sprintf("จองเข้าใช้งาน %s สำเร็จ! เวลา %s - %s น.",
					eq.Name, startTime.Format("15:04"), endTime.Format("15:04")))
				http.Redirect(w, r, "/my-bookings", http.StatusSeeOther)
				return
			}
		}
	}

	equipments, _ := models.GetAllEquipments(h.DB, 0, "", "")
	RenderTemplate(w, r, "booking.html", "จองเวลาใช้อุปกรณ์", "booking", map[string]interface{}{
		"Equipments":      equipments,
		"PreselectedEqID": eqID,
		"SelectedDate":    time.Now().Format("2006-01-02"),
		"MinDate":         time.Now().Format("2006-01-02"),
		"MaxDate":         time.Now().AddDate(0, 0, 14).Format("2006-01-02"),
		"Errors":          errors,
	})
}

func (h *Handler) ShowDashboard(w http.ResponseWriter, r *http.Request) {
	u := middleware.GetCurrentUser(r)
	if u.Role == "admin" {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	upcoming, err := models.GetUpcomingBookingsByUser(h.DB, u.ID, 5)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	allBookings, err := models.GetAllBookingsByUser(h.DB, u.ID, "all")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	upcomingCount := 0
	completedCount := 0
	cancelledCount := 0
	now := time.Now()

	for _, b := range allBookings {
		if b.Status == "cancelled" {
			cancelledCount++
		} else if b.EndTime.Before(now) {
			completedCount++
		} else {
			upcomingCount++
		}
	}

	favoriteEquipments, _ := models.GetUserFavoriteEquipments(h.DB, u.ID)

	RenderTemplate(w, r, "dashboard.html", "แดชบอร์ดสมาชิก", "dashboard", map[string]interface{}{
		"UpcomingBookings":   upcoming,
		"UpcomingCount":      upcomingCount,
		"CompletedCount":     completedCount,
		"CancelledCount":     cancelledCount,
		"FavoriteEquipments": favoriteEquipments,
	})
}

func (h *Handler) ShowMyBookings(w http.ResponseWriter, r *http.Request) {
	u := middleware.GetCurrentUser(r)
	statusFilter := r.URL.Query().Get("status")
	if statusFilter == "" {
		statusFilter = "all"
	}

	bookings, err := models.GetAllBookingsByUser(h.DB, u.ID, statusFilter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	RenderTemplate(w, r, "my_bookings.html", "ประวัติการจองของฉัน", "my-bookings", map[string]interface{}{
		"Bookings":     bookings,
		"FilterStatus": statusFilter,
	})
}

func (h *Handler) ProcessReportIssue(w http.ResponseWriter, r *http.Request) {
	u := middleware.GetCurrentUser(r)
	if u == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if !middleware.VerifyCSRF(r) {
		middleware.SetFlash(w, r, "danger", "Security check failed (CSRF)")
		http.Redirect(w, r, "/equipments", http.StatusSeeOther)
		return
	}

	eqID, _ := strconv.Atoi(r.FormValue("equipment_id"))
	desc := strings.TrimSpace(r.FormValue("issue_description"))

	if eqID <= 0 || desc == "" {
		middleware.SetFlash(w, r, "danger", "กรุณาระบุอาการชำรุดหรือปัญหาที่พบ")
	} else {
		err := models.CreateMaintenanceReport(h.DB, eqID, u.ID, desc)
		if err != nil {
			middleware.SetFlash(w, r, "danger", "ไม่สามารถส่งเรื่องแจ้งซ่อมได้: "+err.Error())
		} else {
			middleware.SetFlash(w, r, "success", "ส่งเรื่องแจ้งซ่อมเรียบร้อยแล้ว! เจ้าหน้าที่แอดมินจะเข้าตรวจเช็คโดยเร็วที่สุด")
		}
	}
	http.Redirect(w, r, "/equipments", http.StatusSeeOther)
}
