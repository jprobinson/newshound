package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ChimeraCoder/anaconda"
	"github.com/jprobinson/newshound"
	"github.com/jprobinson/newshound/bark"
)

func main() {
	config := newshound.NewConfig()

	// SLACK
	for _, slackAlert := range config.SlackAlerts {
		bark.AddSlackAlertBot(d, slackAlert.Key, slackAlert.Bot)
	}
	for _, slackEvent := range config.SlackEvents {
		bark.AddSlackEventBot(d, slackEvent.Key, slackEvent.Bot)
	}

	// TWITTER
	for _, twitterCreds := range config.Twitter {
		anaconda.SetConsumerKey(twitterCreds.ConsumerKey)
		anaconda.SetConsumerSecret(twitterCreds.ConsumerSecret)
		bark.AddTwitterAlertBot(d, twitterCreds.AccessToken, twitterCreds.AccessTokenSecret)
		bark.AddTwitterEventBot(d, twitterCreds.AccessToken, twitterCreds.AccessTokenSecret)
		bark.AddTwitterEventUpdateBot(d, twitterCreds.AccessToken, twitterCreds.AccessTokenSecret)
	}

	// WOOF
	quit := d.ListenAndBark()

	// wait for kill
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	log.Printf("Received signal %s. Initiating stop.", <-ch)

	// signal stop and wait
	errs := make(chan error, 1)
	quit <- errs
	err = <-errs
	if err != nil {
		log.Fatal("shut down error: ", err)
	}
}
