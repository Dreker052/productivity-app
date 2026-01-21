package models

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"` // это имя будет вставляться в отчете в тг
	Email    string `json:"email"`
	Password string `json:"password"`
}
