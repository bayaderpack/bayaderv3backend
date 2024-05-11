package model

import (
	"gorm.io/gorm"
)

// Post model - `posts` table
type Post struct {
	gorm.Model
	Title     string         `json:"title,omitempty" structs:"title,omitempty"`
	Body      string         `json:"body,omitempty" structs:"body,omitempty"`
	UserID    uint     `json:"-"`
}
