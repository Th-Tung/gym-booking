package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

func GenerateCSRFToken(w http.ResponseWriter, r *http.Request) string {
	session, _ := Store.Get(r, SessionName)
	token, ok := session.Values["csrf_token"].(string)
	if !ok || token == "" {
		b := make([]byte, 32)
		_, _ = rand.Read(b)
		token = hex.EncodeToString(b)
		session.Values["csrf_token"] = token
		_ = session.Save(r, w)
	}
	return token
}

func VerifyCSRF(r *http.Request) bool {
	session, _ := Store.Get(r, SessionName)
	expected, ok := session.Values["csrf_token"].(string)
	if !ok || expected == "" {
		return false
	}
	submitted := r.FormValue("csrf_token")
	if submitted == "" {
		submitted = r.Header.Get("X-CSRF-Token")
	}
	return expected == submitted
}
