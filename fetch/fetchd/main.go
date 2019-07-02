package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"gopkg.in/mgo.v2"

	"github.com/gorilla/mux"
	"github.com/jprobinson/newshound/fetch"
)

func main() {
	reparse := flag.Bool("r", false, "reparse all alerts and events")
	flag.Parse()

	ctx := context.Background()
	config := fetch.NewConfig()

	sess, err := config.MgoSession()
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	if *reparse {
		if err := fetch.ReParse(config, sess); err != nil {
			log.Fatal(err)
		}
		return
	}

	go func() {
		mv := mux.NewRouter()
		mv.HandleFunc("/mapreduce", func(w http.ResponseWriter, r *http.Request) {
			err := fetch.MapReduce(sess)
			if err != nil {
				log.Print("problems performing mapreduce: ", err)
			}
			w.WriteHeader(http.StatusOK)
		})
		ok := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}
		mv.HandleFunc("/_ah/warmup", ok)
		mv.HandleFunc("/", ok)
		// for GAE
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		log.Printf("listening on %s", port)
		http.ListenAndServe(":"+port, mv)
	}()

	fetchMail(ctx, config, sess)
}

func fetchMail(ctx context.Context, config *fetch.Config, sess *mgo.Session) {
	for {
		fetch.FetchMail(ctx, config, sess)
		time.Sleep(30 * time.Second)
	}
}
