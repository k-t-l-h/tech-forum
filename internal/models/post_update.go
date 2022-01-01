package models

type PostUpdate struct {
	Id      int    `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
}
