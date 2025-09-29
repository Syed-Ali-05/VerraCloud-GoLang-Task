package handlers

import (
	"html/template"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"go-htmx-auth-inline/internal/auth"
	"go-htmx-auth-inline/internal/models"
)

// LoginHandler handles POST login form and swaps content with dashboard
func LoginHandler(db *gorm.DB, tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")
		log.Printf("login attempt email=%q", email)

		var user models.User
		if err := db.Where("email = ?", email).First(&user).Error; err != nil {
			log.Printf("user not found: %v", err)
			_ = tmpl.ExecuteTemplate(w, "login", map[string]any{"Error": "Invalid credentials", "Email": email})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
			log.Printf("password mismatch for %s: %v", email, err)
			_ = tmpl.ExecuteTemplate(w, "login", map[string]any{"Error": "Invalid credentials", "Email": email})
			return
		}

		log.Printf("login success for %s (id=%d)", email, user.ID)
		auth.SetSession(w, user.ID)

		_ = tmpl.ExecuteTemplate(w, "dashboard", map[string]any{"Email": email})
	}
}

// LogoutHandler clears the session and returns login page
func LogoutHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		auth.ClearSession(w)
		if r.Header.Get("HX-Request") == "true" {
			_ = tmpl.ExecuteTemplate(w, "login", nil)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
