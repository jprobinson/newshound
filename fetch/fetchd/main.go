package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"gopkg.in/mgo.v2"

	"github.com/NYTimes/gizmo/observe"
	"github.com/NYTimes/gizmo/pubsub"
	"github.com/NYTimes/gizmo/pubsub/gcp"
	"github.com/gorilla/mux"
	"github.com/jprobinson/newshound/fetch"
)

func main() {
	reparse := flag.Bool("r", false, "reparse all alerts and events")
	flag.Parse()

	ctx := context.Background()
	config := fetch.NewConfig()

	observe.RegisterAndObserveGCP(func(err error) {
		log.Printf("observe error: %s", err)
	})

	proj := os.Getenv("GOOGLE_CLOUD_PROJECT")
	apub, err := gcp.NewPublisher(ctx, gcp.Config{Topic: "alerts", ProjectID: proj})
	if err != nil {
		log.Fatal("unable to init alerts publisher: ", err)
	}

	epub, err := gcp.NewPublisher(ctx, gcp.Config{Topic: "events", ProjectID: proj})
	if err != nil {
		log.Fatal("unable to init events publisher: ", err)
	}

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

	fetchMail(ctx, config, sess, apub, epub)
}

func fetchMail(ctx context.Context, config *fetch.Config, sess *mgo.Session, apub, epub pubsub.MultiPublisher) {
	for {
		fetch.FetchMail(ctx, config, sess, apub, epub)
		time.Sleep(120 * time.Second)
	}
}
