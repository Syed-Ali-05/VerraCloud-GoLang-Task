package auth

import (
    "fmt"
    "net/http"
)

const sessionCookie = "session_id"

// SetSession stores the userID in a cookie
func SetSession(w http.ResponseWriter, userID uint) {
    cookie := &http.Cookie{
        Name:     sessionCookie,
        Value:    fmt.Sprint(userID),
        HttpOnly: true,
        SameSite: http.SameSiteLaxMode,
        Path:     "/", // ensure cookie is available site-wide
    }
    http.SetCookie(w, cookie)
}

// ClearSession removes the cookie
func ClearSession(w http.ResponseWriter) {
    cookie := &http.Cookie{
        Name:     sessionCookie,
        Value:    "",
        MaxAge:   -1,
        HttpOnly: true,
        SameSite: http.SameSiteLaxMode,
        Path:     "/",
    }
    http.SetCookie(w, cookie)
}

// GetUserID reads userID from the cookie
func GetUserID(r *http.Request) (uint, bool) {
    c, err := r.Cookie(sessionCookie)
    if err != nil || c.Value == "" {
        return 0, false
    }
    var id uint
    _, err = fmt.Sscan(c.Value, &id)
    if err != nil {
        return 0, false
    }
    return id, true
}
