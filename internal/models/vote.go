package models

type Vote struct {
	NickName string `json:"nickname"`
	Voice    int    `json:"voice"`
	Existed  bool
}
