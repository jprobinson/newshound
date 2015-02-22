package fetch

import (
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/jprobinson/eazye"
	"gopkg.in/mgo.v2"

	"github.com/jprobinson/newshound"
)

// https://github.com/golang/go/issues/3575 :(
var procs = runtime.NumCPU()

func GetMail(cfg *newshound.Config, sess *mgo.Session) {
	log.Print("getting mail")
	start := time.Now()
	var count int
	// give it 1000 buffer so we can load whatever IMAP throws at us in memory
	mail := make(chan eazye.Response, 1000)
	alerts := make(chan newshound.NewsAlert, 100)
	go eazye.GenerateUnread(cfg.MailboxInfo, cfg.MarkRead, false, mail)

	var parsers sync.WaitGroup
	for i := 0; i < procs; i++ {
		parsers.Add(1)
		// multi goroutines so we can utilize the CPU while waiting for URLs
		go parseMessages(cfg.User, mail, alerts, &parsers)
	}

	//	save the alerts and do event processing
	go func(sess *mgo.Session, alerts chan newshound.NewsAlert) {
		timeframes := map[int64]struct{}{}
		s := sess.Copy()
		db := s.DB("newshound")
		alertsC := newsAlerts(db)

		var err error
		for alert := range alerts {
			count++
			if err = alertsC.Insert(alert); err != nil {
				log.Print("unable to save alert to db: ", err)
				continue
			}
			if count%10 == 0 {
				log.Printf("fetched %d messages", count)
			}

			// find timeframe bucket to prevent too many refreshes
			aTime := alert.Timestamp.Truncate(10 * time.Minute)
			timeframes[aTime.Unix()] = struct{}{}
			if len(timeframes) > 5 {
				for tf, _ := range timeframes {
					if err = EventRefresh(db, time.Unix(tf, 0)); err != nil {
						log.Print("problems refreshing event: ", err)
					}
				}
				timeframes = map[int64]struct{}{}
			}
		}
		// flush the timeframe buffer at the end
		for tf, _ := range timeframes {
			if err = EventRefresh(db, time.Unix(tf, 0)); err != nil {
				log.Print("problems refreshing event: ", err)
			}
		}
	}(sess, alerts)

	// wait for the parsers to complete and then close the alerts channel
	parsers.Wait()
	close(alerts)

	log.Printf("fetched %d messages in %s", count, time.Since(start))
}

func parseMessages(user string, mail chan eazye.Response, alerts chan<- newshound.NewsAlert, wg *sync.WaitGroup) {
	defer wg.Done()

	var (
		na  newshound.NewsAlert
		err error
	)
	for resp := range mail {
		if resp.Err != nil {
			log.Print("unable to fetch mail: ", resp.Err)
			close(mail)
			return
		}

		if na, err = NewNewsAlert(resp.Email, user); err != nil {
			log.Print("unable parse email: ", err)
			continue
		}

		alerts <- na
	}
}

func newsAlerts(db *mgo.Database) *mgo.Collection {
	return db.C("news_alerts")
}

func newsEvents(db *mgo.Database) *mgo.Collection {
	return db.C("news_events")
}
