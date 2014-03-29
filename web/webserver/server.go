package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/jprobinson/go-utils/utils"
	"github.com/jprobinson/go-utils/web"

	"github.com/jprobinson/newshound/web/webserver/api"
)

const (
	configFile = "/opt/newshound/etc/config.json"

	serverLog = "/var/log/newshound/server.log"
	accessLog = "/var/log/newshound/access.log"

	webDir = "/opt/newshound/www"
)

type config struct {
	DBURL      string `json:"db-url"`
	DBUser     string `json:"db-user"`
	DBPassword string `json:"db-pw"`
}

func main() {
	config := NewConfig()

	logSetup := utils.NewDefaultLogSetup(serverLog)
	logSetup.SetupLogging()
	go utils.ListenForLogSignal(logSetup)

	router := mux.NewRouter()

	api := api.NewNewshoundAPI(config.DBURL, config.DBUser, config.DBPassword)
	apiRouter := router.PathPrefix(api.UrlPrefix()).Subrouter()
	api.Handle(apiRouter)

	staticRouter := router.PathPrefix("/").Subrouter()
	staticRouter.PathPrefix("/").Handler(http.FileServer(http.Dir(webDir)))

	handler := web.AccessLogHandler(accessLog, router)

	log.Fatal(http.ListenAndServe(":8080", handler))
}

func NewConfig() *config {
	config := config{}

	readBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		panic(fmt.Sprintf("Cannot read config file: %s %s", config, err))
	}

	err = json.Unmarshal(readBytes, &config)
	if err != nil {
		panic(fmt.Sprintf("Cannot parse JSON in config file: %s %s", config, err))
	}

	return &config
}
