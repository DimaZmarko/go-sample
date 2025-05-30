package models

import (
	"time"
)

type Team struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Users       []User    `json:"users" gorm:"many2many:team_users;"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   time.Time `json:"deleted_at,omitempty" gorm:"index"`
} 