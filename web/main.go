package main

import (
	"log"
	"os"
	"strconv"

	"github.com/NYTimes/gizmo/server"
	"github.com/jprobinson/newshound/web/api"
)

func main() {
	svc, err := api.NewService()
	if err != nil {
		log.Fatalf("unable to init service: %s", err)
	}

	cfg := server.Config{HTTPPort: 8080}
	port := os.Getenv("PORT")
	if port != "" {
		cfg.HTTPPort, _ = strconv.Atoi(port)
	}

	server.Init("", &cfg)

	server.Register(svc)

	err = server.Run()
	if err != nil {
		log.Fatalf("server encountered a fatal error: %s", err)
	}
}
