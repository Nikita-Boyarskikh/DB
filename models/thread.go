package models

import "github.com/mailru/easyjson/opt"

//easyjson:json
type Thread struct {
	ID      opt.Int32
	Author  string
	Created opt.String
	Forum   opt.String
	Message string
	Title   string
	Slug    opt.String
	Votes   opt.Int32
}

//easyjson:json
type PatchThread struct {
	Title   opt.String
	Message opt.String
}

//easyjson:json
type Threads []Thread

//easyjson:json
type Vote struct {
	Nickname string
	Voice    int32
}
