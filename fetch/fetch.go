package fetch

import (
	"bytes"
	"context"
	"encoding/gob"
	"log"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/bitly/go-nsq"
	"github.com/jprobinson/eazye"

	"github.com/jprobinson/newshound"
)

// https://github.com/golang/go/issues/3575 :(
var procs = runtime.NumCPU()

func FetchMail(ctx context.Context, cfg *newshound.Config, db DB) {
	log.Print("getting mail")
	start := time.Now()

	// SWITCH TO PUBSUB

	alerts := make(chan *newshound.NewsAlert)
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

	completeCount := make(chan int, 1)
	go saveAndRefresh(ctx, db, alerts, completeCount, nil)

	// wait for the parsers to complete and then close the alerts channel
	parsers.Wait()
	close(alerts)
	count := <-completeCount

	log.Printf("fetched %d messages in %s", count, time.Since(start))
}

func ReParse(ctx context.Context, cfg *newshound.Config, db DB) error {
	log.Print("reparsing mail")
	start := time.Now()

	reAlerts := make(chan *newshound.NewsAlert)
	// grab all existing alerts from the main collection
	alerts, err := db.GetAllAlerts(ctx)
	if err != nil {
		return err
	}

	var parsers sync.WaitGroup
	for i := 0; i < procs; i++ {
		parsers.Add(1)
		// multi goroutines so we can utilize the CPU while waiting for URLs
		go reParseMessages(cfg.User, alerts, reAlerts, &parsers)
	}

	completeCount := make(chan int, 1)
	go saveAndRefresh(ctx, db, reAlerts, completeCount, nil)

	// wait for the parsers to complete and then close the alerts channel
	parsers.Wait()
	close(reAlerts)
	count := <-completeCount

	log.Printf("reparsed %d messages in %s", count, time.Since(start))
	return nil
}

func isNotFound(err error) bool {
	return strings.Contains(err.Error(), "not found")
}

// saveAndRefresh will insert all alerts passed through the channel and kick off all event refreshes
func saveAndRefresh(ctx context.Context, db DB, alerts <-chan *newshound.NewsAlert, completeCount chan<- int, producer *nsq.Producer) {
	var count int
	timeframes := map[int64]struct{}{}

	var err error
	for alert := range alerts {
		count++
		//
		if err = db.PutAlert(ctx, alert); err != nil {
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
				if err = EventRefresh(ctx, db, time.Unix(tf, 0), producer); err != nil {
					log.Print("problems refreshing event: ", err)
				}
			}
			timeframes = map[int64]struct{}{}
		}
	}
	// flush the timeframe buffer at the end
	for tf, _ := range timeframes {
		if err = EventRefresh(ctx, db, time.Unix(tf, 0), producer); err != nil {
			log.Print("problems refreshing event: ", err)
		}
	}
	completeCount <- count
}

func reParseMessages(user string, alerts <-chan *newshound.NewsAlert, reAlerts chan<- *newshound.NewsAlert, wg *sync.WaitGroup) {
	defer wg.Done()
	for alert := range alerts {
		na, err := ReParseNewsAlert(alert, user)
		if err != nil {
			// panic so that we stop the reparse and dont lose any data.
			// we're good to die at this point bc temp collections ftw!
			log.Fatal("unable to reparse email: ", err)
		}
		reAlerts <- na
	}
}

func parseMessages(user string, mail chan eazye.Response, alerts chan<- *newshound.NewsAlert, wg *sync.WaitGroup) {
	defer wg.Done()

	for resp := range mail {
		if resp.Err != nil {
			log.Fatalf("unable to fetch mail: %s", resp.Err)
			return
		}

		na, err := NewNewsAlert(resp.Email, user)
		if err != nil {
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
