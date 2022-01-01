package models

type Error struct {
	Message string `json:"message,omitempty"`
}

const (
	OK = iota
	UserNotFound
	ForumConflict
	NotFound
)
