package main

import (
	"os"
	"time"

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

	timer := time.NewTimer(time.Minute * 6)
	go func() {
		<-timer.C
		db.Vacuum()
	}()

	if s.ListenAndServe(config.URI) != nil {
		os.Exit(1)
	}
}
