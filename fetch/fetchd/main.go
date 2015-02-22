package main

import (
	"flag"
	"log"
	"time"

	"github.com/jprobinson/go-utils/utils"
	"gopkg.in/mgo.v2"

	"github.com/jprobinson/newshound"
	"github.com/jprobinson/newshound/fetch"
)

const logPath = "/var/log/newshound/fetchd.log"

var (
	logArg = flag.String("log", logPath, "log path")
)

func main() {

	flag.Parse()

	if *logArg != "stderr" {
		logSetup := utils.NewDefaultLogSetup(*logArg)
		logSetup.SetupLogging()
		go utils.ListenForLogSignal(logSetup)
	} else {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	config := newshound.NewConfig()

	sess, err := config.MgoSession()
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	go fetchMail(config, sess)

	mapReduce(sess)
}

func mapReduce(sess *mgo.Session) {
	for {
		err := fetch.MapReduce(sess)
		if err != nil {
			log.Print("problems performing mapreduce: ", err)

			time.Sleep(5 * time.Minute)
			continue
		}

		time.Sleep(1 * time.Hour)
	}
}

func fetchMail(config *newshound.Config, sess *mgo.Session) {
	for {
		fetch.GetMail(config, sess)
		time.Sleep(30 * time.Second)
	}
}
