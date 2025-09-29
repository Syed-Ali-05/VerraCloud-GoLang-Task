package app

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

"github.com/Syed-Ali-05/VerraCloud-GoLang-Task/internal/models"
)

// GET /
func (a *App) handleRoot(w http.ResponseWriter, r *http.Request) {
	user := a.currentUser(r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if user == nil {
		if err := tmplBaseLogin.ExecuteTemplate(w, "base.templ", map[string]any{
			"Email": "",
			"Error": "",
		}); err != nil {
			httpErrorFragment(w, err)
		}
		return
	}

	if err := tmplBaseDashboard.ExecuteTemplate(w, "base.templ", map[string]any{
		"User": user,
	}); err != nil {
		httpErrorFragment(w, err)
	}
}

// POST /login
func (a *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	var u models.User
	if err := a.DB.First(&u, "email = ?", email).Error; err != nil {
		log.WithField("email", email).Warn("Login failed: user not found")
		renderLoginPartial(w, email, "Invalid email or password.")
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		log.WithField("email", email).Warn("Login failed: wrong password")
		renderLoginPartial(w, email, "Invalid email or password.")
		return
	}

	issueSession(w, u.ID)
	log.WithFields(log.Fields{"email": u.Email, "user_id": u.ID}).Info("Login success")

	if err := tmplDashboardPartial.ExecuteTemplate(w, "dashboard.templ", map[string]any{"User": &u}); err != nil {
		httpErrorFragment(w, err)
	}
}

// POST /logout
func (a *App) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	log.WithField("user", a.currentUser(r)).Info("User logging out")
	clearSession(w, r)
	renderLoginPartial(w, "", "")
}

// GET/POST /items
func (a *App) handleItems(w http.ResponseWriter, r *http.Request) {
	user := a.currentUser(r)
	if user == nil {
		log.Warn("Unauthorized /items access")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `<div class="notice">Unauthorized. Please log in.</div>`)
		return
	}

	const pageSize = 5
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	q := strings.TrimSpace(r.FormValue("q"))
	if q == "" {
		q = strings.TrimSpace(r.URL.Query().Get("q"))
	}

	log.WithFields(log.Fields{
		"user_id": user.ID,
		"method":  r.Method,
		"q":       q,
		"page":    page,
	}).Info("/items request")

	switch r.Method {
	case http.MethodGet:
		// render list
	case http.MethodPost:
		name := strings.TrimSpace(r.FormValue("name"))
		if name == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `<div id="item-list"><div class="error">Name is required.</div></div>`)
			return
		}
		it := models.Item{UserID: user.ID, Name: name}
		if err := a.DB.Create(&it).Error; err != nil {
			httpErrorFragment(w, err)
			return
		}
		log.WithFields(log.Fields{"user_id": user.ID, "name": name}).Info("Added new item")
		page = 1
	default:
		http.NotFound(w, r)
		return
	}

	var items []models.Item
	query := a.DB.Where("user_id = ?", user.ID)
	if q != "" {
		query = query.Where("name LIKE ?", "%"+q+"%")
	}

	var total int64
	query.Model(&models.Item{}).Count(&total)

	if err := query.Order("created_at DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&items).Error; err != nil {
		httpErrorFragment(w, err)
		return
	}

	totalPages := int((total + pageSize - 1) / pageSize)
	if err := tmplItemsPartial.ExecuteTemplate(w, "items.templ", map[string]any{
		"Items":      items,
		"Page":       page,
		"TotalPages": totalPages,
		"Q":          q,
	}); err != nil {
		httpErrorFragment(w, err)
	}
}
