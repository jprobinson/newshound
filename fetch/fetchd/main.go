package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"time"

	"github.com/jprobinson/go-utils/utils"

	"github.com/jprobinson/newshound"
	"github.com/jprobinson/newshound/fetch"

	_ "github.com/lib/pq"
)

const logPath = "/var/log/newshound/fetchd.log"

var (
	logArg  = flag.String("log", logPath, "log path")
	reparse = flag.Bool("r", false, "reparse all alerts and events")
)

func main() {

	flag.Parse()

	ctx := context.Background()

	if *logArg != "stderr" {
		logSetup := utils.NewDefaultLogSetup(*logArg)
		logSetup.SetupLogging()
		go utils.ListenForLogSignal(logSetup)
	} else {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	config := newshound.NewConfig()

	db, err := config.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if *reparse {
		if err := fetch.ReParse(ctx, config, fetch.NewDB(db)); err != nil {
			log.Fatal(err)
		}
		return
	}

	go fetchMail(ctx, config, db)

}

/*
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
*/

func fetchMail(ctx context.Context, config *newshound.Config, db *sql.DB) {
	for {
		fetch.FetchMail(ctx, config, fetch.NewDB(db))
		time.Sleep(30 * time.Second)
	}
}
