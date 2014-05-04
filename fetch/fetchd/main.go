package main

import (
	"log"

	"github.com/jprobinson/newshound"
	"github.com/jprobinson/newshound/fetch"
)

func main() {
	config := newshound.NewConfig()

	sess, err := config.MgoSession()
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	err = fetch.MapReduce(sess)
	if err != nil {
		log.Fatal(err)
	}
}
