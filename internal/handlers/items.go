// package handlers

// import (
// 	"html/template"
// 	"net/http"
// 	"time"

// 	"gorm.io/gorm"

// 	"go-htmx-auth-inline/internal/auth"
// 	"go-htmx-auth-inline/internal/models"
// )

// // ItemsHandler handles both GET (list items) and POST (create item) for the current user.
// // It returns an HTML fragment (items.html) suitable for HTMX swaps.
// func ItemsHandler(db *gorm.DB, tmpl *template.Template) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		// Require session
// 		userID, ok := auth.GetUserID(r)
// 		if !ok {
// 			w.WriteHeader(http.StatusUnauthorized)
// 			_, _ = w.Write([]byte(`<div class="contrast">Unauthorized. Please log in.</div>`))
// 			return
// 		}

// 		switch r.Method {
// 		case http.MethodGet:
// 			var items []models.Item
// 			_ = db.Where("user_id = ?", userID).Order("created_at DESC").Find(&items).Error
// 			_ = tmpl.ExecuteTemplate(w, "items.html", items)

// 		case http.MethodPost:
// 			name := r.FormValue("name")
// 			if name == "" {
// 				// Return current list with an inline error (minimal UX)
// 				var items []models.Item
// 				_ = db.Where("user_id = ?", userID).Order("created_at DESC").Find(&items).Error
// 				w.WriteHeader(http.StatusBadRequest)
// 				_ = tmpl.ExecuteTemplate(w, "items.html", items)
// 				return
// 			}

// 			_ = db.Create(&models.Item{
// 				UserID:    userID,
// 				Name:      name,
// 				CreatedAt: time.Now(),
// 			}).Error

// 			var items []models.Item
// 			_ = db.Where("user_id = ?", userID).Order("created_at DESC").Find(&items).Error
// 			_ = tmpl.ExecuteTemplate(w, "items.html", items)

// 		default:
// 			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
// 		}
// 	}
// }



package handlers

import (
  "html/template"
  "net/http"
  "time"

  "gorm.io/gorm"

  "go-htmx-auth-inline/internal/auth"
  "go-htmx-auth-inline/internal/models"
)

func ItemsHandler(db *gorm.DB, tmpl *template.Template) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    userID, ok := auth.GetUserID(r)
    if !ok {
      w.WriteHeader(http.StatusUnauthorized)
      _, _ = w.Write([]byte(`<div class="contrast">Unauthorized. Please log in.</div>`))
      return
    }

    switch r.Method {
    case http.MethodGet:
      var items []models.Item
      _ = db.Where("user_id = ?", userID).Order("created_at DESC").Find(&items).Error
      _ = tmpl.ExecuteTemplate(w, "items", items)

    case http.MethodPost:
      name := r.FormValue("name")
      if name == "" {
        w.WriteHeader(http.StatusBadRequest)
        var items []models.Item
        _ = db.Where("user_id = ?", userID).Order("created_at DESC").Find(&items).Error
        _ = tmpl.ExecuteTemplate(w, "items", items)
        return
      }
      _ = db.Create(&models.Item{
        UserID:    userID,
        Name:      name,
        CreatedAt: time.Now(),
      }).Error

      var items []models.Item
      _ = db.Where("user_id = ?", userID).Order("created_at DESC").Find(&items).Error
      _ = tmpl.ExecuteTemplate(w, "items", items)

    default:
      http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
    }
  }
}
