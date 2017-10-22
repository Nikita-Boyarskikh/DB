package api

import "github.com/qiangxue/fasthttp-routing"

var router *routing.Router

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

func GetRouter() *routing.Router {
	return router
}
