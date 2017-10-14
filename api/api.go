package api

import (
	stdlog "log"

	"github.com/Nikita-Boyarskikh/DB/logger"
	"github.com/qiangxue/fasthttp-routing"
)

//easyjson:json
type Error struct {
	Message string
}

var (
	router *routing.Router
	log    stdlog.Logger
)

func Init() {
	log = logger.GetLogger()
	log.SetPrefix("API:")
}

func Routing(router *routing.Router) *routing.Router {
	api := router.Group("/api")
	forum := api.Group("/forum")
	ForumRouter(forum)

	post := api.Group("/post")
	PostRouter(post)

	service := api.Group("/service")
	ServiceRouter(service)

	thread := api.Group("/thread")
	ThreadRouter(thread)

	user := api.Group("/user")
	UserRouter(user)

	return router
}

func logApi(ctx *routing.Context) {
	log.Printf("%s %s from %s", ctx.Method(), ctx.URI(), ctx.RemoteIP())
	if ctx.IsPost() {
		log.Printf("\tPost data: '%s'", ctx.PostBody())
	}
}

func GetRouter() *routing.Router {
	return router
}