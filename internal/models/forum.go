package models

type Forum struct {
	//обязательно
	Slug  string `json:"slug"`
	Title string `json:"title"`
	User  string `json:"user"`
	//по желанию
	Posts   int `json:"posts,omitempty"`
	Threads int `json:"threads,omitempty"`
}
