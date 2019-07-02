package main

import (
	"log"

	"github.com/NYTimes/gizmo/server/kit"
	"github.com/jprobinson/newshound/bark"
)

func main() {
	svc, err := bark.NewService()
	if err != nil {
		log.Fatal(err)
	}

	err = kit.Run(svc)
	if err != nil {
		log.Fatal(err)
	}
}
