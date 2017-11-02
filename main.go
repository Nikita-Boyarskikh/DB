package main

import (
	"os"

	"github.com/Nikita-Boyarskikh/DB/api"
	"github.com/Nikita-Boyarskikh/DB/config"
	"github.com/Nikita-Boyarskikh/DB/db"
	"github.com/qiangxue/fasthttp-routing"
)

func main() {
	router := routing.New()
	s := config.Server
	s.Handler = api.Routing(router).HandleRequest

	err := config.Init()
	if err != nil {
		os.Exit(1)
	}

	err = db.Init(config.DBConnection)
	if err != nil {
		os.Exit(1)
	}
	defer db.Close()

	if s.ListenAndServe(config.URI) != nil {
		os.Exit(1)
	}
}
