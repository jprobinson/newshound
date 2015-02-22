package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jprobinson/go-utils/utils"
	"github.com/jprobinson/go-utils/web"

	"github.com/jprobinson/newshound"
	"github.com/jprobinson/newshound/web/webserver/api"
)

func main() {
	config := newshound.NewConfig()

	logSetup := utils.NewDefaultLogSetup(newshound.ServerLog)
	logSetup.SetupLogging()
	go utils.ListenForLogSignal(logSetup)

	router := mux.NewRouter()

	api := api.NewNewshoundAPI(config.DBURL, config.DBUser, config.DBPassword)
	apiRouter := router.PathPrefix(api.UrlPrefix()).Subrouter()
	api.Handle(apiRouter)

	staticRouter := router.PathPrefix("/").Subrouter()
	staticRouter.PathPrefix("/").Handler(http.FileServer(http.Dir(newshound.WebDir)))

	handler := web.AccessLogHandler(newshound.AccessLog, router)

	log.Fatal(http.ListenAndServe(":8080", handler))
}
