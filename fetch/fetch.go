package fetch

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync"
	"time"

	nsq "github.com/bitly/go-nsq"
	"github.com/jprobinson/eazye"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/jprobinson/newshound"
)

// https://github.com/golang/go/issues/3575 :(
var procs = runtime.NumCPU()

func FetchMail(cfg *newshound.Config, sess *mgo.Session) {
	log.Print("getting mail")
	start := time.Now()

	// get a hold of NSQ to we can emit alerts
	producer, err := nsq.NewProducer(cfg.NSQDAddr, nsq.NewConfig())
	if err != nil {
		log.Printf("unable to initiate NSQ producer at @ %s: %s", cfg.NSQDAddr, err)
	}

	// give it 1000 buffer so we can load whatever IMAP throws at us in memory
	alerts := make(chan newshound.NewsAlert, 100)
	mail, err := eazye.GenerateUnread(cfg.MailboxInfo, cfg.MarkRead, false)
	if err != nil {
		log.Fatal("unable to get mail: ", err)
	}

	var parsers sync.WaitGroup
	for i := 0; i < procs; i++ {
		parsers.Add(1)
		// multi goroutines so we can utilize the CPU while waiting for URLs
		go parseMessages(cfg.User, mail, alerts, &parsers)
	}

	s := sess.Copy()
	defer s.Close()
	db := newshoundDB(s)
	na := newsAlerts(db)
	ne := newsEvents(db)

	completeCount := make(chan int, 1)
	go saveAndRefresh(na, ne, alerts, completeCount, producer)

	// wait for the parsers to complete and then close the alerts channel
	parsers.Wait()
	close(alerts)
	count := <-completeCount

	log.Printf("fetched %d messages in %s", count, time.Since(start))
}

func ReParse(cfg *newshound.Config, sess *mgo.Session) error {
	log.Print("reparsing mail")
	start := time.Now()

	s := sess.Copy()
	defer s.Close()

	db := newshoundDB(s)
	// grab temp collections and wipe them in case of prev err
	na := newsAlertsTemp(db)
	if err := na.DropCollection(); err != nil {
		if !isNotFound(err) {
			return err
		}
	}
	ne := newsEventsTemp(db)
	if err := ne.DropCollection(); err != nil {
		if !isNotFound(err) {
			return err
		}
	}

	alerts := make(chan newshound.NewsAlert, 1000)
	reAlerts := make(chan newshound.NewsAlert, 1000)
	// grab all existing alerts from the main collection
	go getAllAlerts(newsAlerts(db), alerts)

	var parsers sync.WaitGroup
	for i := 0; i < procs; i++ {
		parsers.Add(1)
		// multi goroutines so we can utilize the CPU while waiting for URLs
		go reParseMessages(cfg.User, alerts, reAlerts, &parsers)
	}

	completeCount := make(chan int, 1)
	go saveAndRefresh(na, ne, reAlerts, completeCount, nil)

	// wait for the parsers to complete and then close the alerts channel
	parsers.Wait()
	close(reAlerts)
	count := <-completeCount

	// replace the na/ne main colls with the new temps
	if err := replaceColl(s, "news_alerts_temp", "news_alerts"); err != nil {
		return err
	}
	if err := replaceColl(s, "news_events_temp", "news_events"); err != nil {
		return err
	}

	log.Printf("reparsed %d messages in %s", count, time.Since(start))
	return nil
}

func isNotFound(err error) bool {
	return strings.Contains(err.Error(), "not found")
}

// saveAndRefresh will insert all alerts passed through the channel and kick off all event refreshes
func saveAndRefresh(na *mgo.Collection, ne *mgo.Collection, alerts <-chan newshound.NewsAlert, completeCount chan<- int, producer *nsq.Producer) {
	var count int
	timeframes := map[int64]struct{}{}

	var err error
	for alert := range alerts {
		count++
		if err = na.Insert(alert); err != nil {
			log.Print("unable to save alert to db: ", err)
			continue
		}

		// emit alert notification
		if producer != nil {
			var buff bytes.Buffer
			err = gob.NewEncoder(&buff).Encode(&alert.NewsAlertLite)
			if err != nil {
				log.Print("unable to gob alert: ", err)
			} else {
				if err = producer.Publish(newshound.NewsAlertTopic, buff.Bytes()); err != nil {
					log.Print("unable to publish alert: ", err)
				}
			}
		}

		if count%10 == 0 {
			log.Printf("fetched %d messages", count)
		}

		// find timeframe bucket to prevent too many refreshes
		aTime := alert.Timestamp.Truncate(10 * time.Minute)
		timeframes[aTime.Unix()] = struct{}{}
		if len(timeframes) > 5 {
			for tf, _ := range timeframes {
				if err = EventRefresh(na, ne, time.Unix(tf, 0), producer); err != nil {
					log.Print("problems refreshing event: ", err)
				}
			}
			timeframes = map[int64]struct{}{}
		}
	}
	// flush the timeframe buffer at the end
	for tf, _ := range timeframes {
		if err = EventRefresh(na, ne, time.Unix(tf, 0), producer); err != nil {
			log.Print("problems refreshing event: ", err)
		}
	}
	completeCount <- count
}

func reParseMessages(user string, alerts <-chan newshound.NewsAlert, reAlerts chan<- newshound.NewsAlert, wg *sync.WaitGroup) {
	defer wg.Done()

	var (
		na  newshound.NewsAlert
		err error
	)
	for alert := range alerts {
		if na, err = ReParseNewsAlert(alert, user); err != nil {
			// panic so that we stop the reparse and dont lose any data.
			// we're good to die at this point bc temp collections ftw!
			log.Fatal("unable to reparse email: ", err)
		}

		reAlerts <- na
	}
}

func parseMessages(user string, mail chan eazye.Response, alerts chan<- newshound.NewsAlert, wg *sync.WaitGroup) {
	defer wg.Done()

	var (
		na  newshound.NewsAlert
		err error
	)
	for resp := range mail {
		if resp.Err != nil {
			log.Fatalf("unable to fetch mail: %s", resp.Err)
			return
		}

		if na, err = NewNewsAlert(resp.Email, user); err != nil {
			log.Print("unable to parse email: ", err)
			continue
		}

		// only post approved senders
		if Senders[strings.ToLower(na.Sender)] {
			alerts <- na
		} else {
			log.Print("skipping email from: ", na.Sender)
		}
	}
}

func getAllAlerts(na *mgo.Collection, alerts chan<- newshound.NewsAlert) {
	i := na.Find(nil).Batch(1000).Iter()
	var alert newshound.NewsAlert
	for i.Next(&alert) {
		alerts <- alert
	}

	if err := i.Close(); err != nil {
		log.Print("unable to get all alerts from db: ", err)
	}
	close(alerts)
}

func replaceColl(sess *mgo.Session, from, to string) error {
	db := sess.DB("admin")
	from = fmt.Sprint("newshound.", from)
	to = fmt.Sprint("newshound.", to)
	err := db.Run(bson.D{{"renameCollection", from}, {"to", to}, {"dropTarget", true}}, nil)
	if err != nil {
		return fmt.Errorf("unable to replace %s: %s", to, err)
	}
	return nil
}

func newshoundDB(sess *mgo.Session) *mgo.Database {
	return sess.DB("newshound")
}

func newsAlerts(db *mgo.Database) *mgo.Collection {
	return db.C("news_alerts")
}

func newsEvents(db *mgo.Database) *mgo.Collection {
	return db.C("news_events")
}

func newsAlertsTemp(db *mgo.Database) *mgo.Collection {
	return db.C("news_alerts_temp")
}

func newsEventsTemp(db *mgo.Database) *mgo.Collection {
	return db.C("news_events_temp")
}
