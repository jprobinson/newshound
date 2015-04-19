package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jprobinson/go-utils/utils"
	"github.com/jprobinson/newshound"
	"github.com/jprobinson/newshound/bark"
)

const logPath = "/var/log/newshound/barkd.log"

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

	d, err := bark.NewDistributor(config.NSQDAddr, config.NSQLAddr, config.BarkdChannel)
	if err != nil {
		log.Fatal(err)
	}

	for _, slackAlert := range config.SlackAlerts {
		bark.AddSlackAlertBot(d, slackAlert.Key, slackAlert.Bot)
	}
	for _, slackEvent := range config.SlackEvents {
		bark.AddSlackEventBot(d, slackEvent.Key, slackEvent.Bot)
	}

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
