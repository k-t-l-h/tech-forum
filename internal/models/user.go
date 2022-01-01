package models

type User struct {
	About    string `json:"about,omitempty"`
	Email    string `json:"email"`
	FullName string `json:"fullname"`
	NickName string `json:"nickname,omitempty"`
}
