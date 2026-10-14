package middleware

import (
	"net/http"

	"gym-booking/internal/models"

	"github.com/gorilla/sessions"
)

var Store = sessions.NewCookieStore([]byte("gym-secret-key-32-chars-length!!"))

func init() {
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

const SessionName = "gym_session"

type contextKey string

const UserContextKey contextKey = "current_user"

type Flash struct {
	Type    string
	Message string
}

func SetFlash(w http.ResponseWriter, r *http.Request, flashType, message string) {
	session, _ := Store.Get(r, SessionName)
	session.AddFlash(Flash{Type: flashType, Message: message})
	_ = session.Save(r, w)
}

func GetFlash(w http.ResponseWriter, r *http.Request) *Flash {
	session, _ := Store.Get(r, SessionName)
	flashes := session.Flashes()
	if len(flashes) > 0 {
		_ = session.Save(r, w)
		if f, ok := flashes[0].(Flash); ok {
			return &f
		}
	}
	return nil
}

func GetCurrentUser(r *http.Request) *models.User {
	if u, ok := r.Context().Value(UserContextKey).(*models.User); ok {
		return u
	}
	return nil
}

func SetSessionUser(w http.ResponseWriter, r *http.Request, u *models.User) error {
	session, _ := Store.Get(r, SessionName)
	session.Values["user_id"] = u.ID
	session.Values["user_name"] = u.Name
	session.Values["user_email"] = u.Email
	session.Values["user_role"] = u.Role
	session.Values["user_phone"] = u.Phone
	return session.Save(r, w)
}

func ClearSession(w http.ResponseWriter, r *http.Request) error {
	session, _ := Store.Get(r, SessionName)
	session.Values = make(map[interface{}]interface{})
	session.Options.MaxAge = -1
	return session.Save(r, w)
}
