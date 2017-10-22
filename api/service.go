package api

import (
	"github.com/Nikita-Boyarskikh/DB/db"
	"github.com/mailru/easyjson"
	"github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
)

func ServiceRouter(service *routing.RouteGroup) {
	service.Get("/status", func(ctx *routing.Context) error {
		status, err := db.GetStatus()
		if err != nil {
			return err
		}

		json, err := easyjson.Marshal(status)
		if err != nil {
			return err
		}

		ctx.Success("application/json", json)
		return nil
	})

	service.Post("/clear", func(ctx *routing.Context) error {
		if err := db.Clear(); err != nil {
			return err
		}

		ctx.SetStatusCode(fasthttp.StatusOK)
		return nil
	})
}
