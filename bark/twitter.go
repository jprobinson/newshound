package bark

import (
	"fmt"
	"strings"

	"github.com/ChimeraCoder/anaconda"
	"github.com/jprobinson/newshound"
)

type TwitterAlertBarker struct {
	api *anaconda.TwitterApi
}

func AddTwitterAlertBot(d *Distributor, key, secret string) {
	d.AddAlertBarker(&TwitterAlertBarker{anaconda.NewTwitterApi(key, secret)})
}

func (s *TwitterAlertBarker) Bark(alert newshound.NewsAlertLite) error {
	msg := twitterize(fmt.Sprintf("%s - %s", alert.Sender, alert.TopSentence))
	msg = msg + fmt.Sprintf("http://newshound.jprbnsn.com/#/calendar?start=%s&display=alerts&alert=%s",
		alert.Timestamp.Format("2006-01-02"),
		alert.ID.Hex())
	_, err := s.api.PostTweet(msg, nil)
	return err
}

type TwitterEventBarker struct {
	api *anaconda.TwitterApi
}

func AddTwitterEventBot(d *Distributor, key, secret string) {
	d.AddEventBarker(&TwitterEventBarker{anaconda.NewTwitterApi(key, secret)})
}

func (s *TwitterEventBarker) Bark(event newshound.NewsEvent) error {
	msg := fmt.Sprintf("New News Event! %d alerts reporting on ", len(event.NewsAlerts))
	msg = twitterize(msg + strings.TrimPrefix(fmt.Sprintf("%#v", event.Tags), "[]string"))
	msg = msg + fmt.Sprintf("http://newshound.jprbnsn.com/#/calendar?start=%s&display=events&event=%s",
		event.EventStart.Format("2006-01-02"),
		event.ID.Hex())

	_, err := s.api.PostTweet(msg, nil)
	return err
}

type TwitterEventUpdateBarker struct {
	api *anaconda.TwitterApi
}

func AddTwitterEventUpdateBot(d *Distributor, key, secret string) {
	d.AddEventUpdateBarker(&TwitterEventUpdateBarker{anaconda.NewTwitterApi(key, secret)})
}

func (s *TwitterEventUpdateBarker) Bark(event newshound.NewsEvent) error {
	msg := fmt.Sprintf("News Event Update! Now %d alerts are reporting on ", len(event.NewsAlerts))
	msg = twitterize(msg + strings.TrimPrefix(fmt.Sprintf("%#v", event.Tags), "[]string"))
	msg = msg + fmt.Sprintf("http://newshound.jprbnsn.com/#/calendar?start=%s&display=events&event=%s",
		event.EventStart.Format("2006-01-02"),
		event.ID.Hex())

	_, err := s.api.PostTweet(msg, nil)
	return err
}

func twitterize(msg string) string {
	if len(msg) > 114 {
		msg = msg[:114] + "... "
	}
	return msg
}
