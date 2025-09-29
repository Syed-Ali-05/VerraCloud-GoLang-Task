package models

import (
    "os"
    "strings"
    log "github.com/sirupsen/logrus"
    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"
    "github.com/glebarez/sqlite"
)

const (
    AdminEmail = "admin@example.com"
    AdminPass  = "Passw0rd!"
)

type User struct {
    ID           uint   `gorm:"primaryKey"`
    Email        string `gorm:"uniqueIndex;size:255;not null"`
    PasswordHash string `gorm:"not null"`
    CreatedAt    int64  `gorm:"autoCreateTime"`
}

type Item struct {
    ID        uint   `gorm:"primaryKey"`
    UserID    uint   `gorm:"index;not null"`
    Name      string `gorm:"size:255;not null"`
    CreatedAt int64  `gorm:"autoCreateTime"`
}

func InitDB() *gorm.DB {
    dbPath := strings.TrimSpace(os.Getenv("DB_PATH"))
    if dbPath == "" {
        dbPath = "app.db"
    }

    db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
    if err != nil {
        log.Fatalf("open db: %v", err)
    }

    if err := db.AutoMigrate(&User{}, &Item{}); err != nil {
        log.Fatalf("migration failed: %v", err)
    }

    // seed admin
    var count int64
    db.Model(&User{}).Where("email = ?", AdminEmail).Count(&count)
    if count == 0 {
        hash, _ := bcrypt.GenerateFromPassword([]byte(AdminPass), bcrypt.DefaultCost)
        db.Create(&User{Email: AdminEmail, PasswordHash: string(hash)})
        log.Infof("Seeded admin user %s / %s", AdminEmail, AdminPass)
    }

    return db
}
