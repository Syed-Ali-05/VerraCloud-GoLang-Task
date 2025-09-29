package db

import (
    "log"
    "time"

    "github.com/glebarez/sqlite" // pure Go SQLite driver for GORM
    "gorm.io/gorm"
    "golang.org/x/crypto/bcrypt"
    "go-htmx-auth-inline/internal/models"
)

func InitDB() *gorm.DB {
    // Connect using pure Go driver
    database, err := gorm.Open(sqlite.Open("app.db"), &gorm.Config{})
    if err != nil {
        log.Fatal("failed to connect database: ", err)
    }

    // AutoMigrate tables
    err = database.AutoMigrate(&models.User{}, &models.Item{})
    if err != nil {
        log.Fatal("failed migration: ", err)
    }

    // Seed default admin user if none exist
    var count int64
    database.Model(&models.User{}).Count(&count)
    if count == 0 {
        hash, _ := bcrypt.GenerateFromPassword([]byte("Passw0rd!"), bcrypt.DefaultCost)
        database.Create(&models.User{
            Email:        "admin@example.com",
            PasswordHash: string(hash),
            CreatedAt:    time.Now(),
        })
        log.Println("Seeded admin@example.com / Passw0rd!")
    }

    return database
}
