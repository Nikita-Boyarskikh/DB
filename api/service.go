package api

import (
	"github.com/Nikita-Boyarskikh/DB/db"
	"github.com/mailru/easyjson"
	"github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
)

func ServiceRouter(service *routing.RouteGroup) {
	service.Get("/status", func(ctx *routing.Context) error {
		logApi(ctx)

		status, err := db.GetStatus()
		if err != nil {
			log.Println("\t500:\t", err)
			return err
		}

		json, err := easyjson.Marshal(status)
		if err != nil {
			log.Println("\t500:\t", err)
			return err
		}

		ctx.Success("application/json", json)

		log.Println("\t200\t", string(json))
		return nil
	})

	service.Post("/clear", func(ctx *routing.Context) error {
		logApi(ctx)

		if err := db.Clear(); err != nil {
			log.Println("\t500:\t", err)
			return err
		}

		ctx.SetStatusCode(fasthttp.StatusOK)
		log.Println("\t200")
		return nil
	})
}