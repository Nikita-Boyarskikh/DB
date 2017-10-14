package main

import (
	"os"

	"github.com/Nikita-Boyarskikh/DB/api"
	"github.com/Nikita-Boyarskikh/DB/config"
	"github.com/Nikita-Boyarskikh/DB/db"
	"github.com/Nikita-Boyarskikh/DB/logger"
	"github.com/qiangxue/fasthttp-routing"
)

func main() {
	log := logger.Init()
	api.Init()

	router := routing.New()
	s := config.Server
	s.Handler = api.Routing(router).HandleRequest

	err := config.Init()
	if err != nil {
		log.Fatalln("Unable to initialize config: ", err)
		os.Exit(1)
	}

	err = db.Init(config.DBConnection)
	if err != nil {
		log.Fatalln("Unable to initialize database: ", err)
		os.Exit(1)
	}
	defer db.Close()

	log.Println("I'm started!")
	err = s.ListenAndServe(config.URI)
	if err != nil {
		log.Fatalln("error in ListenAndServe: ", err)
	}
}
