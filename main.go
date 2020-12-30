package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/umerm-work/arcTest/config"
	"github.com/umerm-work/arcTest/controller"
	"github.com/umerm-work/arcTest/db"
	"log"
)

func main() {

	conf := config.ReadEnv()

	dpRepo := db.New(conf)
	//svc := service.NewBasicService(dpRepo)
	router := mux.NewRouter()
	cRouter := controller.App{Router: router, DB: dpRepo}
	log.Print("Starting server")
	cRouter.InitializeRoutes()
	cRouter.Run(fmt.Sprintf(`:%s`, conf.AppPort))
}
