package main

import (
	"crypto/rand"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"
    log "github.com/sirupsen/logrus"

	"strconv"

	"golang.org/x/crypto/bcrypt"

	"gorm.io/gorm"
	"github.com/glebarez/sqlite" // PURE-GO SQLite driver (no CGO)
)

// ───────────────────────── Templates ─────────────────────────

//go:embed templates/*.templ
var tmplFS embed.FS

func mustParseSet(files ...string) *template.Template {
	t := template.New("").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
	})
	tt, err := t.ParseFS(tmplFS, files...)
	if err != nil {
		log.Fatalf("template parse: %v", err)
	}
	return tt
}


var (
    tmplBaseLogin     = mustParseSet("templates/base.templ", "templates/login.templ")
    tmplBaseDashboard = mustParseSet("templates/base.templ", "templates/dashboard.templ", "templates/items.templ")
)

var (
    tmplLoginPartial     = mustParseSet("templates/login.templ")
    tmplDashboardPartial = mustParseSet("templates/dashboard.templ")
    tmplItemsPartial     = mustParseSet("templates/items.templ")
)


// ───────────────────────── Models ─────────────────────────

type User struct {
	ID           uint      `gorm:"primaryKey"`
	Email        string    `gorm:"uniqueIndex;size:255;not null"`
	PasswordHash string    `gorm:"not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

type Item struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index;not null"`
	Name      string    `gorm:"size:255;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// ───────────────────────── Sessions ─────────────────────────

type session struct {
	UserID uint
	Exp    time.Time
}

var sessions = map[string]session{}

const (
	cookieName = "sid"
	sessionTTL = 24 * time.Hour

	adminEmail = "admin@example.com"
	adminPass  = "Passw0rd!"
)

// ───────────────────────── App ─────────────────────────

type App struct {
	DB *gorm.DB
}

func main() {

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(log.InfoLevel)

	// DB path (override with DB_PATH env)
	dbPath := strings.TrimSpace(os.Getenv("DB_PATH"))
	if dbPath == "" {
		dbPath = "app.db"
	}

	log.Infof("Starting server...")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&User{}, &Item{}); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}
	log.Infof("Database connected: %s", dbPath)

	

	// Seed admin user (bcrypt)
	var count int64
	db.Model(&User{}).Where("email = ?", adminEmail).Count(&count)
	if count == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte(adminPass), bcrypt.DefaultCost)
		if err := db.Create(&User{Email: adminEmail, PasswordHash: string(hash)}).Error; err != nil {
			log.Fatalf("seed user: %v", err)
		}
		log.Infof("Seeded admin user %s / %s", adminEmail, adminPass)
			}

	app := &App{DB: db}

	mux := http.NewServeMux()
	// Static (optional)
	mux.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))

	// Routes
	mux.HandleFunc("/", app.handleRoot) // GET
	mux.HandleFunc("/login", app.handleLogin)
	mux.HandleFunc("/logout", app.handleLogout)
	mux.HandleFunc("/items", app.handleItems)

	addr := ":8080"
	log.Infof("Listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, securityHeaders(mux)))
}

// ───────────────────────── Handlers ─────────────────────────

// GET /
func (a *App) handleRoot(w http.ResponseWriter, r *http.Request) {
	user := a.currentUser(r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if user == nil {
		if err := tmplBaseLogin.ExecuteTemplate(w, "base.html", map[string]any{
			"Email": "",
			"Error": "",
		}); err != nil {
			httpErrorFragment(w, err)
		}
		return
	}

	if err := tmplBaseDashboard.ExecuteTemplate(w, "base.html", map[string]any{
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

	var u User
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

	// success → issue cookie + show dashboard
	issueSession(w, u.ID)
	log.WithFields(log.Fields{"email": u.Email, "user_id": u.ID}).Info("Login success")

	if err := tmplDashboardPartial.ExecuteTemplate(w, "dashboard.html", map[string]any{"User": &u}); err != nil {
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
// func (a *App) handleItems(w http.ResponseWriter, r *http.Request) {
// 	user := a.currentUser(r)
// 	if user == nil {
// 		w.WriteHeader(http.StatusUnauthorized)
// 		fmt.Fprint(w, `<div class="notice">Unauthorized. Please log in.</div>`)
// 		// w.Header().Set("Content-Type", "text/html; charset=utf-8")
// 		// fmt.Fprint(w, `<div class="notice">Unauthorized. Please log in.</div>`)
// 		return
// 	}

// 	switch r.Method {
// 	case http.MethodGet:
// 		var items []Item
// 		if err := a.DB.Order("created_at DESC").Where("user_id = ?", user.ID).Find(&items).Error; err != nil {
// 			httpErrorFragment(w, err)
// 			return
// 		}
// 		if err := tmplItemsPartial.ExecuteTemplate(w, "items.html", map[string]any{"Items": items}); err != nil {
// 			httpErrorFragment(w, err)
// 		}
// 	case http.MethodPost:
// 		name := strings.TrimSpace(r.FormValue("name"))
// 		if name == "" {
// 			w.WriteHeader(http.StatusBadRequest)
// 			fmt.Fprint(w, `<div id="item-list"><div class="error">Name is required.</div></div>`)
// 			return
// 		}
// 		it := Item{UserID: user.ID, Name: name}
// 		if err := a.DB.Create(&it).Error; err != nil {
// 			httpErrorFragment(w, err)
// 			return
// 		}
// 		var items []Item
// 		if err := a.DB.Order("created_at DESC").Where("user_id = ?", user.ID).Find(&items).Error; err != nil {
// 			httpErrorFragment(w, err)
// 			return
// 		}
// 		if err := tmplItemsPartial.ExecuteTemplate(w, "items.html", map[string]any{"Items": items}); err != nil {
// 			httpErrorFragment(w, err)
// 		}
// 	default:
// 		http.NotFound(w, r)
// 	}
// }
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
	// also check form values for POST
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
		// just render list
	case http.MethodPost:
		// handle add new item
		name := strings.TrimSpace(r.FormValue("name"))
		if name == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `<div id="item-list"><div class="error">Name is required.</div></div>`)
			return
		}
		it := Item{UserID: user.ID, Name: name}
		if err := a.DB.Create(&it).Error; err != nil {
			httpErrorFragment(w, err)
			return
		}
		log.WithFields(log.Fields{
			"user_id": user.ID,
			"name":    name,
		}).Info("Adding new item")
		// reset to first page after adding
		page = 1
	default:
		http.NotFound(w, r)
		return
	}

	// Common query logic for GET and POST
	var items []Item
	query := a.DB.Where("user_id = ?", user.ID)
	if q != "" {
		query = query.Where("name LIKE ?", "%"+q+"%")
	}

	var total int64
	query.Model(&Item{}).Count(&total)

	if err := query.Order("created_at DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&items).Error; err != nil {
		httpErrorFragment(w, err)
		return
	}

	totalPages := int((total + pageSize - 1) / pageSize)
	if err := tmplItemsPartial.ExecuteTemplate(w, "items.html", map[string]any{
		"Items":      items,
		"Page":       page,
		"TotalPages": totalPages,
		"Q":          q,
	}); err != nil {
		httpErrorFragment(w, err)
	}
}



// ───────────────────────── Helpers ─────────────────────────

func renderLoginPartial(w http.ResponseWriter, email, errMsg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmplLoginPartial.ExecuteTemplate(w, "login.html", map[string]any{
		"Email": email,
		"Error": errMsg,
	}); err != nil {
		httpErrorFragment(w, err)
	}
}

func httpErrorFragment(w http.ResponseWriter, err error) {
	log.Errorf("error: %v", err) // logrus error-level logging
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<div class="error">Something went wrong. Please try again.</div>`)
}


func issueSession(w http.ResponseWriter, userID uint) {
	tok := randomToken(32)
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

// func (a *App) currentUser(r *http.Request) *User {
// 	c, err := r.Cookie(cookieName)
// 	if err != nil {
// 		return nil
// 	}
// 	s, ok := sessions[c.Value]
// 	if !ok || time.Now().After(s.Exp) {
// 		return nil
// 	}
// 	var u User
// 	if err := a.DB.First(&u, s.UserID).Error; err != nil {
// 		return nil
// 	}
// 	return &u
// }

func (a *App) currentUser(r *http.Request) *User {
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

	var u User
	if err := a.DB.First(&u, s.UserID).Error; err != nil {
		log.WithError(err).Warnf("Failed to load user for session %s", c.Value)
		return nil
	}

	log.WithFields(log.Fields{
		"user_id": u.ID,
		"email":   u.Email,
	}).Debug("Resolved current user from session")

	return &u
}


func randomToken(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	const alpha = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for i := range b {
		b[i] = alpha[int(b[i])%len(alpha)]
	}
	return string(b)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

func isTLS() bool { return os.Getenv("TLS") == "1" }

// Ensure database/sql is linked
var _ = sql.ErrNoRows
var _ = errors.New
