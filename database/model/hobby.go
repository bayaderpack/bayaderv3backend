package model

import (
	"gorm.io/gorm"
)

// Hobby model - `hobbies` table
type Hobby struct {
	gorm.Model
	Hobby     string         `json:"hobby,omitempty"`
	// UserID     []User         `json:"-" `
}
