package main

import (
    "html/template"
    "log"
    "net/http"

    "go-htmx-auth-inline/internal/db"
    "go-htmx-auth-inline/internal/handlers"
)

func main() {
    // Init DB
    database := db.InitDB()

    // Parse templates
    templates := template.Must(template.ParseGlob("templates/*.html"))

    mux := http.NewServeMux()
    mux.HandleFunc("/", handlers.IndexHandler(templates))
    mux.HandleFunc("/login", handlers.LoginHandler(database, templates))
    mux.HandleFunc("/logout", handlers.LogoutHandler(templates)) // ✅ matches signature
    mux.HandleFunc("/items", handlers.ItemsHandler(database, templates))

    log.Println("listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", mux))
}
