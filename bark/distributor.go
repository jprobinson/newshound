package bark

import (
	"bytes"
	"encoding/gob"
	"log"

	"github.com/bitly/go-nsq"
	"github.com/jprobinson/newshound"
)

type Distributor struct {
	nsqdAddr string
	nsqlAddr string

	alertsIn  *nsq.Consumer
	alertsOut []AlertBarker

	eventsIn  *nsq.Consumer
	eventsOut []EventBarker

	eventUpdatesIn  *nsq.Consumer
	eventUpdatesOut []EventBarker
}

type AlertBarker interface {
	Bark(alert newshound.NewsAlertLite) error
}

type AlertBarkerFunc func(alert newshound.NewsAlertLite) error

func (a AlertBarkerFunc) Bark(alert newshound.NewsAlertLite) error {
	return a(alert)
}

type EventBarker interface {
	Bark(alert newshound.NewsEvent) error
}

type EventBarkerFunc func(event newshound.NewsEvent) error

func (e EventBarkerFunc) Bark(event newshound.NewsEvent) error {
	return e(event)
}

func NewDistributor(nsqdAddr, nsqlAddr, channel string) (d *Distributor, err error) {
	d = &Distributor{nsqdAddr: nsqdAddr, nsqlAddr: nsqlAddr}
	d.alertsIn, err = nsq.NewConsumer(newshound.NewsAlertTopic, channel, nsq.NewConfig())
	if err != nil {
		return d, err
	}
	d.alertsIn.AddHandler(d.alertHandler())

	d.eventsIn, err = nsq.NewConsumer(newshound.NewsEventTopic, channel, nsq.NewConfig())
	if err != nil {
		return d, err
	}
	d.eventsIn.AddHandler(d.eventHandler())

	d.eventUpdatesIn, err = nsq.NewConsumer(newshound.NewsEventUpdateTopic, channel, nsq.NewConfig())
	if err != nil {
		return d, err
	}
	d.eventUpdatesIn.AddHandler(d.eventUpdateHandler())

	return d, nil
}

func (d *Distributor) AddAlertBarker(b AlertBarker) {
	d.alertsOut = append(d.alertsOut, b)
}

func (d *Distributor) AddEventBarker(b EventBarker) {
	d.eventsOut = append(d.eventsOut, b)
}

func (d *Distributor) AddEventUpdateBarker(b EventBarker) {
	d.eventUpdatesOut = append(d.eventUpdatesOut, b)
}

func (d *Distributor) alertHandler() nsq.HandlerFunc {
	return nsq.HandlerFunc(func(m *nsq.Message) error {
		var alert newshound.NewsAlertLite
		err := gob.NewDecoder(bytes.NewBuffer(m.Body)).Decode(&alert)
		if err != nil {
			log.Printf("unable to read alert message: %s\n%q", err, alert)
			return nil
		}

		for _, barker := range d.alertsOut {
			go func(barker AlertBarker, alert newshound.NewsAlertLite) {
				if err = barker.Bark(alert); err != nil {
					log.Print("problems barking about alert: ", err)
				}
			}(barker, alert)
		}
		return nil
	})
}

func (d *Distributor) eventHandler() nsq.HandlerFunc {
	return nsq.HandlerFunc(func(m *nsq.Message) error {
		var event newshound.NewsEvent
		err := gob.NewDecoder(bytes.NewBuffer(m.Body)).Decode(&event)
		if err != nil {
			log.Printf("unable to read event message: %s\n%q", err, event)
			return nil
		}

		for _, barker := range d.eventsOut {
			go func(barker EventBarker, event newshound.NewsEvent) {
				if err = barker.Bark(event); err != nil {
					log.Print("problems barking about event: ", err)
				}
			}(barker, event)
		}
		return nil
	})
}

func (d *Distributor) eventUpdateHandler() nsq.HandlerFunc {
	return nsq.HandlerFunc(func(m *nsq.Message) error {
		var event newshound.NewsEvent
		err := gob.NewDecoder(bytes.NewBuffer(m.Body)).Decode(&event)
		if err != nil {
			log.Printf("unable to read event update message: %s\n%q", err, event)
			return nil
		}

		for _, barker := range d.eventUpdatesOut {
			go func(barker EventBarker, event newshound.NewsEvent) {
				if err := barker.Bark(event); err != nil {
					log.Print("problems barking about event update: ", err)
				}
			}(barker, event)
		}
		return nil
	})
}

// ListenAndBark will sit on alert feed and notify all hooks
func (a *Distributor) ListenAndBark() chan chan error {
	err := a.alertsIn.ConnectToNSQD(a.nsqdAddr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("alerts connected to %s", a.nsqdAddr)

	// why doesnt lookup work?! -- complains that we're not sending v2 proto??
	/*	err = a.alertsIn.ConnectToNSQLookupd(a.nsqlAddr)
		if err != nil {
			log.Fatal("unable to connect to NSQ: ", err)
		}
	*/

	err = a.eventsIn.ConnectToNSQD(a.nsqdAddr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("events connected to %s", a.nsqdAddr)

	/*
		err = a.eventsIn.ConnectToNSQLookupd(a.nsqlAddr)
		if err != nil {
			log.Fatal("unable to connect to NSQ: ", err)
		}

	*/
	quit := make(chan chan error, 1)
	go func(quit chan chan error) {
		q := <-quit
		a.alertsIn.Stop()
		a.eventsIn.Stop()
		a.eventUpdatesIn.Stop()
		q <- nil
	}(quit)
	return quit
}

var SenderColors = map[string]string{
	"cnn":                 "#B60002",
	"foxnews.com":         "#234E6C",
	"foxbusiness.com":     "#343434",
	"nbcnews.com":         "#343434",
	"nytimes.com":         "#1A1A1A",
	"the washington post": "#222",
	"wsj.com":             "#444242",
	"politico":            "#256396",
	"los angeles times":   "#000",
	"cbs":                 "#313943",
	"abc":                 "#1b6295",
	"usatoday.com":        "#1877B6",
	"yahoo":               "#7B0099",
	"ft":                  "#FFF1E0",
	"bbc":                 "#c00000",
	"npr":                 "#5f82be",
	"time":                "#e90606",
	"bloomberg.com":       "#110c09",
	"time":                "#e90606",
}
