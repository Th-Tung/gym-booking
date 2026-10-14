package middleware

import (
	"context"
	"database/sql"
	"net/http"

	"gym-booking/internal/models"
)

// InjectUser middleware inspects the session and loads current user into context
func InjectUser(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, _ := Store.Get(r, SessionName)
			userID, ok := session.Values["user_id"].(int)
			if ok && userID > 0 {
				u, err := models.GetUserByID(db, userID)
				if err == nil && u != nil {
					ctx := context.WithValue(r.Context(), UserContextKey, u)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAuth middleware protects routes requiring logged in status
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := GetCurrentUser(r)
		if u == nil {
			SetFlash(w, r, "warning", "กรุณาเข้าสู่ระบบก่อนดำเนินการต่อ (Please log in first)")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

// RequireAdmin middleware protects admin-only routes
func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := GetCurrentUser(r)
		if u == nil {
			SetFlash(w, r, "warning", "กรุณาเข้าสู่ระบบก่อนดำเนินการต่อ")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if u.Role != "admin" {
			SetFlash(w, r, "danger", "คุณไม่มีสิทธิ์เข้าถึงหน้านี้ (Admin only)")
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}
