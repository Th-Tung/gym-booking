package handlers

import (
	"net/http"
	"strings"

	"gym-booking/internal/middleware"
	"gym-booking/internal/models"
)

func (h *Handler) ShowRegister(w http.ResponseWriter, r *http.Request) {
	if middleware.GetCurrentUser(r) != nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	RenderTemplate(w, r, "register.html", "สมัครสมาชิกใหม่", "register", map[string]interface{}{
		"Name":   "",
		"Email":  "",
		"Phone":  "",
		"Errors": []string{},
	})
}

func (h *Handler) ProcessRegister(w http.ResponseWriter, r *http.Request) {
	if !middleware.VerifyCSRF(r) {
		middleware.SetFlash(w, r, "danger", "Session หมดอายุหรือไม่ถูกต้อง กรุณาลองใหม่อีกครั้ง")
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	phone := strings.TrimSpace(r.FormValue("phone"))
	password := r.FormValue("password")
	passwordConfirm := r.FormValue("password_confirm")

	var errors []string
	if name == "" {
		errors = append(errors, "กรุณาระบุชื่อ-นามสกุล")
	}
	if email == "" || !strings.Contains(email, "@") {
		errors = append(errors, "กรุณาระบุอีเมลที่ถูกต้อง")
	}
	if len(password) < 6 {
		errors = append(errors, "รหัสผ่านต้องมีความยาวอย่างน้อย 6 ตัวอักษร")
	}
	if password != passwordConfirm {
		errors = append(errors, "รหัสผ่านยืนยันไม่ตรงกัน")
	}

	if len(errors) == 0 {
		existing, err := models.GetUserByEmail(h.DB, email)
		if err != nil {
			errors = append(errors, "เกิดข้อผิดพลาดในการตรวจสอบระบบ: "+err.Error())
		} else if existing != nil {
			errors = append(errors, "อีเมลนี้ถูกใช้งานแล้วในระบบ กรุณาใช้อีเมลอื่นหรือเข้าสู่ระบบ")
		} else {
			_, err := models.CreateUser(h.DB, name, email, password, phone)
			if err != nil {
				errors = append(errors, "ไม่สามารถสร้างบัญชีได้: "+err.Error())
			} else {
				middleware.SetFlash(w, r, "success", "สมัครสมาชิกสำเร็จ! คุณสามารถเข้าสู่ระบบเพื่อเริ่มจองอุปกรณ์ได้ทันที")
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
		}
	}

	RenderTemplate(w, r, "register.html", "สมัครสมาชิกใหม่", "register", map[string]interface{}{
		"Name":   name,
		"Email":  email,
		"Phone":  phone,
		"Errors": errors,
	})
}

func (h *Handler) ShowLogin(w http.ResponseWriter, r *http.Request) {
	u := middleware.GetCurrentUser(r)
	if u != nil {
		if u.Role == "admin" {
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		}
		return
	}

	RenderTemplate(w, r, "login.html", "เข้าสู่ระบบ", "login", map[string]interface{}{
		"Email": "",
		"Error": "",
	})
}

func (h *Handler) ProcessLogin(w http.ResponseWriter, r *http.Request) {
	if !middleware.VerifyCSRF(r) {
		middleware.SetFlash(w, r, "danger", "Session หมดอายุหรือไม่ถูกต้อง กรุณาลองใหม่อีกครั้ง")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	password := r.FormValue("password")

	if email == "" || password == "" {
		RenderTemplate(w, r, "login.html", "เข้าสู่ระบบ", "login", map[string]interface{}{
			"Email": email,
			"Error": "กรุณากรอกทั้งอีเมลและรหัสผ่าน",
		})
		return
	}

	u, err := models.GetUserByEmail(h.DB, email)
	if err != nil || u == nil || !u.CheckPassword(password) {
		RenderTemplate(w, r, "login.html", "เข้าสู่ระบบ", "login", map[string]interface{}{
			"Email": email,
			"Error": "อีเมลหรือรหัสผ่านไม่ถูกต้อง",
		})
		return
	}

	_ = middleware.SetSessionUser(w, r, u)
	middleware.SetFlash(w, r, "success", "ยินดีต้อนรับกลับคุณ "+u.Name+"!")

	if u.Role == "admin" {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

func (h *Handler) ProcessLogout(w http.ResponseWriter, r *http.Request) {
	_ = middleware.ClearSession(w, r)
	middleware.SetFlash(w, r, "info", "คุณได้ออกจากระบบเรียบร้อยแล้ว (You have logged out)")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

