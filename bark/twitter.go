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

func NewTwitterAlertBarker(consumerKey, consumerSecret, token, secret string) *TwitterAlertBarker {
	anaconda.SetConsumerKey(consumerKey)
	anaconda.SetConsumerSecret(consumerSecret)
	return &TwitterAlertBarker{api: anaconda.NewTwitterApi(token, secret)}
}

type TwitterAlertBarker struct {
	api *anaconda.TwitterApi
}

func (s *TwitterAlertBarker) Bark(alert newshound.NewsAlertLite) error {
	msg := twitterize(fmt.Sprintf("%s - %s", strings.TrimSuffix(alert.Sender, ".com"), alert.TopSentence))
	msg = msg + alertLink(alert)
	_, err := s.api.PostTweet(msg, url.Values{})
	return err
}

func NewTwitterEventBarker(token, secret string) *TwitterEventBarker {
	return &TwitterEventBarker{api: anaconda.NewTwitterApi(token, secret)}
}

type TwitterEventBarker struct {
	api *anaconda.TwitterApi
}

func (s *TwitterEventBarker) Bark(event newshound.NewsEvent) error {
	msg := fmt.Sprintf("New News Event! %d alerts reporting on ", len(event.NewsAlerts))
	msg = twitterize(msg + strings.TrimPrefix(fmt.Sprintf("%#v", event.Tags), "[]string"))
	msg = msg + eventLink(event)
	_, err := s.api.PostTweet(msg, nil)
	return err
}

func twitterize(msg string) string {
	msg = strings.Replace(msg, "\n", " ", -1)
	msg = strings.Replace(msg, ".com", "", -1)
	if utf8.RuneCountInString(msg) >= 167 {
		msg = utf8string.NewString(msg).Slice(0, 160) + "... "
	} else if !strings.HasSuffix(msg, " ") {
		msg = msg + " "
	}

	return msg
}
