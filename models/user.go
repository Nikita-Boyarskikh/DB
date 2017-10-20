package models

import "github.com/mailru/easyjson/opt"

//easyjson:json
type User struct {
	Nickname opt.String
	Fullname string
	Email    string
	About    opt.String
}

//easyjson:json
type Users []User
