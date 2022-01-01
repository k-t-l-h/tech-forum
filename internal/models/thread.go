package models

import "time"

type Thread struct {
	Author  string `json:"author"`
	Message string `json:"message"`
	Title   string `json:"title"`

	CreatedAt time.Time `json:"created,omitempty"`
	Forum     string    `json:"forum"`
	Id        int       `json:"id,omitempty"`
	Slug      string    `json:"slug,omitempty"`
	Votes     int       `json:"votes,omitempty"`
}
