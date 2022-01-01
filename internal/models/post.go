package models

type Post struct {
	//обязательно
	Author    string `json:"author"`
	CreatedAt string `json:"created,omitempty"`
	Forum     string `json:"forum,omitempty"`
	Id        int    `json:"id,omitempty"`
	IsEdited  bool   `json:"isEdited,omitempty"`
	Message   string `json:"message"`
	Parent    int    `json:"parent,omitempty"`
	Thread    int    `json:"thread,omitempty"`
}
