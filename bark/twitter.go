package bark

import (
	"fmt"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/ChimeraCoder/anaconda"
	"github.com/jprobinson/newshound"
	"golang.org/x/exp/utf8string"
)

type TwitterAlertBarker struct {
	api *anaconda.TwitterApi
}

func AddTwitterAlertBot(d *Distributor, token, secret string) {
	d.AddAlertBarker(&TwitterAlertBarker{anaconda.NewTwitterApi(token, secret)})
}

func (s *TwitterAlertBarker) Bark(alert newshound.NewsAlertLite) error {
	msg := twitterize(fmt.Sprintf("%s - %s", strings.TrimSuffix(alert.Sender, ".com"), alert.TopSentence))
	msg = msg + alertLink(alert)
	_, err := s.api.PostTweet(msg, url.Values{})
	return err
}

type TwitterEventBarker struct {
	api *anaconda.TwitterApi
}

func AddTwitterEventBot(d *Distributor, token, secret string) {
	d.AddEventBarker(&TwitterEventBarker{anaconda.NewTwitterApi(token, secret)})
}

func (s *TwitterEventBarker) Bark(event newshound.NewsEvent) error {
	msg := fmt.Sprintf("New News Event! %d alerts reporting on ", len(event.NewsAlerts))
	msg = twitterize(msg + strings.TrimPrefix(fmt.Sprintf("%#v", event.Tags), "[]string"))
	msg = msg + eventLink(event)
	_, err := s.api.PostTweet(msg, nil)
	return err
}

type TwitterEventUpdateBarker struct {
	api *anaconda.TwitterApi
}

func AddTwitterEventUpdateBot(d *Distributor, token, secret string) {
	d.AddEventUpdateBarker(&TwitterEventUpdateBarker{anaconda.NewTwitterApi(token, secret)})
}

func (s *TwitterEventUpdateBarker) Bark(event newshound.NewsEvent) error {
	msg := fmt.Sprintf("News Event Update! Now %d alerts are reporting on ", len(event.NewsAlerts))
	msg = twitterize(msg + strings.TrimPrefix(fmt.Sprintf("%#v", event.Tags), "[]string"))
	msg = msg + eventLink(event)
	_, err := s.api.PostTweet(msg, nil)
	return err
}

func twitterize(msg string) string {
	msg = strings.Replace(msg, "\n", " ", -1)
	msg = strings.Replace(msg, ".com", "", -1)
	if utf8.RuneCountInString(msg) >= 113 {
		msg = utf8string.NewString(msg).Slice(0, 113) + "... "
	} else if !strings.HasSuffix(msg, " ") {
		msg = msg + " "
	}

	return msg
}
