package web

import (
	"net/http"
	"os"
	"time"
)

var AuthCookieName = "X-academic-sesi"

func GetCookieConfig(value string, expiresAt time.Time) http.Cookie {
	return http.Cookie{
		Name:     AuthCookieName,
		Value:    value,
		Expires:  expiresAt,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Secure:   os.Getenv("APP_ENV") == "production",
		HttpOnly: true,
	}

}

// GetClearCookieConfig returns a cookie config that instructs the browser to
// remove the auth cookie immediately (MaxAge=-1).
func GetClearCookieConfig() http.Cookie {
	return http.Cookie{
		Name:     AuthCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
		Secure:   os.Getenv("APP_ENV") == "production",
		HttpOnly: true,
	}
}
