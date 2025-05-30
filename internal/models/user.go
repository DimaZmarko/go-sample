package models

import (
	"time"
)

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"unique"`
	Name      string    `json:"name"`
	Teams     []Team    `json:"teams" gorm:"many2many:team_users;"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at,omitempty" gorm:"index"`
} 