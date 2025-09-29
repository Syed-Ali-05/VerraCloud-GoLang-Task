package app

import (
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

"github.com/Syed-Ali-05/VerraCloud-GoLang-Task/internal/models"
"github.com/Syed-Ali-05/VerraCloud-GoLang-Task/internal/utils"

)

// session represents a simple in-memory session with an expiry.
type session struct {
	UserID uint
	Exp    time.Time
}

var sessions = map[string]session{}

const (
	cookieName = "sid"
	sessionTTL = 24 * time.Hour
)

// issueSession creates a session token, stores it in memory, and sets a cookie.
func issueSession(w http.ResponseWriter, userID uint) {
	tok := utils.RandomToken(32) // ← use utils.RandomToken
	sessions[tok] = session{UserID: userID, Exp: time.Now().Add(sessionTTL)}
	c := &http.Cookie{
		Name:     cookieName,
		Value:    tok,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	if isTLS() {
		c.Secure = true
	}
	http.SetCookie(w, c)
}

func clearSession(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(cookieName); err == nil {
		delete(sessions, c.Value)
		c.Value = ""
		c.Path = "/"
		c.MaxAge = -1
		c.HttpOnly = true
		c.SameSite = http.SameSiteLaxMode
		if isTLS() {
			c.Secure = true
		}
		http.SetCookie(w, c)
	}
}

// currentUser resolves the logged-in user from the session cookie; returns nil if missing/expired.
func (a *App) currentUser(r *http.Request) *models.User {
	c, err := r.Cookie(cookieName)
	if err != nil {
		log.Debug("No session cookie found")
		return nil
	}
	s, ok := sessions[c.Value]
	if !ok {
		log.Debug("Invalid session token")
		return nil
	}
	if time.Now().After(s.Exp) {
		log.Debug("Session expired")
		return nil
	}

	var u models.User
	if err := a.DB.First(&u, s.UserID).Error; err != nil {
		log.WithError(err).Warnf("Failed to load user for session %s", c.Value)
		return nil
	}

	return &u
}

func isTLS() bool { return os.Getenv("TLS") == "1" }
