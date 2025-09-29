package models

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey"`
	Email        string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`

	// If a user is deleted, delete their items (handy for demos)
	Items []Item `gorm:"constraint:OnDelete:CASCADE;"`
}

type Item struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index;not null"`   // index for per-user queries
	Name      string    `gorm:"not null"`         // required
	CreatedAt time.Time `gorm:"autoCreateTime;index"` // index helps with ORDER BY created_at DESC
}
