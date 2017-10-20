package models

import "github.com/mailru/easyjson/opt"

//easyjson:json
type Forum struct {
	Posts   opt.Int64
	Slug    string
	Threads opt.Int32
	Title   string
	User    string
}
