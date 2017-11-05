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

func (p Posts) Len() int {
	return len(p)
}

func (p Posts) Less(i, j int) bool {
	return p[i].Forum.V < p[j].Forum.V || p[i].Forum.V == p[j].Forum.V && p[i].Author < p[j].Author
}

func (p Posts) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

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
