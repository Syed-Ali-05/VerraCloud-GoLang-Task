package app

import (
	"net/http"
	"os"
	"gorm.io/gorm"


	log "github.com/sirupsen/logrus"

"github.com/Syed-Ali-05/VerraCloud-GoLang-Task/internal/models"
"github.com/Syed-Ali-05/VerraCloud-GoLang-Task/internal/utils"

)

type App struct {
    DB *gorm.DB
}

// Run bootstraps DB, routes, and starts the server.
func Run(addr string) error {
	if addr == "" {
		addr = ":8080"
	}
	if envAddr := os.Getenv("ADDR"); envAddr != "" {
		addr = envAddr
	}

	// DB init (AutoMigrate + seed inside)
	db := models.InitDB()
	a := &App{DB: db}

	mux := http.NewServeMux()

	// Static (optional)
	mux.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))

	// Routes
	mux.HandleFunc("/", a.handleRoot)
	mux.HandleFunc("/login", a.handleLogin)
	mux.HandleFunc("/logout", a.handleLogout)
	mux.HandleFunc("/items", a.handleItems)

	log.Infof("Listening on %s", addr)
	return http.ListenAndServe(addr, utils.SecurityHeaders(mux))
}
