package fetch

import (
	"log"
	"sync"
	"time"

	"github.com/jprobinson/eazye"

	"github.com/jprobinson/newshound"
)

//var procs = runtime.NumCPU() * 2
var procs = 1

func GetMail(cfg *newshound.Config) {
	log.Print("getting mail")
	start := time.Now()
	var count int
	mail := make(chan eazye.Response, 10)
	alerts := make(chan newshound.NewsAlert, 10)
	go eazye.GenerateUnread(cfg.MailboxInfo, cfg.MarkRead, false, mail)

	var parsers sync.WaitGroup
	for i := 0; i < procs; i++ {
		parsers.Add(1)
		// multi goroutines so we can utilize the CPU while waiting for URLs
		go parseMessages(cfg.User, mail, alerts, &parsers)
	}

	//	save the alerts and do event processing
	go func() {
		for alert := range alerts {
			count++
			log.Print("sender:", alert.Sender, ":subject: ", alert.Subject, "\n\n")
			//log.Print("tags: ", alert.Tags)
			for _, sent := range alert.Sentences {
				log.Print("sentence: ", sent.Value)
				log.Printf("phrases: %+v", sent.Phrases)
			}
			log.Print("NEXT!\n\n")
		}
	}()

	// wait for the parsers to complete and then close the alerts channel
	parsers.Wait()
	close(alerts)

	log.Printf("fetched %d messages in %s", count, time.Since(start))
}

func parseMessages(user string, mail <-chan eazye.Response, alerts chan<- newshound.NewsAlert, wg *sync.WaitGroup) {
	defer wg.Done()

	var (
		na  newshound.NewsAlert
		err error
	)
	for resp := range mail {
		if resp.Err != nil {
			log.Print("unable to fetch mail: ", err)
			return
		}

		if na, err = NewNewsAlert(resp.Email, user); err != nil {
			log.Print("unable parse email: ", err)
			continue
		}

		alerts <- na
	}
}
