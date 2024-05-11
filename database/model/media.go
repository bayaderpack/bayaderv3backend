package model

// Hobby model - `hobbies` table
type Media struct {
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	IsFolder bool    `json:"isFolder"`
	Children []Media `json:"children,omitempty"`
}
