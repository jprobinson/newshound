package main

import (
	"flag"
	"log"

	"github.com/jprobinson/go-utils/utils"

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

	fetchMail(config)

	/*
		errs := 0
		for {
			err = fetch.MapReduce(sess)
			if err != nil {
				if errs > 10 {
					log.Fatal(err)
				}
				log.Print(err)
				continue
			}

		time.Sleep(1 * time.Hour)
		}
	*/
}

func fetchMail(config *newshound.Config) {
	fetch.GetMail(config)
}
