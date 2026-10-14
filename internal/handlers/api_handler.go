package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"gym-booking/internal/middleware"
	"gym-booking/internal/models"
)

type SlotInfo struct {
	ID        int    `json:"id"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Status    string `json:"status"`
}

type AvailabilityResponse struct {
	Success     bool              `json:"success"`
	Message     string            `json:"message,omitempty"`
	Date        string            `json:"date,omitempty"`
	Equipment   *models.Equipment `json:"equipment,omitempty"`
	BookedSlots []SlotInfo        `json:"booked_slots,omitempty"`
}

func (h *Handler) GetAvailabilityAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	eqID, _ := strconv.Atoi(r.URL.Query().Get("equipment_id"))
	date := r.URL.Query().Get("date")

	if eqID <= 0 || date == "" {
		_ = json.NewEncoder(w).Encode(AvailabilityResponse{
			Success: false,
			Message: "Missing equipment_id or date parameter",
		})
		return
	}

	// Validate date format YYYY-MM-DD
	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		_ = json.NewEncoder(w).Encode(AvailabilityResponse{
			Success: false,
			Message: "Invalid date format. Expected YYYY-MM-DD.",
		})
		return
	}

	equipment, err := models.GetEquipmentByID(h.DB, eqID)
	if err != nil || equipment == nil {
		_ = json.NewEncoder(w).Encode(AvailabilityResponse{
			Success: false,
			Message: "Equipment not found",
		})
		return
	}

	bookedSlots, err := models.GetBookedSlotsForDate(h.DB, eqID, date)
	if err != nil {
		_ = json.NewEncoder(w).Encode(AvailabilityResponse{
			Success: false,
			Message: "Failed to retrieve booked slots: " + err.Error(),
		})
		return
	}

	slotInfos := make([]SlotInfo, 0, len(bookedSlots))
	for _, b := range bookedSlots {
		slotInfos = append(slotInfos, SlotInfo{
			ID:        b.ID,
			StartTime: b.StartTime.Format("2006-01-02 15:04:05"),
			EndTime:   b.EndTime.Format("2006-01-02 15:04:05"),
			Status:    b.Status,
		})
	}

	_ = json.NewEncoder(w).Encode(AvailabilityResponse{
		Success:     true,
		Date:        date,
		Equipment:   equipment,
		BookedSlots: slotInfos,
	})
}

func (h *Handler) CancelBookingAPI(w http.ResponseWriter, r *http.Request) {
	u := middleware.GetCurrentUser(r)
	if u == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if !middleware.VerifyCSRF(r) {
		middleware.SetFlash(w, r, "danger", "Security check failed (Invalid CSRF Token)")
		http.Redirect(w, r, "/my-bookings", http.StatusSeeOther)
		return
	}

	bookingID, _ := strconv.Atoi(r.FormValue("booking_id"))
	reason := r.FormValue("cancel_reason")
	isAdminRequest := (r.FormValue("is_admin_request") == "1")

	if bookingID <= 0 {
		middleware.SetFlash(w, r, "danger", "ไม่พบรหัสการจอง")
		http.Redirect(w, r, "/my-bookings", http.StatusSeeOther)
		return
	}

	isAdmin := (u.Role == "admin")
	err := models.CancelBooking(h.DB, bookingID, u.ID, isAdmin, reason)
	if err != nil {
		middleware.SetFlash(w, r, "danger", "เกิดข้อผิดพลาดในการยกเลิก: "+err.Error())
	} else {
		middleware.SetFlash(w, r, "success", "ยกเลิกรายการจองเรียบร้อยแล้ว")
	}

	if isAdmin && isAdminRequest {
		http.Redirect(w, r, "/admin/bookings", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/my-bookings", http.StatusSeeOther)
	}
}
