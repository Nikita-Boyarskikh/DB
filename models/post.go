package models

import "github.com/mailru/easyjson/opt"

//easyjson:json
type Post struct {
	ID       opt.Int64
	Author   string
	Created  opt.String
	Forum    opt.String
	IsEdited opt.Bool
	Message  string
	Parent   opt.Int64
	Thread   opt.Int32
}

//easyjson:json
type Posts []Post

//easyjson:json
type PostFull struct {
	Post   Post
	Author *User
	Thread *Thread
	Forum  *Forum
}

//easyjson:json
type EditPost struct {
	Message opt.String
}
