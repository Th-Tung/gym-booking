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

func (h *Handler) ShowAdminDashboard(w http.ResponseWriter, r *http.Request) {
	stats, err := models.GetDashboardStats(h.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	today := time.Now().Format("2006-01-02")
	todayBookings, err := models.GetAllBookings(h.DB, today, 0, "all")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	popularEquipments, _ := models.GetPopularEquipments(h.DB)
	maintenanceList, _ := models.GetAllEquipments(h.DB, 0, "maintenance", "")

	RenderTemplate(w, r, "admin/index.html", "แดชบอร์ดภาพรวม", "admin", map[string]interface{}{
		"Stats":               stats,
		"TodayBookings":       todayBookings,
		"PopularEquipments":   popularEquipments,
		"MaintenanceList":     maintenanceList,
		"PendingReportsCount": models.GetPendingReportsCount(h.DB),
	})
}

func (h *Handler) ShowAdminEquipments(w http.ResponseWriter, r *http.Request) {
	equipments, err := models.GetAllEquipments(h.DB, 0, "", "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	categories, err := models.GetAllCategories(h.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	RenderTemplate(w, r, "admin/equipments.html", "จัดการเครื่องเล่นและอุปกรณ์", "admin", map[string]interface{}{
		"Equipments": equipments,
		"Categories": categories,
	})
}

func (h *Handler) ProcessAddEquipment(w http.ResponseWriter, r *http.Request) {
	if !middleware.VerifyCSRF(r) {
		middleware.SetFlash(w, r, "danger", "CSRF verification failed")
		http.Redirect(w, r, "/admin/equipments", http.StatusSeeOther)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	catID, _ := strconv.Atoi(r.FormValue("category_id"))
	description := strings.TrimSpace(r.FormValue("description"))
	imagePath := strings.TrimSpace(r.FormValue("image_path"))
	status := r.FormValue("status")

	if name == "" || catID <= 0 {
		middleware.SetFlash(w, r, "danger", "กรุณากรอกชื่ออุปกรณ์และเลือกหมวดหมู่")
	} else {
		if imagePath == "" {
			imagePath = "assets/images/treadmill.svg"
		}
		if status != "maintenance" {
			status = "available"
		}
		err := models.CreateEquipment(h.DB, name, catID, description, imagePath, status)
		if err != nil {
			middleware.SetFlash(w, r, "danger", "ไม่สามารถเพิ่มอุปกรณ์ได้: "+err.Error())
		} else {
			middleware.SetFlash(w, r, "success", "เพิ่มอุปกรณ์ "+name+" เรียบร้อยแล้ว")
		}
	}
	http.Redirect(w, r, "/admin/equipments", http.StatusSeeOther)
}

func (h *Handler) ProcessEditEquipment(w http.ResponseWriter, r *http.Request) {
	if !middleware.VerifyCSRF(r) {
		middleware.SetFlash(w, r, "danger", "CSRF verification failed")
		http.Redirect(w, r, "/admin/equipments", http.StatusSeeOther)
		return
	}

	id, _ := strconv.Atoi(r.FormValue("id"))
	name := strings.TrimSpace(r.FormValue("name"))
	catID, _ := strconv.Atoi(r.FormValue("category_id"))
	description := strings.TrimSpace(r.FormValue("description"))
	imagePath := strings.TrimSpace(r.FormValue("image_path"))
	status := r.FormValue("status")

	if id <= 0 || name == "" || catID <= 0 {
		middleware.SetFlash(w, r, "danger", "ข้อมูลไม่ครบถ้วน")
	} else {
		err := models.UpdateEquipment(h.DB, id, name, catID, description, imagePath, status)
		if err != nil {
			middleware.SetFlash(w, r, "danger", "ไม่สามารถแก้ไขได้: "+err.Error())
		} else {
			middleware.SetFlash(w, r, "success", "แก้ไขข้อมูลอุปกรณ์เรียบร้อยแล้ว")
		}
	}
	http.Redirect(w, r, "/admin/equipments", http.StatusSeeOther)
}

func (h *Handler) ProcessToggleEquipment(w http.ResponseWriter, r *http.Request) {
	if !middleware.VerifyCSRF(r) {
		middleware.SetFlash(w, r, "danger", "CSRF verification failed")
		http.Redirect(w, r, "/admin/equipments", http.StatusSeeOther)
		return
	}

	id, _ := strconv.Atoi(r.FormValue("id"))
	if id > 0 {
		newStatus, err := models.ToggleEquipmentStatus(h.DB, id)
		if err != nil {
			middleware.SetFlash(w, r, "danger", "เกิดข้อผิดพลาด: "+err.Error())
		} else {
			statusThai := "พร้อมใช้งาน"
			if newStatus == "maintenance" {
				statusThai = "ส่งซ่อมบำรุง"
			}
			middleware.SetFlash(w, r, "success", fmt.Sprintf("เปลี่ยนสถานะอุปกรณ์เป็น %s แล้ว", statusThai))
		}
	}
	http.Redirect(w, r, "/admin/equipments", http.StatusSeeOther)
}

func (h *Handler) ProcessDeleteEquipment(w http.ResponseWriter, r *http.Request) {
	if !middleware.VerifyCSRF(r) {
		middleware.SetFlash(w, r, "danger", "CSRF verification failed")
		http.Redirect(w, r, "/admin/equipments", http.StatusSeeOther)
		return
	}

	id, _ := strconv.Atoi(r.FormValue("id"))
	if id > 0 {
		err := models.DeleteEquipment(h.DB, id)
		if err != nil {
			middleware.SetFlash(w, r, "danger", "ไม่สามารถลบอุปกรณ์ได้: "+err.Error())
		} else {
			middleware.SetFlash(w, r, "success", "ลบอุปกรณ์ออกจากระบบแล้ว")
		}
	}
	http.Redirect(w, r, "/admin/equipments", http.StatusSeeOther)
}

func (h *Handler) ShowAdminCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := models.GetAllCategories(h.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	RenderTemplate(w, r, "admin/categories.html", "จัดการหมวดหมู่อุปกรณ์", "admin", map[string]interface{}{
		"Categories": categories,
	})
}

func (h *Handler) ProcessAddCategory(w http.ResponseWriter, r *http.Request) {
	if !middleware.VerifyCSRF(r) {
		middleware.SetFlash(w, r, "danger", "CSRF verification failed")
		http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	description := strings.TrimSpace(r.FormValue("description"))
	icon := strings.TrimSpace(r.FormValue("icon"))

	if name == "" {
		middleware.SetFlash(w, r, "danger", "กรุณาระบุชื่อหมวดหมู่")
	} else {
		err := models.CreateCategory(h.DB, name, description, icon)
		if err != nil {
			middleware.SetFlash(w, r, "danger", err.Error())
		} else {
			middleware.SetFlash(w, r, "success", "เพิ่มหมวดหมู่ "+name+" เรียบร้อยแล้ว")
		}
	}
	http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
}

func (h *Handler) ProcessEditCategory(w http.ResponseWriter, r *http.Request) {
	if !middleware.VerifyCSRF(r) {
		middleware.SetFlash(w, r, "danger", "CSRF verification failed")
		http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
		return
	}

	id, _ := strconv.Atoi(r.FormValue("id"))
	name := strings.TrimSpace(r.FormValue("name"))
	description := strings.TrimSpace(r.FormValue("description"))
	icon := strings.TrimSpace(r.FormValue("icon"))

	if id <= 0 || name == "" {
		middleware.SetFlash(w, r, "danger", "ข้อมูลไม่ถูกต้อง")
	} else {
		err := models.UpdateCategory(h.DB, id, name, description, icon)
		if err != nil {
			middleware.SetFlash(w, r, "danger", err.Error())
		} else {
			middleware.SetFlash(w, r, "success", "แก้ไขหมวดหมู่เรียบร้อยแล้ว")
		}
	}
	http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
}

func (h *Handler) ProcessDeleteCategory(w http.ResponseWriter, r *http.Request) {
	if !middleware.VerifyCSRF(r) {
		middleware.SetFlash(w, r, "danger", "CSRF verification failed")
		http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
		return
	}

	id, _ := strconv.Atoi(r.FormValue("id"))
	if id > 0 {
		err := models.DeleteCategory(h.DB, id)
		if err != nil {
			middleware.SetFlash(w, r, "danger", err.Error())
		} else {
			middleware.SetFlash(w, r, "success", "ลบหมวดหมู่เรียบร้อยแล้ว")
		}
	}
	http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
}

func (h *Handler) ShowAdminBookings(w http.ResponseWriter, r *http.Request) {
	dateFilter := r.URL.Query().Get("date")
	eqID, _ := strconv.Atoi(r.URL.Query().Get("equipment_id"))
	statusFilter := r.URL.Query().Get("status")
	if statusFilter == "" {
		statusFilter = "all"
	}

	equipments, _ := models.GetAllEquipments(h.DB, 0, "", "")
	bookings, err := models.GetAllBookings(h.DB, dateFilter, eqID, statusFilter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	RenderTemplate(w, r, "admin/bookings.html", "ตารางการจองอุปกรณ์ทั้งหมด", "admin", map[string]interface{}{
		"Bookings":        bookings,
		"Equipments":      equipments,
		"FilterDate":      dateFilter,
		"FilterEquipment": eqID,
		"FilterStatus":    statusFilter,
	})
}

func (h *Handler) ShowAdminUsers(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	users, err := models.GetAllUsers(h.DB, query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	RenderTemplate(w, r, "admin/users.html", "รายชื่อสมาชิกและสิทธิ์", "admin", map[string]interface{}{
		"Users":       users,
		"SearchQuery": query,
	})
}

func (h *Handler) ProcessUpdateUserRole(w http.ResponseWriter, r *http.Request) {
	currentAdmin := middleware.GetCurrentUser(r)
	if !middleware.VerifyCSRF(r) {
		middleware.SetFlash(w, r, "danger", "CSRF verification failed")
		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
		return
	}

	targetID, _ := strconv.Atoi(r.FormValue("user_id"))
	newRole := r.FormValue("role")

	if targetID == currentAdmin.ID && newRole != "admin" {
		middleware.SetFlash(w, r, "danger", "คุณไม่สามารถลดระดับสิทธิ์ของบัญชีตัวเองได้")
	} else if targetID > 0 && (newRole == "admin" || newRole == "member") {
		err := models.UpdateUserRole(h.DB, targetID, newRole)
		if err != nil {
			middleware.SetFlash(w, r, "danger", err.Error())
		} else {
			middleware.SetFlash(w, r, "success", "อัปเดตสิทธิ์ผู้ใช้เรียบร้อยแล้ว")
		}
	}
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (h *Handler) ShowAdminReports(w http.ResponseWriter, r *http.Request) {
	statusFilter := r.URL.Query().Get("status")
	if statusFilter == "" {
		statusFilter = "all"
	}

	reports, err := models.GetAllMaintenanceReports(h.DB, statusFilter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pendingCount := models.GetPendingReportsCount(h.DB)

	RenderTemplate(w, r, "admin/reports.html", "รายการแจ้งซ่อมอุปกรณ์", "admin", map[string]interface{}{
		"Reports":      reports,
		"FilterStatus": statusFilter,
		"PendingCount": pendingCount,
	})
}

func (h *Handler) ProcessConfirmReport(w http.ResponseWriter, r *http.Request) {
	if !middleware.VerifyCSRF(r) {
		middleware.SetFlash(w, r, "danger", "CSRF verification failed")
		http.Redirect(w, r, "/admin/reports", http.StatusSeeOther)
		return
	}

	reportID, _ := strconv.Atoi(r.FormValue("report_id"))
	adminNotes := strings.TrimSpace(r.FormValue("admin_notes"))

	if reportID <= 0 {
		middleware.SetFlash(w, r, "danger", "ไม่พบรหัสรายการแจ้งซ่อม")
		http.Redirect(w, r, "/admin/reports", http.StatusSeeOther)
		return
	}

	err := models.ConfirmMaintenanceReport(h.DB, reportID, adminNotes)
	if err != nil {
		middleware.SetFlash(w, r, "danger", "เกิดข้อผิดพลาดในการยืนยันแจ้งซ่อม: "+err.Error())
	} else {
		middleware.SetFlash(w, r, "warning", "ยืนยันรับเรื่องแล้ว! อุปกรณ์ถูกปรับเป็นสถานะ 'ส่งซ่อมบำรุง' ชั่วคราว และระบบได้ยกเลิกคิวที่จองไว้เรียบร้อยแล้ว")
	}
	http.Redirect(w, r, "/admin/reports", http.StatusSeeOther)
}

func (h *Handler) ProcessResolveReport(w http.ResponseWriter, r *http.Request) {
	if !middleware.VerifyCSRF(r) {
		middleware.SetFlash(w, r, "danger", "CSRF verification failed")
		http.Redirect(w, r, "/admin/reports", http.StatusSeeOther)
		return
	}

	reportID, _ := strconv.Atoi(r.FormValue("report_id"))
	adminNotes := strings.TrimSpace(r.FormValue("admin_notes"))

	if reportID <= 0 {
		middleware.SetFlash(w, r, "danger", "ไม่พบรหัสรายการแจ้งซ่อม")
		http.Redirect(w, r, "/admin/reports", http.StatusSeeOther)
		return
	}

	err := models.ResolveMaintenanceReport(h.DB, reportID, adminNotes)
	if err != nil {
		middleware.SetFlash(w, r, "danger", "เกิดข้อผิดพลาดในการบันทึก: "+err.Error())
	} else {
		middleware.SetFlash(w, r, "success", "บันทึกซ่อมเสร็จสิ้นเรียบร้อยแล้ว! อุปกรณ์กลับมาเปิดให้บริการพร้อมใช้งานตามปกติ")
	}
	http.Redirect(w, r, "/admin/reports", http.StatusSeeOther)
}

func (h *Handler) ProcessRejectReport(w http.ResponseWriter, r *http.Request) {
	if !middleware.VerifyCSRF(r) {
		middleware.SetFlash(w, r, "danger", "CSRF verification failed")
		http.Redirect(w, r, "/admin/reports", http.StatusSeeOther)
		return
	}

	reportID, _ := strconv.Atoi(r.FormValue("report_id"))
	adminNotes := strings.TrimSpace(r.FormValue("admin_notes"))

	if reportID <= 0 {
		middleware.SetFlash(w, r, "danger", "ไม่พบรหัสรายการแจ้งซ่อม")
		http.Redirect(w, r, "/admin/reports", http.StatusSeeOther)
		return
	}

	err := models.RejectMaintenanceReport(h.DB, reportID, adminNotes)
	if err != nil {
		middleware.SetFlash(w, r, "danger", "เกิดข้อผิดพลาด: "+err.Error())
	} else {
		middleware.SetFlash(w, r, "info", "ปฏิเสธเรื่องแจ้งซ่อมแล้ว (อุปกรณ์ยังเปิดให้บริการตามปกติ)")
	}
	http.Redirect(w, r, "/admin/reports", http.StatusSeeOther)
}

