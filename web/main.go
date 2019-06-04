package main

import (
	"log"

	"github.com/NYTimes/gizmo/server"
	"github.com/jprobinson/newshound/web/api"
)

func main() {
	svc, err := api.NewService()
	if err != nil {
		log.Fatalf("unable to init service: %s", err)
	}

	server.Init("", &server.Config{HTTPPort: 8080})

	server.Register(svc)

	err = server.Run()
	if err != nil {
		log.Fatalf("server encountered a fatal error: %s", err)
	}
}
